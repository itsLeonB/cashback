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

type friendshipRepositoryGorm struct {
	crud.Repository[users.Friendship]
	db *gorm.DB
}

func NewFriendshipRepository(db *gorm.DB) *friendshipRepositoryGorm {
	return &friendshipRepositoryGorm{
		crud.NewRepository[users.Friendship](db),
		db,
	}
}

func (fr *friendshipRepositoryGorm) Insert(ctx context.Context, friendship users.Friendship) (users.Friendship, error) {
	db, err := fr.GetGormInstance(ctx)
	if err != nil {
		return users.Friendship{}, err
	}

	if err = db.Create(&friendship).Error; err != nil {
		return users.Friendship{}, ungerr.Wrap(err, appconstant.ErrDataInsert)
	}

	return friendship, nil
}

func (fr *friendshipRepositoryGorm) FindFirstBySpec(ctx context.Context, spec users.FriendshipSpecification) (users.Friendship, error) {
	var friendship users.Friendship

	db, err := fr.GetGormInstance(ctx)
	if err != nil {
		return users.Friendship{}, err
	}

	query := db.
		Scopes(
			crud.WhereBySpec(spec.Model),
			crud.PreloadRelations(spec.PreloadRelations),
		).
		Joins("JOIN user_profiles AS up1 ON up1.id = friendships.profile_id1").
		Joins("JOIN user_profiles AS up2 ON up2.id = friendships.profile_id2")

	if spec.Name != "" {
		query = query.Where(
			db.Where("up1.name = ? AND friendships.profile_id1 <> ?", spec.Name, spec.Model.ProfileID1).
				Or("up2.name = ? AND friendships.profile_id2 <> ?", spec.Name, spec.Model.ProfileID1),
		)
	}

	err = query.Take(&friendship).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return users.Friendship{}, nil
		}
		return users.Friendship{}, ungerr.Wrap(err, appconstant.ErrDataSelect)
	}

	return friendship, nil
}

func (fr *friendshipRepositoryGorm) FindAllBySpec(ctx context.Context, spec users.FriendshipSpecification) ([]users.Friendship, error) {
	var friendships []users.Friendship

	db, err := fr.GetGormInstance(ctx)
	if err != nil {
		return nil, err
	}

	err = db.
		Where(users.Friendship{ProfileID1: spec.Model.ProfileID1}).
		Or(users.Friendship{ProfileID2: spec.Model.ProfileID1}).
		Scopes(
			crud.PreloadRelations(spec.PreloadRelations),
			crud.DefaultOrder(),
		).
		Find(&friendships).
		Error

	if err != nil {
		return nil, ungerr.Wrap(err, appconstant.ErrDataSelect)
	}

	return friendships, nil
}

func (fr *friendshipRepositoryGorm) FindByProfileIDs(ctx context.Context, profileID1, profileID2 uuid.UUID) (users.Friendship, error) {
	db, err := fr.GetGormInstance(ctx)
	if err != nil {
		return users.Friendship{}, err
	}

	var friendship users.Friendship
	err = db.Where("(profile_id1 = ? AND profile_id2 = ?) OR (profile_id1 = ? AND profile_id2 = ?)", profileID1, profileID2, profileID2, profileID1).
		First(&friendship).
		Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return users.Friendship{}, nil
		}
		return users.Friendship{}, ungerr.Wrap(err, appconstant.ErrDataSelect)
	}

	return friendship, nil
}
