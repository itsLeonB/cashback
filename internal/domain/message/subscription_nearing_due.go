package message

import "github.com/google/uuid"

type SubscriptionNearingDue struct {
	UserIDs []uuid.UUID `json:"userIds"`
}

func (SubscriptionNearingDue) Type() string {
	return "subscription-nearing-due"
}
