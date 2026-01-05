package users

import (
	"github.com/google/uuid"
	"github.com/itsLeonB/go-crud"
)

type RelatedProfile struct {
	crud.BaseEntity
	RealProfileID uuid.UUID
	AnonProfileID uuid.UUID
}
