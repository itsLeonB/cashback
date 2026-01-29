package mapper

import (
	"github.com/itsLeonB/cashback/internal/core/logger"
	"github.com/itsLeonB/cashback/internal/domain/dto"
	"github.com/itsLeonB/cashback/internal/domain/entity"
	"github.com/itsLeonB/cashback/internal/domain/mapper/notification"
)

func NotificationToResponse(n entity.Notification) dto.NotificationResponse {
	resp := dto.NotificationResponse{
		ID:         n.ID,
		Type:       n.Type,
		EntityType: n.EntityType,
		EntityID:   n.EntityID,
		Metadata:   n.Metadata,
		CreatedAt:  n.CreatedAt,
	}

	if n.ReadAt.Valid {
		resp.ReadAt = n.ReadAt.Time
	}

	if title, err := notification.ResolveTitle(n); err != nil {
		logger.Errorf("error resolving notification title: %v", err)
		resp.Title = "Notification"
	} else {
		resp.Title = title
	}

	return resp
}
