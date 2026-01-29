package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
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
	db, err := nr.GetGormInstance(ctx)
	if err != nil {
		return entity.Notification{}, err
	}

	err = db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "profile_id"}, {Name: "type"}, {Name: "entity_type"}, {Name: "entity_id"}},
		DoNothing: true,
	}).Create(&notification).Error
	if err != nil {
		return entity.Notification{}, ungerr.Wrap(err, appconstant.ErrDataInsert)
	}

	return notification, nil
}

func (nr *notificationRepositoryGorm) CreateMany(ctx context.Context, notifications []entity.Notification) ([]entity.Notification, error) {
	if len(notifications) == 0 {
		return []entity.Notification{}, nil
	}

	db, err := nr.GetGormInstance(ctx)
	if err != nil {
		return []entity.Notification{}, err
	}

	err = db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "profile_id"}, {Name: "type"}, {Name: "entity_type"}, {Name: "entity_id"}},
		DoNothing: true,
	}).Create(&notifications).Error
	if err != nil {
		return []entity.Notification{}, ungerr.Wrap(err, appconstant.ErrDataInsert)
	}

	return notifications, nil
}

func (nr *notificationRepositoryGorm) GetByProfileID(ctx context.Context, profileID uuid.UUID, unreadOnly bool) ([]entity.Notification, error) {
	db, err := nr.GetGormInstance(ctx)
	if err != nil {
		return nil, err
	}

	query := db.Where("profile_id = ?", profileID)

	if unreadOnly {
		query = query.Where("read_at IS NULL")
	}

	var notifications []entity.Notification
	if err = query.Order("created_at DESC").Find(&notifications).Error; err != nil {
		return nil, ungerr.Wrap(err, appconstant.ErrDataSelect)
	}

	return notifications, nil
}

func (nr *notificationRepositoryGorm) MarkAsRead(ctx context.Context, profileID, notificationID uuid.UUID) error {
	db, err := nr.GetGormInstance(ctx)
	if err != nil {
		return err
	}

	if err = db.
		Model(&entity.Notification{}).
		Where("id = ? AND profile_id = ?", notificationID, profileID).
		Update("read_at", time.Now()).Error; err != nil {
		return ungerr.Wrap(err, appconstant.ErrDataUpdate)
	}

	return nil
}

func (nr *notificationRepositoryGorm) MarkAllAsRead(ctx context.Context, profileID uuid.UUID) error {
	db, err := nr.GetGormInstance(ctx)
	if err != nil {
		return err
	}

	if err = db.
		Model(&entity.Notification{}).
		Where("profile_id = ? AND read_at IS NULL", profileID).
		Update("read_at", time.Now()).Error; err != nil {
		return ungerr.Wrap(err, appconstant.ErrDataUpdate)
	}

	return nil
}
