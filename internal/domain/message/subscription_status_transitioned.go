package message

import (
	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/domain/entity/monetization"
)

type SubscriptionStatusTransitioned struct {
	ID     uuid.UUID                       `json:"id"`
	Status monetization.SubscriptionStatus `json:"status"`
}

func (SubscriptionStatusTransitioned) Type() string {
	return "subscription-status-transitioned"
}
