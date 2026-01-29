package entity

import (
	"database/sql"

	"github.com/google/uuid"
	"github.com/itsLeonB/go-crud"
	"gorm.io/datatypes"
)

type Notification struct {
	crud.BaseEntity
	ProfileID  uuid.UUID
	Type       string
	EntityType string
	EntityID   uuid.UUID
	Metadata   datatypes.JSON
	ReadAt     sql.NullTime
	PushedAt   sql.NullTime
}
