package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/domain/dto"
	"github.com/itsLeonB/cashback/internal/domain/mapper"
	"github.com/itsLeonB/ezutil/v2"
	"github.com/itsLeonB/ungerr"
)

type friendDetailsServiceImpl struct {
	debtSvc       DebtService
	profileSvc    ProfileService
	friendshipSvc FriendshipService
}

func NewFriendDetailsService(
	debtSvc DebtService,
	profileSvc ProfileService,
	friendshipSvc FriendshipService,
) FriendDetailsService {
	return &friendDetailsServiceImpl{
		debtSvc,
		profileSvc,
		friendshipSvc,
	}
}

func (fds *friendDetailsServiceImpl) GetDetails(ctx context.Context, profileID, friendshipID uuid.UUID) (dto.FriendDetailsResponse, error) {
	response, err := fds.friendshipSvc.GetDetails(ctx, profileID, friendshipID)
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
		return fds.returnRedirectResponse(ctx, profileID, friendProfile.RealProfileID)
	}

	debtTransactions, userAssociatedIDs, err := fds.debtSvc.GetAllByProfileIDs(ctx, profileID, friendProfileID)
	if err != nil {
		return dto.FriendDetailsResponse{}, err
	}

	return mapper.MapToFriendDetailsResponse(response, debtTransactions, userAssociatedIDs)
}

func (fds *friendDetailsServiceImpl) returnRedirectResponse(
	ctx context.Context,
	profileID,
	friendRealProfileID uuid.UUID,
) (dto.FriendDetailsResponse, error) {
	realFriendships, err := fds.friendshipSvc.GetAll(ctx, friendRealProfileID)
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
	if realFriendshipID == uuid.Nil {
		return dto.FriendDetailsResponse{}, ungerr.Unknownf("real friendship not found. friendRealProfileID: %s", friendRealProfileID)
	}

	return dto.FriendDetailsResponse{
		RedirectToRealFriendship: realFriendshipID,
	}, nil
}
