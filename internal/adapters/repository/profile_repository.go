package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/appconstant"
	"github.com/itsLeonB/cashback/internal/domain/entity/users"
	"github.com/itsLeonB/go-crud"
	"github.com/itsLeonB/ungerr"
	"gorm.io/gorm"
)

type profileRepositoryGorm struct {
	crud.Repository[users.UserProfile]
}

func NewProfileRepository(db *gorm.DB) *profileRepositoryGorm {
	return &profileRepositoryGorm{
		crud.NewRepository[users.UserProfile](db),
	}
}

func (pr *profileRepositoryGorm) FindByIDs(ctx context.Context, ids []uuid.UUID) ([]users.UserProfile, error) {
	var profiles []users.UserProfile

	db, err := pr.GetGormInstance(ctx)
	if err != nil {
		return nil, err
	}

	if err = db.
		Scopes(crud.PreloadRelations([]string{
			"RelatedRealProfile",
			"RelatedAnonProfiles",
		})).
		Where("id IN ?", ids).
		Find(&profiles).
		Error; err != nil {
		return nil, ungerr.Wrap(err, appconstant.ErrDataSelect)
	}

	return profiles, nil
}

func (pr *profileRepositoryGorm) FindRealProfiles(ctx context.Context) ([]users.UserProfile, error) {
	db, err := pr.GetGormInstance(ctx)
	if err != nil {
		return nil, err
	}

	var profiles []users.UserProfile
	if err := db.
		Where("user_id IS NOT NULL").
		Find(&profiles).Error; err != nil {
		return nil, ungerr.Wrap(err, appconstant.ErrDataSelect)
	}

	return profiles, nil
}

func (pr *profileRepositoryGorm) SearchByName(ctx context.Context, query string, limit int) ([]users.UserProfile, error) {
	db, err := pr.GetGormInstance(ctx)
	if err != nil {
		return nil, err
	}

	var results []users.UserProfile
	if err := db.
		Table("user_profiles").
		Where("name % ?", query).
		Where("user_id IS NOT NULL").
		Order(gorm.Expr("similarity(name, ?) DESC", query)).
		Limit(limit).
		Find(&results).Error; err != nil {
		return nil, ungerr.Wrap(err, appconstant.ErrDataSelect)
	}
	return results, nil
}
