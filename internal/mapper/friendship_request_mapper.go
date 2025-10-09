package mapper

import (
	"github.com/google/uuid"
	"github.com/itsLeonB/orcashtrator/internal/domain/friendship"
	"github.com/itsLeonB/orcashtrator/internal/dto"
)

func GetFriendshipRequestSimpleMapper(userProfileID uuid.UUID) func(friendship.Request) dto.FriendshipRequestResponse {
	return func(r friendship.Request) dto.FriendshipRequestResponse {
		return friendshipRequestToResponse(r, userProfileID)
	}
}

func friendshipRequestToResponse(fr friendship.Request, userProfileID uuid.UUID) dto.FriendshipRequestResponse {
	return dto.FriendshipRequestResponse{
		ID:               fr.ID,
		SenderAvatar:     fr.Sender.Avatar,
		SenderName:       fr.Sender.Name,
		RecipientAvatar:  fr.Recipient.Avatar,
		RecipientName:    fr.Recipient.Name,
		Message:          fr.Message,
		CreatedAt:        fr.CreatedAt,
		BlockedAt:        fr.BlockedAt,
		IsSentByUser:     fr.Sender.ID == userProfileID,
		IsReceivedByUser: fr.Recipient.ID == userProfileID,
		IsBlocked:        !fr.BlockedAt.IsZero(),
	}
}
