package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/appconstant"
	"github.com/itsLeonB/cashback/internal/core/logger"
	"github.com/itsLeonB/cashback/internal/domain/entity/debts"
	"github.com/itsLeonB/go-crud"
	"github.com/itsLeonB/ungerr"
	"gorm.io/gorm"
)

type debtTransactionRepositoryGorm struct {
	crud.Repository[debts.DebtTransaction]
	db *gorm.DB
}

func NewDebtTransactionRepository(db *gorm.DB) *debtTransactionRepositoryGorm {
	return &debtTransactionRepositoryGorm{
		crud.NewRepository[debts.DebtTransaction](db),
		db,
	}
}

func (dtr *debtTransactionRepositoryGorm) FindAllByMultipleProfileIDs(ctx context.Context, userProfileIDs, friendProfileIDs []uuid.UUID) ([]debts.DebtTransaction, error) {
	if len(userProfileIDs) == 0 || len(friendProfileIDs) == 0 {
		logger.Warn("DebtTransactionRepository.FindAllByMultipleProfileIDs input is empty slice")
		return []debts.DebtTransaction{}, nil
	}
	var transactions []debts.DebtTransaction

	db, err := dtr.GetGormInstance(ctx)
	if err != nil {
		return nil, err
	}

	err = db.
		Scopes(crud.ForUpdate(true)).
		Where("lender_profile_id IN ? AND borrower_profile_id IN ?", userProfileIDs, friendProfileIDs).
		Or("lender_profile_id IN ? AND borrower_profile_id IN ?", friendProfileIDs, userProfileIDs).
		Find(&transactions).
		Error

	if err != nil {
		return nil, ungerr.Wrap(err, appconstant.ErrDataSelect)
	}

	return transactions, nil
}

func (dtr *debtTransactionRepositoryGorm) FindAllByUserProfileID(ctx context.Context, userProfileID uuid.UUID) ([]debts.DebtTransaction, error) {
	var transactions []debts.DebtTransaction

	db, err := dtr.GetGormInstance(ctx)
	if err != nil {
		return nil, err
	}

	err = db.
		Where("lender_profile_id = ?", userProfileID).
		Or("borrower_profile_id = ?", userProfileID).
		Preload("TransferMethod").
		Scopes(crud.DefaultOrder()).
		Find(&transactions).
		Error

	if err != nil {
		return nil, ungerr.Wrap(err, appconstant.ErrDataSelect)
	}

	return transactions, nil
}
