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

type expenseItemRepositoryGorm struct {
	db *gorm.DB
	crud.Repository[expenses.ExpenseItem]
}

func NewExpenseItemRepository(db *gorm.DB) *expenseItemRepositoryGorm {
	return &expenseItemRepositoryGorm{
		db,
		crud.NewRepository[expenses.ExpenseItem](db),
	}
}

func (ger *expenseItemRepositoryGorm) SyncParticipants(ctx context.Context, expenseItemID uuid.UUID, participants []expenses.ItemParticipant) error {
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
			DoUpdates: clause.AssignmentColumns([]string{"share"}),
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
