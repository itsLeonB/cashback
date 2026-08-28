package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/appconstant"
	"github.com/itsLeonB/cashback/internal/core/otel"
	"github.com/itsLeonB/cashback/internal/domain/entity/expenses"
	"github.com/itsLeonB/go-crud"
	"github.com/itsLeonB/ungerr"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type groupExpenseRepositoryGorm struct {
	crud.Repository[expenses.GroupExpense]
}

func NewGroupExpenseRepository(db *gorm.DB) *groupExpenseRepositoryGorm {
	return &groupExpenseRepositoryGorm{
		crud.NewRepository[expenses.GroupExpense](db),
	}
}

func (ger *groupExpenseRepositoryGorm) Update(ctx context.Context, model expenses.GroupExpense) (expenses.GroupExpense, error) {
	ctx, span := otel.Tracer.Start(ctx, "GroupExpenseRepository.Update")
	defer span.End()

	if model.IsZero() {
		return expenses.GroupExpense{}, ungerr.Unknown("model cannot be zero value")
	}

	db, err := ger.GetGormInstance(ctx)
	if err != nil {
		return expenses.GroupExpense{}, err
	}

	// ponytail: only Bill is omitted from Save's association cascade. GORM's Save()
	// upserts populated associations by default. Since callers preload Bill via
	// GetUnconfirmedForUpdate, an un-omitted Save() here attempts to write the
	// expense_bills row too - which is the row cashback-worker holds FOR UPDATE locked
	// for the full duration of its LLM call, causing this Update to block for seconds.
	// Items/OtherFees/Participants must keep cascading: UpdateDraft relies on this Save
	// to persist freshly-parsed items/fees, with no other repository call for them.
	if err := db.Omit("Bill").Save(&model).Error; err != nil {
		return expenses.GroupExpense{}, ungerr.Wrap(err, appconstant.ErrDataUpdate)
	}

	return model, nil
}

func (ger *groupExpenseRepositoryGorm) SyncParticipants(ctx context.Context, groupExpenseID uuid.UUID, participants []expenses.ExpenseParticipant) error {
	ctx, span := otel.Tracer.Start(ctx, "GroupExpenseRepository.SyncParticipants")
	defer span.End()

	db, err := ger.GetGormInstance(ctx)
	if err != nil {
		return err
	}

	profileIDs := make([]uuid.UUID, len(participants))
	for i, p := range participants {
		participants[i].GroupExpenseID = groupExpenseID
		profileIDs[i] = p.ParticipantProfileID
	}

	if len(participants) > 0 {
		// Upsert: insert new or update existing
		if err := db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "group_expense_id"}, {Name: "participant_profile_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"share_amount", "proxy_profile_id"}),
		}).Create(&participants).Error; err != nil {
			return ungerr.Wrap(err, appconstant.ErrDataUpdate)
		}
	}

	// Delete participants not in the new list
	if len(profileIDs) > 0 {
		if err := db.Where("group_expense_id = ? AND participant_profile_id NOT IN ?", groupExpenseID, profileIDs).
			Delete(&expenses.ExpenseParticipant{}).Error; err != nil {
			return ungerr.Wrap(err, "error deleting removed participants")
		}
	} else {
		// If no participants provided, delete all
		if err := db.Where("group_expense_id = ?", groupExpenseID).
			Delete(&expenses.ExpenseParticipant{}).Error; err != nil {
			return ungerr.Wrap(err, "error deleting all participants")
		}
	}

	return nil
}

// RepointProfile repoints group_expenses.payer/creator and group_expense_participants
// referencing anonProfileID onto realProfileID. If realProfileID is already a participant in
// the same group expense as anonProfileID, their share amounts are summed and the anonymous
// row is dropped instead of violating unique_expense_profile.
func (ger *groupExpenseRepositoryGorm) RepointProfile(ctx context.Context, anonProfileID, realProfileID uuid.UUID) error {
	ctx, span := otel.Tracer.Start(ctx, "GroupExpenseRepository.RepointProfile")
	defer span.End()

	db, err := ger.GetGormInstance(ctx)
	if err != nil {
		return err
	}

	if err := db.Model(&expenses.GroupExpense{}).
		Where("payer_profile_id = ?", anonProfileID).
		Update("payer_profile_id", realProfileID).Error; err != nil {
		return ungerr.Wrap(err, appconstant.ErrDataUpdate)
	}
	if err := db.Model(&expenses.GroupExpense{}).
		Where("creator_profile_id = ?", anonProfileID).
		Update("creator_profile_id", realProfileID).Error; err != nil {
		return ungerr.Wrap(err, appconstant.ErrDataUpdate)
	}

	var anonRows []expenses.ExpenseParticipant
	if err := db.Where("participant_profile_id = ?", anonProfileID).Find(&anonRows).Error; err != nil {
		return ungerr.Wrap(err, appconstant.ErrDataSelect)
	}

	if len(anonRows) > 0 {
		groupExpenseIDs := make([]uuid.UUID, len(anonRows))
		anonProxyByExpense := make(map[uuid.UUID]uuid.NullUUID, len(anonRows))
		for i, row := range anonRows {
			groupExpenseIDs[i] = row.GroupExpenseID
			anonProxyByExpense[row.GroupExpenseID] = row.ProxyProfileID
		}

		// Reject up front if any colliding real row already names a different proxy —
		// summing share amounts is safe, but silently picking one side's proxy is not
		// (ProxyProfileID determines who owes whom, see GroupExpenseToDebtTransactions).
		var existingRealRows []expenses.ExpenseParticipant
		if err := db.Where("participant_profile_id = ? AND group_expense_id IN ?", realProfileID, groupExpenseIDs).
			Find(&existingRealRows).Error; err != nil {
			return ungerr.Wrap(err, appconstant.ErrDataSelect)
		}
		for _, realRow := range existingRealRows {
			anonProxy := anonProxyByExpense[realRow.GroupExpenseID]
			if anonProxy.Valid && realRow.ProxyProfileID.Valid && realRow.ProxyProfileID.UUID != anonProxy.UUID {
				return ungerr.ConflictError("cannot merge: colliding group expense participants have different proxy profiles")
			}
		}

		merged := make([]expenses.ExpenseParticipant, len(anonRows))
		for i, row := range anonRows {
			merged[i] = expenses.ExpenseParticipant{
				GroupExpenseID:       row.GroupExpenseID,
				ParticipantProfileID: realProfileID,
				ProxyProfileID:       row.ProxyProfileID,
				ShareAmount:          row.ShareAmount,
			}
		}

		if err := db.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "group_expense_id"}, {Name: "participant_profile_id"}},
			DoUpdates: clause.Assignments(map[string]any{
				"share_amount":     gorm.Expr("group_expense_participants.share_amount + excluded.share_amount"),
				"proxy_profile_id": gorm.Expr("COALESCE(group_expense_participants.proxy_profile_id, excluded.proxy_profile_id)"),
			}),
		}).Create(&merged).Error; err != nil {
			return ungerr.Wrap(err, appconstant.ErrDataUpdate)
		}

		if err := db.Where("participant_profile_id = ?", anonProfileID).Delete(&expenses.ExpenseParticipant{}).Error; err != nil {
			return ungerr.Wrap(err, "error deleting merged expense participants")
		}
	}

	// proxy_profile_id carries no uniqueness constraint, so a plain update is safe
	if err := db.Model(&expenses.ExpenseParticipant{}).
		Where("proxy_profile_id = ?", anonProfileID).
		Update("proxy_profile_id", realProfileID).Error; err != nil {
		return ungerr.Wrap(err, appconstant.ErrDataUpdate)
	}

	return nil
}

