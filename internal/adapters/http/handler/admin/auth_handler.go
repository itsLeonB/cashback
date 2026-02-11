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
	return server.Handler(http.StatusCreated, func(ctx *gin.Context) (any, error) {
		request, err := server.BindJSON[dto.RegisterRequest](ctx)
		if err != nil {
			return nil, err
		}

		return nil, ah.authSvc.Register(ctx, request)
	})
}

func (ah *AuthHandler) HandleLogin() gin.HandlerFunc {
	return server.Handler(http.StatusOK, func(ctx *gin.Context) (any, error) {
		request, err := server.BindJSON[dto.InternalLoginRequest](ctx)
		if err != nil {
			return nil, err
		}

		return ah.authSvc.Login(ctx, request)
	})
}
