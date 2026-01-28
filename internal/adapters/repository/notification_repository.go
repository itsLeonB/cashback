package repository

import (
	"context"

	"github.com/itsLeonB/cashback/internal/appconstant"
	"github.com/itsLeonB/cashback/internal/domain/entity"
	"github.com/itsLeonB/go-crud"
	"github.com/itsLeonB/ungerr"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type notificationRepositoryGorm struct {
	crud.Repository[entity.Notification]
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) *notificationRepositoryGorm {
	return &notificationRepositoryGorm{
		crud.NewRepository[entity.Notification](db),
		db,
	}
}

func (nr *notificationRepositoryGorm) New(ctx context.Context, notification entity.Notification) (entity.Notification, error) {
	err := nr.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "profile_id"}, {Name: "type"}, {Name: "entity_type"}, {Name: "entity_id"}},
		DoNothing: true,
	}).Create(&notification).Error
	if err != nil {
		return entity.Notification{}, ungerr.Wrap(err, appconstant.ErrDataInsert)
	}

	return notification, nil
}
