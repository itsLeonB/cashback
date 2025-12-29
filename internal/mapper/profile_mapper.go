package mapper

import (
	"github.com/google/uuid"
	"github.com/itsLeonB/orcashtrator/internal/domain/profile"
	"github.com/itsLeonB/orcashtrator/internal/dto"
)

func ProfileToResponse(response profile.Profile) dto.ProfileResponse {
	return dto.ProfileResponse{
		ID:                       response.ID,
		UserID:                   response.UserID,
		Name:                     response.Name,
		Avatar:                   response.Avatar,
		Email:                    response.Email,
		CreatedAt:                response.CreatedAt,
		UpdatedAt:                response.UpdatedAt,
		DeletedAt:                response.DeletedAt,
		IsAnonymous:              response.UserID == uuid.Nil,
		AssociatedAnonProfileIDs: response.AssociatedAnonProfileIDs,
		RealProfileID:            response.RealProfileID,
	}
}

func ProfileResponseToParticipant(resp dto.ProfileResponse, userProfileID uuid.UUID) dto.Participant {
	return dto.Participant{
		ProfileID: resp.ID,
		Name:      resp.Name,
		Avatar:    resp.Avatar,
		IsUser:    userProfileID == resp.ID,
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

func SimpleProfileMapper(userProfileID uuid.UUID) func(dto.ProfileResponse) dto.SimpleProfile {
	return func(pr dto.ProfileResponse) dto.SimpleProfile {
		return ToSimpleProfile(pr, userProfileID)
	}
}
