package profile

import (
	"time"

	"github.com/google/uuid"
)

type Profile struct {
	ID                       uuid.UUID
	UserID                   uuid.UUID
	Name                     string
	Avatar                   string
	Email                    string
	AssociatedAnonProfileIDs []uuid.UUID
	RealProfileID            uuid.UUID
	CreatedAt                time.Time
	UpdatedAt                time.Time
	DeletedAt                time.Time
}
