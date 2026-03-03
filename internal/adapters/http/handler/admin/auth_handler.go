package admin

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/itsLeonB/cashback/internal/domain/dto"
	"github.com/itsLeonB/cashback/internal/domain/service/admin"
	"github.com/itsLeonB/ginkgo/pkg/server"
)

type AuthHandler struct {
	authSvc admin.AuthService
}

func (ah *AuthHandler) HandleRegister() gin.HandlerFunc {
	return server.Handler("AuthHandler.HandleRegister", http.StatusCreated, func(ctx *gin.Context) (any, error) {
		request, err := server.BindJSON[dto.RegisterRequest](ctx)
		if err != nil {
			return nil, err
		}

		return nil, ah.authSvc.Register(ctx.Request.Context(), request)
	})
}

func (ah *AuthHandler) HandleLogin() gin.HandlerFunc {
	return server.Handler("AuthHandler.HandleLogin", http.StatusOK, func(ctx *gin.Context) (any, error) {
		request, err := server.BindJSON[dto.InternalLoginRequest](ctx)
		if err != nil {
			return nil, err
		}

		return ah.authSvc.Login(ctx.Request.Context(), request)
	})
}

func (ah *AuthHandler) HandleMe() gin.HandlerFunc {
	return server.Handler("AuthHandler.HandleMe", http.StatusOK, func(ctx *gin.Context) (any, error) {
		userID, err := getUserID(ctx)
		if err != nil {
			return nil, err
		}

		return ah.authSvc.Me(ctx.Request.Context(), userID)
	})
}
