package monetization

import (
	"context"

	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/appconstant"
	"github.com/itsLeonB/cashback/internal/domain/entity/monetization"
	"github.com/itsLeonB/go-crud"
	"github.com/itsLeonB/ungerr"
	"gorm.io/gorm"
)

type planVersionRepository struct {
	crud.Repository[monetization.PlanVersion]
	db *gorm.DB
}

func NewPlanVersionRepository(db *gorm.DB) *planVersionRepository {
	return &planVersionRepository{
		crud.NewRepository[monetization.PlanVersion](db),
		db,
	}
}

func (pvr *planVersionRepository) SetAsDefault(ctx context.Context, id uuid.UUID) error {
	db, err := pvr.GetGormInstance(ctx)
	if err != nil {
		return err
	}

	err = db.Model(&monetization.PlanVersion{}).
		Where("id <> ?", id).
		Update("is_default", false).
		Error

	if err != nil {
		return ungerr.Wrap(err, appconstant.ErrDataUpdate)
	}

	return nil
}
