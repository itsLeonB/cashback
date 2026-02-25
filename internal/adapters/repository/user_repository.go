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

type userRepository struct {
	crud.Repository[users.User]
}

func NewUserRepository(db *gorm.DB) *userRepository {
	return &userRepository{crud.NewRepository[users.User](db)}
}

func (ur *userRepository) FindByIDs(ctx context.Context, ids []uuid.UUID) ([]users.User, error) {
	db, err := ur.GetGormInstance(ctx)
	if err != nil {
		return nil, err
	}

	var result []users.User
	if err = db.Preload("Profile").Where("id IN ?", ids).Find(&result).Error; err != nil {
		return nil, ungerr.Wrap(err, appconstant.ErrDataSelect)
	}

	return result, nil
}
