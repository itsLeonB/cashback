package dto

import (
	"time"

	"github.com/google/uuid"
)

type SubscriptionResponse struct {
	BaseDTO
	ProfileID  uuid.UUID `json:"profileId"`
	EndsAt     time.Time `json:"endsAt,omitzero"`
	CanceledAt time.Time `json:"canceledAt,omitzero"`
}
