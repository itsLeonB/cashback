package friendship

import (
	"time"

	"github.com/google/uuid"
	"github.com/itsLeonB/orcashtrator/internal/appconstant"
)

type Friendship struct {
	ID            uuid.UUID
	Type          appconstant.FriendshipType
	ProfileID     uuid.UUID
	ProfileName   string
	ProfileAvatar string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     time.Time
}

type FriendDetails struct {
	ID         uuid.UUID
	ProfileID  uuid.UUID
	Name       string
	Type       appconstant.FriendshipType
	Email      string
	Phone      string
	Avatar     string
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  time.Time
	ProfileID1 uuid.UUID
	ProfileID2 uuid.UUID
}
