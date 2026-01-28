package repository

import (
	"context"

	"github.com/itsLeonB/cashback/internal/domain/entity"
)

type NotificationRepository interface {
	New(ctx context.Context, notification entity.Notification) (entity.Notification, error)
}
