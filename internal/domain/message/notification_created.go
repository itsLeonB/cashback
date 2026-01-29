package message

import "github.com/google/uuid"

type NotificationCreated struct {
	ID uuid.UUID `json:"id"`
}

func (NotificationCreated) Type() string {
	return "notification-created"
}
