package message

import "github.com/google/uuid"

type FriendRequestAccepted struct {
	SenderProfileID uuid.UUID `json:"senderProfileId"`
	FriendshipID    uuid.UUID `json:"friendshipId"`
}

func (FriendRequestAccepted) Type() string {
	return "friend-request-accepted"
}

type FriendRequestAcceptedMetadata struct {
	FriendName string `json:"friendName"`
}
