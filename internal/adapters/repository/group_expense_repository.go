package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/appconstant"
	"github.com/itsLeonB/cashback/internal/domain/entity/expenses"
	"github.com/itsLeonB/go-crud"
	"github.com/itsLeonB/ungerr"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type groupExpenseRepositoryGorm struct {
	crud.Repository[expenses.GroupExpense]
	db *gorm.DB
}

func NewGroupExpenseRepository(db *gorm.DB) *groupExpenseRepositoryGorm {
	return &groupExpenseRepositoryGorm{
		crud.NewRepository[expenses.GroupExpense](db),
		db,
	}
}

func (ger *groupExpenseRepositoryGorm) SyncParticipants(ctx context.Context, groupExpenseID uuid.UUID, participants []expenses.ExpenseParticipant) error {
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
			DoUpdates: clause.AssignmentColumns([]string{"share_amount"}),
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

func (ger *groupExpenseRepositoryGorm) DeleteItemParticipants(ctx context.Context, expenseID uuid.UUID, newParticipantProfileIDs []uuid.UUID) error {
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
