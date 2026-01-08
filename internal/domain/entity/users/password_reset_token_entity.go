package users

import (
	"time"

	"github.com/google/uuid"
	"github.com/itsLeonB/go-crud"
)

type PasswordResetToken struct {
	crud.BaseEntity
	UserID    uuid.UUID
	Token     string
	ExpiresAt time.Time
}

func (prt PasswordResetToken) IsValid() bool {
	return !prt.IsZero() && prt.ExpiresAt.After(time.Now())
}
