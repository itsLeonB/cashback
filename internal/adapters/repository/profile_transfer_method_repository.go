package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/appconstant"
	"github.com/itsLeonB/cashback/internal/core/otel"
	"github.com/itsLeonB/cashback/internal/domain/entity/debts"
	"github.com/itsLeonB/go-crud"
	"github.com/itsLeonB/ungerr"
	"gorm.io/gorm"
)

type profileTransferMethodRepositoryGorm struct {
	crud.Repository[debts.ProfileTransferMethod]
}

func NewProfileTransferMethodRepository(db *gorm.DB) *profileTransferMethodRepositoryGorm {
	return &profileTransferMethodRepositoryGorm{
		crud.NewRepository[debts.ProfileTransferMethod](db),
	}
}

// RepointProfile repoints every transfer method referencing anonProfileID onto realProfileID.
// No unique constraint on this column, so this is a plain update.
func (pr *profileTransferMethodRepositoryGorm) RepointProfile(ctx context.Context, anonProfileID, realProfileID uuid.UUID) error {
	ctx, span := otel.Tracer.Start(ctx, "ProfileTransferMethodRepository.RepointProfile")
	defer span.End()

	db, err := pr.GetGormInstance(ctx)
	if err != nil {
		return err
	}

	if err := db.Model(&debts.ProfileTransferMethod{}).
		Where("profile_id = ?", anonProfileID).
		Update("profile_id", realProfileID).Error; err != nil {
		return ungerr.Wrap(err, appconstant.ErrDataUpdate)
	}

	return nil
}
