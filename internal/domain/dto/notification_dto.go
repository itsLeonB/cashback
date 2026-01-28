package dto

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type NotificationResponse struct {
	ID         uuid.UUID      `json:"id"`
	Type       string         `json:"type"`
	EntityType string         `json:"entityType"`
	EntityID   uuid.UUID      `json:"entityId"`
	Metadata   datatypes.JSON `json:"metadata"`
	ReadAt     time.Time      `json:"readAt,omitzero"`
	CreatedAt  time.Time      `json:"createdAt"`
}
