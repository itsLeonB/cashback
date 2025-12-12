package dto

import (
	"time"

	"github.com/google/uuid"
)

type ProfileResponse struct {
	ID                       uuid.UUID   `json:"id"`
	UserID                   uuid.UUID   `json:"userId"`
	Name                     string      `json:"name"`
	Avatar                   string      `json:"avatar"`
	Email                    string      `json:"email"`
	CreatedAt                time.Time   `json:"createdAt"`
	UpdatedAt                time.Time   `json:"updatedAt"`
	DeletedAt                time.Time   `json:"deletedAt,omitzero"`
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
