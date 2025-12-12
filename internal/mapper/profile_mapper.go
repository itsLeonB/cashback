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
