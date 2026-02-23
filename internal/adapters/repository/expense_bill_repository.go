package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/appconstant"
	"github.com/itsLeonB/cashback/internal/domain/entity/expenses"
	"github.com/itsLeonB/go-crud"
	"github.com/itsLeonB/ungerr"
	"gorm.io/gorm"
)

type expenseBillRepo struct {
	crud.Repository[expenses.ExpenseBill]
	db *gorm.DB
}

func NewExpenseBillRepository(db *gorm.DB) *expenseBillRepo {
	return &expenseBillRepo{
		crud.NewRepository[expenses.ExpenseBill](db),
		db,
	}
}

func (sr *expenseBillRepo) CountUploadedByDateRange(ctx context.Context, profileID uuid.UUID, start, end time.Time) (int, error) {
	db, err := sr.GetGormInstance(ctx)
	if err != nil {
		return 0, err
	}

	var count int64
	err = db.Model(&expenses.ExpenseBill{}).
		Joins("JOIN group_expenses ON group_expenses.id = expense_bills.group_expense_id").
		Where("group_expenses.creator_profile_id = ?", profileID).
		Where("expense_bills.created_at >= ? AND expense_bills.created_at <= ?", start, end).
		Count(&count).Error

	if err != nil {
		return 0, ungerr.Wrap(err, appconstant.ErrDataSelect)
	}

	return int(count), nil
}
