package friendship

import (
	"time"

	"github.com/google/uuid"
	"github.com/itsLeonB/orcashtrator/internal/domain/profile"
)

type Request struct {
	ID        uuid.UUID
	Sender    profile.Profile
	Recipient profile.Profile
	CreatedAt time.Time
	BlockedAt time.Time
}
