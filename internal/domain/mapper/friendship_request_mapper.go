package mapper

import (
	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/domain/dto"
	"github.com/itsLeonB/cashback/internal/domain/entity/users"
)

func GetFriendshipRequestSimpleMapper(userProfileID uuid.UUID) func(users.FriendshipRequest) dto.FriendshipRequestResponse {
	return func(r users.FriendshipRequest) dto.FriendshipRequestResponse {
		return FriendshipRequestToResponse(r, userProfileID)
	}
}

func FriendshipRequestToResponse(fr users.FriendshipRequest, userProfileID uuid.UUID) dto.FriendshipRequestResponse {
	return dto.FriendshipRequestResponse{
		BaseDTO:          BaseToDTO(fr.BaseEntity),
		SenderAvatar:     fr.SenderProfile.Avatar,
		SenderName:       fr.SenderProfile.Name,
		RecipientAvatar:  fr.RecipientProfile.Avatar,
		RecipientName:    fr.RecipientProfile.Name,
		BlockedAt:        fr.BlockedAt.Time,
		IsSentByUser:     fr.SenderProfile.ID == userProfileID || fr.RecipientProfile.ID != userProfileID,
		IsReceivedByUser: fr.RecipientProfile.ID == userProfileID || fr.SenderProfile.ID != userProfileID,
		IsBlocked:        fr.BlockedAt.Valid,
	}
}
