package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/itsLeonB/ezutil/v2"
	"github.com/itsLeonB/orcashtrator/internal/domain/friendship"
	"github.com/itsLeonB/orcashtrator/internal/dto"
	"github.com/itsLeonB/orcashtrator/internal/mapper"
	"github.com/itsLeonB/ungerr"
)

type friendDetailsServiceImpl struct {
	friendshipClient friendship.FriendshipClient
	debtSvc          DebtService
	profileSvc       ProfileService
}

func NewFriendDetailsService(
	friendshipClient friendship.FriendshipClient,
	debtSvc DebtService,
	profileSvc ProfileService,
) FriendDetailsService {
	return &friendDetailsServiceImpl{
		friendshipClient,
		debtSvc,
		profileSvc,
	}
}

func (fds *friendDetailsServiceImpl) GetDetails(ctx context.Context, profileID, friendshipID uuid.UUID) (dto.FriendDetailsResponse, error) {
	request := friendship.GetDetailsRequest{
		ProfileID:    profileID,
		FriendshipID: friendshipID,
	}

	response, err := fds.friendshipClient.GetDetails(ctx, request)
	if err != nil {
		return dto.FriendDetailsResponse{}, err
	}

	// Ensure the requester is part of the friendship
	if ezutil.CompareUUID(profileID, response.ProfileID1) != 0 && ezutil.CompareUUID(profileID, response.ProfileID2) != 0 {
		return dto.FriendDetailsResponse{}, ungerr.ForbiddenError(fmt.Sprintf("profileID %s is not part of friendship %s", profileID, friendshipID))
	}

	// Pick the friend’s profile ID
	friendProfileID := response.ProfileID2
	if ezutil.CompareUUID(profileID, response.ProfileID2) == 0 {
		friendProfileID = response.ProfileID1
	}

	friendProfile, err := fds.profileSvc.GetByID(ctx, friendProfileID)
	if err != nil {
		return dto.FriendDetailsResponse{}, err
	}

	if friendProfile.RealProfileID != uuid.Nil {
		realFriendships, err := fds.friendshipClient.GetAll(ctx, friendProfile.RealProfileID)
		if err != nil {
			return dto.FriendDetailsResponse{}, err
		}

		var realFriendshipID uuid.UUID
		for _, realFriendship := range realFriendships {
			if ezutil.CompareUUID(realFriendship.ProfileID, profileID) == 0 {
				realFriendshipID = realFriendship.ID
				break
			}
		}

		return dto.FriendDetailsResponse{
			RedirectToRealFriendship: realFriendshipID,
		}, nil
	}

	debtTransactions, err := fds.debtSvc.GetAllByProfileIDs(ctx, profileID, friendProfileID)
	if err != nil {
		return dto.FriendDetailsResponse{}, err
	}

	return mapper.MapToFriendDetailsResponse(profileID, response, debtTransactions)
}
