package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/domain/entity/debts"
	"github.com/itsLeonB/go-crud"
	"github.com/itsLeonB/ungerr"
	"gorm.io/gorm"
)

type transferMethodRepositoryGorm struct {
	crud.Repository[debts.TransferMethod]
	db *gorm.DB
}

func NewTransferMethodRepository(db *gorm.DB) *transferMethodRepositoryGorm {
	return &transferMethodRepositoryGorm{
		crud.NewRepository[debts.TransferMethod](db),
		db,
	}
}

func (tmr *transferMethodRepositoryGorm) GetAllByParentFilter(ctx context.Context, filter debts.ParentFilter, profileID uuid.UUID) ([]debts.TransferMethod, error) {
	var methods []debts.TransferMethod

	query := tmr.db
	switch filter {
	case debts.ParentOnly:
		query = query.Where("parent_id IS NULL")
	case debts.ChildOnly:
		query = query.Where("parent_id IS NOT NULL")
	case debts.ForTransaction:
		query = query.Table("transfer_methods tm").
			Joins("JOIN profile_transfer_methods ptm ON tm.id = ptm.transfer_method_id").
			Where("tm.parent_id IS NULL OR ptm.profile_id = ?", profileID)
	default:
		// no filter
	}

	if err := query.Find(&methods).Error; err != nil {
		return nil, ungerr.Wrapf(err, "error querying transfer methods by filter: %s", filter)
	}

	return methods, nil
}
