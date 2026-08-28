package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/appconstant"
	"github.com/itsLeonB/cashback/internal/core/otel"
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

// Delete removes a profile row. Before doing so it purges any leftover related_profiles row
// referencing it as the anon side: that table is retired (the old pre-physical-merge anon
// linking model), but its anon_profile_id FK has no ON DELETE behavior, so a legacy row left
// over from before the merge refactor would otherwise block this delete.
func (pr *profileRepositoryGorm) Delete(ctx context.Context, model users.UserProfile) error {
	ctx, span := otel.Tracer.Start(ctx, "ProfileRepository.Delete")
	defer span.End()

	if model.IsZero() {
		return ungerr.Unknown("model cannot be zero value")
	}

	db, err := pr.GetGormInstance(ctx)
	if err != nil {
		return err
	}

	if err := db.Exec("DELETE FROM related_profiles WHERE anon_profile_id = ?", model.ID).Error; err != nil {
		return ungerr.Wrap(err, "error deleting legacy related_profiles row")
	}

	if err := db.Unscoped().Delete(&model).Error; err != nil {
		return ungerr.Wrap(err, "error deleting data")
	}

	return nil
}

func (pr *profileRepositoryGorm) FindByIDs(ctx context.Context, ids []uuid.UUID) ([]users.UserProfile, error) {
	ctx, span := otel.Tracer.Start(ctx, "ProfileRepository.FindByIDs")
	defer span.End()

	var profiles []users.UserProfile

	db, err := pr.GetGormInstance(ctx)
	if err != nil {
		return nil, err
	}

	if err = db.
		Where("id IN ?", ids).
		Find(&profiles).
		Error; err != nil {
		return nil, ungerr.Wrap(err, appconstant.ErrDataSelect)
	}

	return profiles, nil
}

func (pr *profileRepositoryGorm) FindRealProfiles(ctx context.Context) ([]users.UserProfile, error) {
	ctx, span := otel.Tracer.Start(ctx, "ProfileRepository.FindRealProfiles")
	defer span.End()

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
	ctx, span := otel.Tracer.Start(ctx, "ProfileRepository.SearchByName")
	defer span.End()

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
