package mapper

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/domain/dto"
	"github.com/itsLeonB/cashback/internal/domain/entity/debts"
	"github.com/itsLeonB/cashback/internal/domain/entity/users"
	"github.com/itsLeonB/ezutil/v2"
	"github.com/itsLeonB/ungerr"
)

func SelectProfiles(userProfileID uuid.UUID, friendship users.Friendship) (users.UserProfile, users.UserProfile, error) {
	switch userProfileID {
	case friendship.ProfileID1:
		return friendship.Profile1, friendship.Profile2, nil
	case friendship.ProfileID2:
		return friendship.Profile2, friendship.Profile1, nil
	default:
		return users.UserProfile{}, users.UserProfile{}, ungerr.Unknown(fmt.Sprintf(
			"mismatched user profile ID: %s with friendship ID: %s",
			userProfileID,
			friendship.ID,
		))
	}
}

func FriendshipToResponse(userProfileID uuid.UUID, friendship users.Friendship) (dto.FriendshipResponse, error) {
	_, friendProfile, err := SelectProfiles(userProfileID, friendship)
	if err != nil {
		return dto.FriendshipResponse{}, err
	}

	return dto.FriendshipResponse{
		BaseDTO:       BaseToDTO(friendship.BaseEntity),
		Type:          friendship.Type,
		ProfileID:     friendProfile.ID,
		ProfileName:   friendProfile.Name,
		ProfileAvatar: friendProfile.Avatar,
	}, nil
}

func OrderProfilesToFriendship(userProfile, friendProfile dto.ProfileResponse) (users.Friendship, error) {
	switch ezutil.CompareUUID(userProfile.ID, friendProfile.ID) {
	case 1:
		return users.Friendship{
			ProfileID1: friendProfile.ID,
			ProfileID2: userProfile.ID,
		}, nil
	case -1:
		return users.Friendship{
			ProfileID1: userProfile.ID,
			ProfileID2: friendProfile.ID,
		}, nil
	default:
		return users.Friendship{}, ungerr.Unknown("both IDs are equal, cannot create friendship")
	}
}

func MapToFriendshipWithProfile(userProfileID uuid.UUID, friendship users.Friendship) (dto.FriendshipWithProfile, error) {
	friendshipResponse, err := FriendshipToResponse(userProfileID, friendship)
	if err != nil {
		return dto.FriendshipWithProfile{}, err
	}

	userProfile, friendProfile, err := SelectProfiles(userProfileID, friendship)
	if err != nil {
		return dto.FriendshipWithProfile{}, err
	}

	return dto.FriendshipWithProfile{
		Friendship:    friendshipResponse,
		UserProfile:   ProfileToResponse(userProfile, "", nil, uuid.Nil, dto.SubscriptionResponse{}),
		FriendProfile: ProfileToResponse(friendProfile, "", nil, uuid.Nil, dto.SubscriptionResponse{}),
	}, nil
}

func MapToFriendDetails(userProfileID uuid.UUID, friendship users.Friendship) (dto.FriendDetails, error) {
	friendshipWithProfile, err := MapToFriendshipWithProfile(userProfileID, friendship)
	if err != nil {
		return dto.FriendDetails{}, err
	}

	friendProfile := friendshipWithProfile.FriendProfile

	return dto.FriendDetails{
		BaseDTO:    friendProfile.BaseDTO,
		ProfileID:  friendProfile.ID,
		Name:       friendProfile.Name,
		Email:      friendProfile.Email,
		Avatar:     friendProfile.Avatar,
		Type:       friendship.Type,
		ProfileID1: friendship.ProfileID1,
		ProfileID2: friendship.ProfileID2,
	}, nil
}

func MapToFriendDetailsResponse(
	friendDetails dto.FriendDetails,
	debtTransactions []debts.DebtTransaction,
	userAssociatedIDs []uuid.UUID,
) (dto.FriendDetailsResponse, error) {
	return dto.FriendDetailsResponse{
		Friend:  friendDetails,
		Balance: MapToFriendBalanceSummary(debtTransactions, userAssociatedIDs),
	}, nil
}
