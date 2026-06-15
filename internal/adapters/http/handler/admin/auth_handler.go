package admin

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/itsLeonB/cashback/internal/core/util"
	"github.com/itsLeonB/cashback/internal/domain/dto"
	adminEntity "github.com/itsLeonB/cashback/internal/domain/entity/admin"
	"github.com/itsLeonB/go-authkit/authgin"
	"github.com/itsLeonB/go-crud"
	"github.com/itsLeonB/ginkgo/pkg/server"
	"github.com/itsLeonB/ungerr"
)

type AuthHandler struct {
	stateless *authgin.StatelessHandler
	userRepo  crud.Repository[adminEntity.User]
}

func (ah *AuthHandler) HandleRegister() gin.HandlerFunc {
	return ah.stateless.Register()
}

func (ah *AuthHandler) HandleLogin() gin.HandlerFunc {
	return ah.stateless.Login()
}

func (ah *AuthHandler) HandleMe() gin.HandlerFunc {
	return server.Handler("AuthHandler.HandleMe", http.StatusOK, func(ctx *gin.Context) (any, error) {
		userID, err := getUserID(ctx)
		if err != nil {
			return nil, err
		}
		spec := crud.Specification[adminEntity.User]{}
		spec.Model.ID = userID
		user, err := ah.userRepo.FindFirst(ctx.Request.Context(), spec)
		if err != nil {
			return nil, err
		}
		if user.IsZero() {
			return nil, ungerr.UnauthorizedError("user not found")
		}
		return dto.AdminMe{ID: user.ID, FullName: util.GetNameFromEmail(user.Email)}, nil
	})
}
