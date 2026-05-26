package dto

import (
	"github.com/google/uuid"
)

type ProfileResponse struct {
	BaseDTO
	UserID                   uuid.UUID            `json:"userId"`
	Name                     string               `json:"name"`
	Avatar                   string               `json:"avatar"`
	Email                    string               `json:"email"`
	HomeCurrency             string               `json:"homeCurrency"`
	IsAnonymous              bool                 `json:"isAnonymous"`
	AssociatedAnonProfileIDs []uuid.UUID          `json:"associatedAnonProfileIds"`
	RealProfileID            uuid.UUID            `json:"realProfileId"`
	CurrentSubscription      SubscriptionResponse `json:"currentSubscription"`
	IsOnboarded              bool                 `json:"isOnboarded"`
}

type UpdateProfileRequest struct {
	ID           uuid.UUID `json:"-"`
	Name         string    `json:"name" binding:"required,min=3,max=255"`
	HomeCurrency string    `json:"homeCurrency" binding:"required,len=3"`
}

type SearchRequest struct {
	Query string `form:"query" binding:"required,min=3,max=255"`
}

type AssociateProfileRequest struct {
	RealProfileID uuid.UUID `json:"realProfileId" binding:"required"`
	AnonProfileID uuid.UUID `json:"anonProfileId" binding:"required"`
}

type SimpleProfile struct {
	ID     uuid.UUID `json:"id"`
	Name   string    `json:"name"`
	Avatar string    `json:"avatar"`
	IsUser bool      `json:"isUser"`
}

type NewProfileRequest struct {
	UserID       uuid.UUID
	Name         string `validate:"required,min=1,max=255"`
	Avatar       string
	HomeCurrency string `validate:"required,len=3"`
	GenerateSlug bool
}
