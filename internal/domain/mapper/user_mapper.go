package mapper

import (
	"time"

	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/appconstant"
	"github.com/itsLeonB/cashback/internal/domain/dto"
	"github.com/itsLeonB/cashback/internal/domain/entity/users"
)

func SessionToAuthData(session users.Session) map[string]any {
	return map[string]any{
		appconstant.ContextUserID.String():    session.ID,
		appconstant.ContextSessionID.String(): session.ID,
		appconstant.ContextExp.String():       time.Now().Add(15 * time.Minute).Unix(),
		appconstant.ContextIat.String():       time.Now(),
	}
}

func UserToResponse(user users.User) dto.UserResponse {
	return dto.UserResponse{
		BaseDTO: BaseToDTO(user.BaseEntity),
		Email:   user.Email,
		Profile: ProfileToResponse(user.Profile, user.Email, nil, uuid.Nil),
	}
}
