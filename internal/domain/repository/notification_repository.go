package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/domain/entity"
	"github.com/itsLeonB/go-crud"
)

type NotificationRepository interface {
	crud.Repository[entity.Notification]
	New(ctx context.Context, notification entity.Notification) (entity.Notification, error)
	GetByProfileID(ctx context.Context, profileID uuid.UUID, unreadOnly bool) ([]entity.Notification, error)
	MarkAsRead(ctx context.Context, profileID, notificationID uuid.UUID) error
	MarkAllAsRead(ctx context.Context, profileID uuid.UUID) error
	CreateMany(ctx context.Context, notifications []entity.Notification) ([]entity.Notification, error)
	// RepointProfile repoints notifications referencing anonProfileID onto realProfileID. If
	// realProfileID already has a matching (type, entity_type, entity_id) notification, the
	// anonymous row is dropped instead of violating the unique constraint.
	RepointProfile(ctx context.Context, anonProfileID, realProfileID uuid.UUID) error
}
