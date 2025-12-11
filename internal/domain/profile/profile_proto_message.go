package profile

import "github.com/google/uuid"

type CreateRequest struct {
	UserID uuid.UUID
	Name   string `validate:"required,min=3"`
}

type UpdateRequest struct {
	ID     uuid.UUID
	UserID uuid.UUID
	Name   string
	Avatar string
}

type AssociateRequest struct {
	UserProfileID uuid.UUID
	RealProfileID uuid.UUID
	AnonProfileID uuid.UUID
}
