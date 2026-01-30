package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/appconstant"
	"github.com/itsLeonB/cashback/internal/domain/dto"
	"github.com/itsLeonB/cashback/internal/domain/service"
	"github.com/itsLeonB/ginkgo/pkg/server"
	"github.com/itsLeonB/ungerr"
)

type AuthHandler struct {
	authService    service.AuthService
	oAuthService   service.OAuthService
	sessionService service.SessionService
}

func NewAuthHandler(
	authService service.AuthService,
	oAuthService service.OAuthService,
	sessionService service.SessionService,
) *AuthHandler {
	return &AuthHandler{
		authService,
		oAuthService,
		sessionService,
	}
}

func (ah *AuthHandler) HandleRegister() gin.HandlerFunc {
	return server.Handler(http.StatusCreated, func(ctx *gin.Context) (any, error) {
		request, err := server.BindJSON[dto.RegisterRequest](ctx)
		if err != nil {
			return nil, err
		}

		response, err := ah.authService.Register(ctx, request)
		if err != nil {
			return nil, err
		}

		return response, nil
	})
}

func (ah *AuthHandler) HandleInternalLogin() gin.HandlerFunc {
	return server.Handler(http.StatusOK, func(ctx *gin.Context) (any, error) {
		request, err := server.BindJSON[dto.InternalLoginRequest](ctx)
		if err != nil {
			return nil, err
		}

		response, err := ah.authService.InternalLogin(ctx, request)
		if err != nil {
			return nil, err
		}

		return response, nil
	})
}

func (ah *AuthHandler) HandleOAuth2Login() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		provider, err := ah.getProvider(ctx)
		if err != nil {
			_ = ctx.Error(ungerr.BadRequestError("missing oauth provider"))
			return
		}

		url, err := ah.oAuthService.GetOAuthURL(ctx, provider)
		if err != nil {
			_ = ctx.Error(err)
			return
		}

		ctx.Redirect(http.StatusTemporaryRedirect, url)
	}
}

func (ah *AuthHandler) HandleOAuth2Callback() gin.HandlerFunc {
	return server.Handler(http.StatusOK, func(ctx *gin.Context) (any, error) {
		provider, err := ah.getProvider(ctx)
		if err != nil {
			return nil, err
		}
		code := ctx.Query("code")
		state := ctx.Query("state")

		response, err := ah.oAuthService.HandleOAuthCallback(ctx, dto.OAuthCallbackData{
			Provider: provider,
			Code:     code,
			State:    state,
		})
		if err != nil {
			return nil, err
		}

		return response, nil
	})
}

func (ah *AuthHandler) HandleVerifyRegistration() gin.HandlerFunc {
	return server.Handler(http.StatusOK, func(ctx *gin.Context) (any, error) {
		token := ctx.Query("token")
		if token == "" {
			return nil, ungerr.BadRequestError("missing token")
		}

		response, err := ah.authService.VerifyRegistration(ctx, token)
		if err != nil {
			return nil, err
		}

		return response, nil
	})
}

func (ah *AuthHandler) HandleSendPasswordReset() gin.HandlerFunc {
	return server.Handler(http.StatusCreated, func(ctx *gin.Context) (any, error) {
		request, err := server.BindJSON[dto.SendPasswordResetRequest](ctx)
		if err != nil {
			return nil, err
		}

		if err = ah.authService.SendPasswordReset(ctx, request.Email); err != nil {
			return nil, err
		}

		return nil, nil
	})
}

func (ah *AuthHandler) HandleResetPassword() gin.HandlerFunc {
	return server.Handler(http.StatusOK, func(ctx *gin.Context) (any, error) {
		request, err := server.BindJSON[dto.ResetPasswordRequest](ctx)
		if err != nil {
			return nil, err
		}

		response, err := ah.authService.ResetPassword(ctx, request.Token, request.Password)
		if err != nil {
			return nil, err
		}

		return response, nil
	})
}

func (ah *AuthHandler) HandleRefreshToken() gin.HandlerFunc {
	return server.Handler(http.StatusOK, func(ctx *gin.Context) (any, error) {
		request, err := server.BindJSON[dto.RefreshTokenRequest](ctx)
		if err != nil {
			return nil, err
		}

		return ah.sessionService.RefreshToken(ctx, request)
	})
}

func (ah *AuthHandler) getProvider(ctx *gin.Context) (string, error) {
	provider := ctx.Param(appconstant.ContextProvider.String())
	if provider == "" {
		return "", ungerr.BadRequestError("missing oauth provider")
	}
	return provider, nil
}

func (ah *AuthHandler) HandleLogout() gin.HandlerFunc {
	return server.Handler(http.StatusNoContent, func(ctx *gin.Context) (any, error) {
		sessionID, err := server.GetFromContext[uuid.UUID](ctx, appconstant.ContextSessionID.String())
		if err != nil {
			return nil, err
		}

		return nil, ah.authService.Logout(ctx, sessionID)
	})
}
