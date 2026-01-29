package dto

import "github.com/google/uuid"

type PushSubscriptionRequest struct {
	ProfileID uuid.UUID            `json:"-"`
	Endpoint  string               `json:"endpoint" binding:"required"`
	Keys      PushSubscriptionKeys `json:"keys" binding:"required"`
	UserAgent string               `json:"userAgent,omitempty"`
}

type PushSubscriptionKeys struct {
	P256dh string `json:"p256dh" binding:"required"`
	Auth   string `json:"auth" binding:"required"`
}

type PushUnsubscribeRequest struct {
	ProfileID uuid.UUID `json:"-"`
	Endpoint  string    `json:"endpoint" binding:"required"`
}