func (ger *groupExpenseRepositoryGorm) DeleteItemParticipants(ctx context.Context, expenseID uuid.UUID, newParticipantProfileIDs []uuid.UUID) error {
	ctx, span := otel.Tracer.Start(ctx, "GroupExpenseRepository.DeleteItemParticipants")
	defer span.End()

	db, err := ger.GetGormInstance(ctx)
	if err != nil {
		return err
	}

	// GORM doesn't support DELETE with JOIN directly, so we use a subquery
	subQuery := db.Table("group_expense_items").
		Select("id").
		Where("group_expense_id = ?", expenseID)

	query := db.Where("expense_item_id IN (?)", subQuery)

	if len(newParticipantProfileIDs) > 0 {
		query = query.Where("profile_id NOT IN ?", newParticipantProfileIDs)
	}

	if err := query.Delete(&expenses.ItemParticipant{}).Error; err != nil {
		return ungerr.Wrap(err, "error deleting item participants")
	}

	return nil
}

func (ger *groupExpenseRepositoryGorm) FindAllByOwnership(ctx context.Context, profileID uuid.UUID, ownership expenses.ExpenseOwnership, status expenses.ExpenseStatus, limit int) ([]expenses.GroupExpense, error) {
	ctx, span := otel.Tracer.Start(ctx, "GroupExpenseRepository.FindAllByOwnership")
	defer span.End()

	db, err := ger.GetGormInstance(ctx)
	if err != nil {
		return nil, err
	}

	var groupExpenses []expenses.GroupExpense
	query := db.Preload("Items").Preload("OtherFees").Preload("Participants").Preload("Payer").Preload("Creator")

	switch ownership {
	case expenses.OwnedExpense:
		query = query.Where("creator_profile_id = ?", profileID)
	case expenses.ParticipatingExpense:
		query = query.Joins("JOIN group_expense_participants ON group_expenses.id = group_expense_participants.group_expense_id").
			Where("group_expense_participants.participant_profile_id = ? AND creator_profile_id != ?", profileID, profileID)
	default:
		return nil, ungerr.BadRequestError("invalid ownership filter")
	}

	// Handle status filtering
	if status != "" {
		if status == "UNCONFIRMED" {
			query = query.Where("status IN ?", []expenses.ExpenseStatus{expenses.DraftExpense, expenses.ReadyExpense})
		} else {
			query = query.Where("status = ?", status)
		}
	}

	query = query.Order("updated_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err = query.Find(&groupExpenses).Error
	if err != nil {
		return nil, ungerr.Wrap(err, appconstant.ErrDataSelect)
	}

	return groupExpenses, nil
}

func (ger *groupExpenseRepositoryGorm) FindRecentByProfileID(ctx context.Context, profileID uuid.UUID, limit int) ([]expenses.GroupExpense, error) {
	ctx, span := otel.Tracer.Start(ctx, "GroupExpenseRepository.FindRecentByProfileID")
	defer span.End()

	db, err := ger.GetGormInstance(ctx)
	if err != nil {
		return nil, err
	}

	var groupExpenses []expenses.GroupExpense

	query := db.Preload("Creator").
		Where("creator_profile_id = ? OR id IN (SELECT group_expense_id FROM group_expense_participants WHERE participant_profile_id = ?)", profileID, profileID).
		Order("updated_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err = query.Find(&groupExpenses).Error
	if err != nil {
		return nil, ungerr.Wrap(err, appconstant.ErrDataSelect)
	}

	return groupExpenses, nil
}
