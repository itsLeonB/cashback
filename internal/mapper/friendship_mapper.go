package mapper

import (
	"github.com/google/uuid"
	"github.com/itsLeonB/orcashtrator/internal/domain/friendship"
	"github.com/itsLeonB/orcashtrator/internal/dto"
)

func MapToFriendDetailsResponse(
	userProfileID uuid.UUID,
	friendDetails friendship.FriendDetails,
	debtTransactions []dto.DebtTransactionResponse,
) (dto.FriendDetailsResponse, error) {
	txs := debtTransactions
	if txs == nil {
		txs = make([]dto.DebtTransactionResponse, 0)
	}

	return dto.FriendDetailsResponse{
		Friend: dto.FriendDetails{
			ID:        friendDetails.ID,
			ProfileID: friendDetails.ProfileID,
			Name:      friendDetails.Name,
			Type:      friendDetails.Type,
			Email:     friendDetails.Email,
			Phone:     friendDetails.Phone,
			Avatar:    friendDetails.Avatar,
			CreatedAt: friendDetails.CreatedAt,
			UpdatedAt: friendDetails.UpdatedAt,
			DeletedAt: friendDetails.DeletedAt,
		},
		Balance:      MapToFriendBalanceSummary(userProfileID, txs),
		Transactions: txs,
	}, nil
}

func FriendshipToResponse(fs friendship.Friendship) dto.FriendshipResponse {
	return dto.FriendshipResponse{
		ID:            fs.ID,
		Type:          fs.Type,
		ProfileID:     fs.ProfileID,
		ProfileName:   fs.ProfileName,
		ProfileAvatar: fs.ProfileAvatar,
		CreatedAt:     fs.CreatedAt,
		UpdatedAt:     fs.UpdatedAt,
		DeletedAt:     fs.DeletedAt,
	}
}
