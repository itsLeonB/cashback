package mapper

import (
	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/appconstant"
	"github.com/itsLeonB/cashback/internal/domain/dto"
	"github.com/itsLeonB/cashback/internal/domain/entity/users"
)

func UserToAuthData(user users.User) map[string]any {
	return map[string]any{
		appconstant.ContextUserID.String(): user.ID,
	}
}

func UserToResponse(user users.User) dto.UserResponse {
	return dto.UserResponse{
		Email:   user.Email,
		Profile: ProfileToResponse(user.Profile, user.Email, nil, uuid.Nil),
	}
}
