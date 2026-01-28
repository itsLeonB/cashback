package message

import "github.com/google/uuid"

type DebtCreated struct {
	ID               uuid.UUID `json:"id"`
	CreatorProfileID uuid.UUID `json:"creatorProfileId"`
}

func (DebtCreated) Type() string {
	return "debt-created"
}

type DebtCreatedMetadata struct {
	FriendshipID uuid.UUID `json:"friendshipId"`
	FriendName   string    `json:"friendName"`
}
