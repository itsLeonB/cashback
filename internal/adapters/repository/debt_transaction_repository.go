package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/appconstant"
	"github.com/itsLeonB/cashback/internal/core/logger"
	"github.com/itsLeonB/cashback/internal/core/otel"
	"github.com/itsLeonB/cashback/internal/domain/entity/debts"
	"github.com/itsLeonB/go-crud"
	"github.com/itsLeonB/ungerr"
	"gorm.io/gorm"
)

type debtTransactionRepositoryGorm struct {
	crud.Repository[debts.DebtTransaction]
}

func NewDebtTransactionRepository(db *gorm.DB) *debtTransactionRepositoryGorm {
	return &debtTransactionRepositoryGorm{
		crud.NewRepository[debts.DebtTransaction](db),
	}
}

func (dtr *debtTransactionRepositoryGorm) FindAllByMultipleProfileIDs(ctx context.Context, userProfileIDs, friendProfileIDs []uuid.UUID) ([]debts.DebtTransaction, error) {
	ctx, span := otel.Tracer.Start(ctx, "DebtTransactionRepository.FindAllByMultipleProfileIDs")
	defer span.End()

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
		Where("lender_profile_id IN ? AND borrower_profile_id IN ?", userProfileIDs, friendProfileIDs).
		Or("lender_profile_id IN ? AND borrower_profile_id IN ?", friendProfileIDs, userProfileIDs).
		Preload("TransferMethod").
		Order("created_at DESC").
		Find(&transactions).
		Error

	if err != nil {
		return nil, ungerr.Wrap(err, appconstant.ErrDataSelect)
	}

	return transactions, nil
}

func (dtr *debtTransactionRepositoryGorm) FindAllByProfileIDs(
	ctx context.Context,
	profileIDs []uuid.UUID,
	limit int,
	debtsOnly bool,
) ([]debts.DebtTransaction, error) {
	ctx, span := otel.Tracer.Start(ctx, "DebtTransactionRepository.FindAllByProfileIDs")
	defer span.End()

	if len(profileIDs) < 1 {
		return []debts.DebtTransaction{}, nil
	}

	var transactions []debts.DebtTransaction

	db, err := dtr.GetGormInstance(ctx)
	if err != nil {
		return nil, err
	}

	query := db.
		Where(db.Where("lender_profile_id IN ?", profileIDs).
			Or("borrower_profile_id IN ?", profileIDs)).
		Preload("TransferMethod").
		Scopes(crud.DefaultOrder())

	if limit > 0 {
		query = query.Limit(limit)
	}

	if debtsOnly {
		query = query.Where("group_expense_id IS NULL")
	}

	if err = query.Find(&transactions).Error; err != nil {
		return nil, ungerr.Wrap(err, appconstant.ErrDataSelect)
	}

	return transactions, nil
}

// RepointProfile repoints every debt transaction referencing anonProfileID (as lender or
// borrower) onto realProfileID. No unique constraint on these columns, so this is a plain update.
func (dtr *debtTransactionRepositoryGorm) RepointProfile(ctx context.Context, anonProfileID, realProfileID uuid.UUID) error {
	ctx, span := otel.Tracer.Start(ctx, "DebtTransactionRepository.RepointProfile")
	defer span.End()

	db, err := dtr.GetGormInstance(ctx)
	if err != nil {
		return err
	}

	// A debt transaction directly between the anon and real profile would become a
	// self-referencing (lender == borrower) row if repointed, which has no sensible
	// resolution (unlike a share amount, a directional debt can't just be summed).
	// Refuse the merge instead of corrupting the ledger. This check isn't locked against a
	// concurrent insert landing between it and the updates below, but the no_self_debt CHECK
	// constraint (see migrations) is the real backstop: it makes any such race fail the
	// UPDATE below with a constraint violation instead of silently corrupting the row.
	var conflicting int64
	if err := db.Model(&debts.DebtTransaction{}).
		Where("(lender_profile_id = ? AND borrower_profile_id = ?) OR (lender_profile_id = ? AND borrower_profile_id = ?)",
			anonProfileID, realProfileID, realProfileID, anonProfileID).
		Count(&conflicting).Error; err != nil {
		return ungerr.Wrap(err, appconstant.ErrDataSelect)
	}
	if conflicting > 0 {
		return ungerr.ConflictError("cannot merge: a debt transaction already exists directly between the anonymous and real profile")
	}

	if err := db.Model(&debts.DebtTransaction{}).
		Where("lender_profile_id = ?", anonProfileID).
		Update("lender_profile_id", realProfileID).Error; err != nil {
		return ungerr.Wrap(err, appconstant.ErrDataUpdate)
	}

	if err := db.Model(&debts.DebtTransaction{}).
		Where("borrower_profile_id = ?", anonProfileID).
		Update("borrower_profile_id", realProfileID).Error; err != nil {
		return ungerr.Wrap(err, appconstant.ErrDataUpdate)
	}

	return nil
}
