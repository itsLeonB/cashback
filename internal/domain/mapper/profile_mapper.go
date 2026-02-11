package mapper

import (
	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/domain/dto"
	"github.com/itsLeonB/cashback/internal/domain/entity/users"
)

func ProfileToResponse(profile users.UserProfile, email string, anonProfileIDs []uuid.UUID, realProfileID uuid.UUID) dto.ProfileResponse {
	associatedAnonProfileIDs := anonProfileIDs
	if len(associatedAnonProfileIDs) < 1 {
		for _, anonProfile := range profile.RelatedAnonProfiles {
			associatedAnonProfileIDs = append(associatedAnonProfileIDs, anonProfile.AnonProfileID)
		}
	}

	if realProfileID == uuid.Nil {
		realProfileID = profile.RelatedRealProfile.RealProfileID
	}

	return dto.ProfileResponse{
		BaseDTO:                  BaseToDTO(profile.BaseEntity),
		UserID:                   profile.UserID.UUID,
		Name:                     profile.Name,
		Avatar:                   profile.Avatar,
		Email:                    email,
		IsAnonymous:              !profile.UserID.Valid,
		AssociatedAnonProfileIDs: associatedAnonProfileIDs,
		RealProfileID:            realProfileID,
		CurrentSubscription:      SubscriptionToResponse(profile.CurrentSubscription),
	}
}

func ToSimpleProfile(profile dto.ProfileResponse, userProfileID uuid.UUID) dto.SimpleProfile {
	return dto.SimpleProfile{
		ID:     profile.ID,
		Name:   profile.Name,
		Avatar: profile.Avatar,
		IsUser: profile.ID == userProfileID,
	}
}

func ProfileToSimple(profile users.UserProfile, userProfileID uuid.UUID) dto.SimpleProfile {
	return dto.SimpleProfile{
		ID:     profile.ID,
		Name:   profile.Name,
		Avatar: profile.Avatar,
		IsUser: profile.ID == userProfileID,
	}
}

func SimpleProfileMapper(userProfileID uuid.UUID) func(dto.ProfileResponse) dto.SimpleProfile {
	return func(pr dto.ProfileResponse) dto.SimpleProfile {
		return ToSimpleProfile(pr, userProfileID)
	}
}
