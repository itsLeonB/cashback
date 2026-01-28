package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/domain/entity"
)

type NotificationRepository interface {
	New(ctx context.Context, notification entity.Notification) (entity.Notification, error)
	GetByProfileID(ctx context.Context, profileID uuid.UUID, unreadOnly bool) ([]entity.Notification, error)
	MarkAsRead(ctx context.Context, profileID, notificationID uuid.UUID) error
	MarkAllAsRead(ctx context.Context, profileID uuid.UUID) error
	InsertMany(ctx context.Context, notifications []entity.Notification) ([]entity.Notification, error)
}
