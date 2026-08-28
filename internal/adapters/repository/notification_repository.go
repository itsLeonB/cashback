package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/appconstant"
	"github.com/itsLeonB/cashback/internal/core/otel"
	"github.com/itsLeonB/cashback/internal/domain/entity"
	"github.com/itsLeonB/go-crud"
	"github.com/itsLeonB/ungerr"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type notificationRepositoryGorm struct {
	crud.Repository[entity.Notification]
}

func NewNotificationRepository(db *gorm.DB) *notificationRepositoryGorm {
	return &notificationRepositoryGorm{
		crud.NewRepository[entity.Notification](db),
	}
}

func (nr *notificationRepositoryGorm) New(ctx context.Context, notification entity.Notification) (entity.Notification, error) {
	ctx, span := otel.Tracer.Start(ctx, "NotificationRepository.New")
	defer span.End()

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
	ctx, span := otel.Tracer.Start(ctx, "NotificationRepository.CreateMany")
	defer span.End()

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
	ctx, span := otel.Tracer.Start(ctx, "NotificationRepository.GetByProfileID")
	defer span.End()

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
	ctx, span := otel.Tracer.Start(ctx, "NotificationRepository.MarkAsRead")
	defer span.End()

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
	ctx, span := otel.Tracer.Start(ctx, "NotificationRepository.MarkAllAsRead")
	defer span.End()

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

// RepointProfile repoints notifications referencing anonProfileID onto realProfileID. If
// realProfileID already has a matching (type, entity_type, entity_id) notification, the
// anonymous row is dropped instead of violating notifications_unique_entity_idx.
func (nr *notificationRepositoryGorm) RepointProfile(ctx context.Context, anonProfileID, realProfileID uuid.UUID) error {
	ctx, span := otel.Tracer.Start(ctx, "NotificationRepository.RepointProfile")
	defer span.End()

	db, err := nr.GetGormInstance(ctx)
	if err != nil {
		return err
	}

	var anonRows []entity.Notification
	if err := db.Where("profile_id = ?", anonProfileID).Find(&anonRows).Error; err != nil {
		return ungerr.Wrap(err, appconstant.ErrDataSelect)
	}
	if len(anonRows) == 0 {
		return nil
	}

	moved := make([]entity.Notification, len(anonRows))
	anonRowIDs := make([]uuid.UUID, len(anonRows))
	for i, n := range anonRows {
		moved[i] = entity.Notification{
			ProfileID:  realProfileID,
			Type:       n.Type,
			EntityType: n.EntityType,
			EntityID:   n.EntityID,
			Metadata:   n.Metadata,
			ReadAt:     n.ReadAt,
			PushedAt:   n.PushedAt,
		}
		anonRowIDs[i] = n.ID
	}

	// If the real profile already has a matching (type, entity_type, entity_id)
	// notification, keep it and drop the anon duplicate below.
	if err := db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "profile_id"}, {Name: "type"}, {Name: "entity_type"}, {Name: "entity_id"}},
		DoNothing: true,
	}).Create(&moved).Error; err != nil {
		return ungerr.Wrap(err, appconstant.ErrDataInsert)
	}

	// Restricted to the exact rows read above so a row inserted concurrently after the
	// SELECT - and never merged into realProfileID - survives instead of being silently
	// dropped.
	if err := db.Where("id IN ?", anonRowIDs).Delete(&entity.Notification{}).Error; err != nil {
		return ungerr.Wrap(err, "error deleting merged notifications")
	}

	return nil
}
