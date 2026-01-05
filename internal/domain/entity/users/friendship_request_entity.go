package users

import (
	"database/sql"

	"github.com/google/uuid"
	"github.com/itsLeonB/go-crud"
)

type FriendshipRequest struct {
	crud.BaseEntity
	SenderProfileID    uuid.UUID
	RecipientProfileID uuid.UUID
	BlockedAt          sql.NullTime
	SenderProfile      UserProfile
	RecipientProfile   UserProfile
}
