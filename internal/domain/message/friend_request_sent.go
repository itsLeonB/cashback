package message

import "github.com/google/uuid"

type FriendRequestSent struct {
	ID uuid.UUID `json:"id"`
}

func (FriendRequestSent) Type() string {
	return "friend-request-sent"
}
