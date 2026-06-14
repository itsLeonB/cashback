package mapper

import (
	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/appconstant"
	"github.com/itsLeonB/cashback/internal/domain/dto"
	"github.com/itsLeonB/cashback/internal/domain/entity/users"
	"github.com/itsLeonB/cashback/internal/domain/service/auth"
)

func SessionToAuthData(session auth.Session, fgpHash string) map[string]any {
	return map[string]any{
		appconstant.ContextUserID.String():      session.UserID,
		appconstant.ContextSessionID.String():   session.ID,
		appconstant.ContextFingerprint.String(): fgpHash,
	}
}

func UserToResponse(user users.User) dto.UserResponse {
	return dto.UserResponse{
		BaseDTO: BaseToDTO(user.BaseEntity),
		Email:   user.Email,
		Profile: ProfileToResponse(user.Profile, user.Email, nil, uuid.Nil, dto.SubscriptionResponse{}),
	}
}
