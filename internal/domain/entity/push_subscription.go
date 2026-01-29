package entity

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type PushSubscription struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:uuidv7()"`
	ProfileID uuid.UUID
	Endpoint  string
	Keys      datatypes.JSON
	UserAgent sql.NullString
	CreatedAt time.Time
}

type PushSubscriptionKeys struct {
	P256dh string `json:"p256dh"`
	Auth   string `json:"auth"`
}
