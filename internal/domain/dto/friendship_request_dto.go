package dto

import (
	"time"
)

type FriendshipRequestResponse struct {
	BaseDTO
	SenderAvatar     string    `json:"senderAvatar"`
	SenderName       string    `json:"senderName"`
	RecipientAvatar  string    `json:"recipientAvatar"`
	RecipientName    string    `json:"recipientName"`
	BlockedAt        time.Time `json:"blockedAt"`
	IsSentByUser     bool      `json:"isSentByUser"`
	IsReceivedByUser bool      `json:"isReceivedByUser"`
	IsBlocked        bool      `json:"isBlocked"`
}
