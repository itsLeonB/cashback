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

type expenseItemRepositoryGorm struct {
	crud.Repository[expenses.ExpenseItem]
}

func NewExpenseItemRepository(db *gorm.DB) *expenseItemRepositoryGorm {
	return &expenseItemRepositoryGorm{
		crud.NewRepository[expenses.ExpenseItem](db),
	}
}

func (ger *expenseItemRepositoryGorm) SyncParticipants(ctx context.Context, expenseItemID uuid.UUID, participants []expenses.ItemParticipant) error {
	ctx, span := otel.Tracer.Start(ctx, "ExpenseItemRepository.SyncParticipants")
	defer span.End()

	db, err := ger.GetGormInstance(ctx)
	if err != nil {
		return err
	}

	profileIDs := make([]uuid.UUID, len(participants))
	for i, p := range participants {
		participants[i].ExpenseItemID = expenseItemID
		profileIDs[i] = p.ProfileID
	}

	if len(participants) > 0 {
		// For PostgreSQL
		if err := db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "expense_item_id"}, {Name: "profile_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"weight", "allocated_amount"}),
		}).Create(&participants).Error; err != nil {
			return ungerr.Wrap(err, appconstant.ErrDataUpdate)
		}
	}

	query := db.Where("expense_item_id = ?", expenseItemID)
	if len(profileIDs) > 0 {
		query = query.Where("profile_id NOT IN ?", profileIDs)
	}
	if err := query.Delete(&expenses.ItemParticipant{}).Error; err != nil {
		return err
	}

	return nil
}

// RepointParticipants merges group_expense_item_participants rows referencing anonProfileID
// onto realProfileID. If realProfileID already has a row for the same expense item, weight and
// allocated_amount are summed and the anonymous row is dropped instead of violating
// unique_expense_item_profile.
func (ger *expenseItemRepositoryGorm) RepointParticipants(ctx context.Context, anonProfileID, realProfileID uuid.UUID) error {
	ctx, span := otel.Tracer.Start(ctx, "ExpenseItemRepository.RepointParticipants")
	defer span.End()

	db, err := ger.GetGormInstance(ctx)
	if err != nil {
		return err
	}

	var anonRows []expenses.ItemParticipant
	if err := db.Where("profile_id = ?", anonProfileID).Find(&anonRows).Error; err != nil {
		return ungerr.Wrap(err, appconstant.ErrDataSelect)
	}
	if len(anonRows) == 0 {
		return nil
	}

	merged := make([]expenses.ItemParticipant, len(anonRows))
	for i, row := range anonRows {
		merged[i] = expenses.ItemParticipant{
			ExpenseItemID:   row.ExpenseItemID,
			ProfileID:       realProfileID,
			Weight:          row.Weight,
			AllocatedAmount: row.AllocatedAmount,
		}
	}

	if err := db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "expense_item_id"}, {Name: "profile_id"}},
		DoUpdates: clause.Assignments(map[string]any{
			"weight":           gorm.Expr("group_expense_item_participants.weight + excluded.weight"),
			"allocated_amount": gorm.Expr("group_expense_item_participants.allocated_amount + excluded.allocated_amount"),
		}),
	}).Create(&merged).Error; err != nil {
		return ungerr.Wrap(err, appconstant.ErrDataUpdate)
	}

	if err := db.Where("profile_id = ?", anonProfileID).Delete(&expenses.ItemParticipant{}).Error; err != nil {
		return ungerr.Wrap(err, "error deleting merged item participants")
	}

	return nil
}
