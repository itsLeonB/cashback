package dto

import (
	"time"

	"github.com/google/uuid"
)

type FriendshipRequestResponse struct {
	ID               uuid.UUID `json:"id"`
	SenderAvatar     string    `json:"senderAvatar"`
	SenderName       string    `json:"senderName"`
	RecipientAvatar  string    `json:"recipientAvatar"`
	RecipientName    string    `json:"recipientName"`
	CreatedAt        time.Time `json:"createdAt"`
	BlockedAt        time.Time `json:"blockedAt"`
	IsSentByUser     bool      `json:"isSentByUser"`
	IsReceivedByUser bool      `json:"isReceivedByUser"`
	IsBlocked        bool      `json:"isBlocked"`
}
