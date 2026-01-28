package entity

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type Notification struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey;default:uuidv7()"`
	ProfileID  uuid.UUID
	Type       string
	EntityType string
	EntityID   uuid.UUID
	Metadata   datatypes.JSON
	ReadAt     sql.NullTime
	CreatedAt  time.Time
}
