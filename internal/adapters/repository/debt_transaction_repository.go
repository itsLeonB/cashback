package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/appconstant"
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

func (dtr *debtTransactionRepositoryGorm) FindAllByProfileIDs(ctx context.Context, userProfileID, friendProfileID uuid.UUID) ([]debts.DebtTransaction, error) {
	var transactions []debts.DebtTransaction

	db, err := dtr.GetGormInstance(ctx)
	if err != nil {
		return nil, err
	}

	err = db.
		Scopes(crud.ForUpdate(true)).
		Where("lender_profile_id = ? AND borrower_profile_id = ?", userProfileID, friendProfileID).
		Or("lender_profile_id = ? AND borrower_profile_id = ?", friendProfileID, userProfileID).
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
