package dto

import (
	"github.com/google/uuid"
)

type ProfileResponse struct {
	BaseDTO
	UserID                   uuid.UUID   `json:"userId"`
	Name                     string      `json:"name"`
	Avatar                   string      `json:"avatar"`
	Email                    string      `json:"email"`
	IsAnonymous              bool        `json:"isAnonymous"`
	AssociatedAnonProfileIDs []uuid.UUID `json:"associatedAnonProfileIds"`
	RealProfileID            uuid.UUID   `json:"realProfileId"`
}

type UpdateProfileRequest struct {
	Name string `json:"name" binding:"required,min=3,max=255"`
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
	UserID uuid.UUID
	Name   string `validate:"required,min=1,max=255"`
	Avatar string
}
