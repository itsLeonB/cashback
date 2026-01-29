package entity

import (
	"database/sql"

	"github.com/google/uuid"
	"github.com/itsLeonB/go-crud"
	"gorm.io/datatypes"
)

type PushSubscription struct {
	crud.BaseEntity
	ProfileID uuid.UUID
	Endpoint  string
	Keys      datatypes.JSON
	UserAgent sql.NullString
}

type PushSubscriptionKeys struct {
	P256dh string `json:"p256dh"`
	Auth   string `json:"auth"`
}
