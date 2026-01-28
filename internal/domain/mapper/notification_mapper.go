package mapper

import (
	"github.com/itsLeonB/cashback/internal/domain/dto"
	"github.com/itsLeonB/cashback/internal/domain/entity"
)

func NotificationToResponse(notification entity.Notification) dto.NotificationResponse {
	resp := dto.NotificationResponse{
		ID:         notification.ID,
		Type:       notification.Type,
		EntityType: notification.EntityType,
		EntityID:   notification.EntityID,
		Metadata:   notification.Metadata,
		CreatedAt:  notification.CreatedAt,
	}

	if notification.ReadAt.Valid {
		resp.ReadAt = notification.ReadAt.Time
	}

	return resp
}
