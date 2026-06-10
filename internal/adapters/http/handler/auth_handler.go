package handler

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/adapters/http/cookie"
	"github.com/itsLeonB/cashback/internal/appconstant"
	"github.com/itsLeonB/cashback/internal/core/otel"
	"github.com/itsLeonB/cashback/internal/domain/dto"
	"github.com/itsLeonB/cashback/internal/domain/service"
	_ "github.com/itsLeonB/ginkgo/pkg/response"
	"github.com/itsLeonB/ginkgo/pkg/server"
	"github.com/itsLeonB/ungerr"
)

type AuthHandler struct {
	authService    service.AuthService
	oAuthService   service.OAuthService
	sessionService service.SessionService
	cookieCfg      cookie.Config
}

func NewAuthHandler(
	authService service.AuthService,
	oAuthService service.OAuthService,
	sessionService service.SessionService,
	cookieCfg cookie.Config,
) *AuthHandler {
	return &AuthHandler{
		authService:    authService,
		oAuthService:   oAuthService,
		sessionService: sessionService,
		cookieCfg:      cookieCfg,
	}
}

func (ah *AuthHandler) setTokenCookies(ctx *gin.Context, tokenResp dto.TokenResponse) {
	cookie.SetAccessToken(ctx, ah.cookieCfg, tokenResp.Token)
	cookie.SetRefreshToken(ctx, ah.cookieCfg, tokenResp.RefreshToken)
	ah.setCSRFCookie(ctx)
}

func (ah *AuthHandler) setCSRFCookie(ctx *gin.Context) {
	b := make([]byte, 16)
	rand.Read(b)
	cookie.SetCSRFToken(ctx, ah.cookieCfg, hex.EncodeToString(b))
}

// HandleRegister godoc
// @Summary      Register a new user
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body dto.RegisterRequest true "Register payload"
// @Success      201  {object}  response.JSONResponse[dto.RegisterResponse]
// @Failure      400  {object}  map[string]any
// @Router       /auth/register [post]

func (ah *AuthHandler) HandleRegister() gin.HandlerFunc {
	return server.Handler("AuthHandler.HandleRegister", http.StatusCreated, func(ctx *gin.Context) (any, error) {
		request, err := server.BindJSON[dto.RegisterRequest](ctx)
		if err != nil {
			return nil, err
		}

		return ah.authService.Register(ctx.Request.Context(), request)
	})
}

// HandleInternalLogin godoc
// @Summary      Login with email and password
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body dto.InternalLoginRequest true "Login payload"
// @Success      200  {object}  response.JSONResponse[map[string]string]
// @Failure      400  {object}  map[string]any
// @Failure      401  {object}  map[string]any
// @Router       /auth/login [post]
func (ah *AuthHandler) HandleInternalLogin() gin.HandlerFunc {
	return server.Handler("AuthHandler.HandleInternalLogin", http.StatusOK, func(ctx *gin.Context) (any, error) {
		request, err := server.BindJSON[dto.InternalLoginRequest](ctx)
		if err != nil {
			return nil, err
		}

		tokenResp, err := ah.authService.InternalLogin(ctx.Request.Context(), request)
		if err != nil {
			return nil, err
		}

		ah.setTokenCookies(ctx, tokenResp)
		return map[string]string{"message": "ok"}, nil
	})
}

// HandleOAuth2Login godoc
// @Summary      Initiate OAuth2 login
// @Tags         auth
// @Param        provider path string true "OAuth provider (e.g. google)"
// @Success      307
// @Router       /auth/{provider} [get]
func (ah *AuthHandler) HandleOAuth2Login() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		c, span := otel.Tracer.Start(ctx.Request.Context(), "AuthHandler.HandleOAuth2Login")
		defer span.End()
		ctx.Request = ctx.Request.WithContext(c)

		provider, err := ah.getProvider(ctx)
		if err != nil {
			_ = ctx.Error(ungerr.BadRequestError("missing oauth provider"))
			return
		}

		url, err := ah.oAuthService.GetOAuthURL(c, provider)
		if err != nil {
			_ = ctx.Error(err)
			return
		}

		ctx.Redirect(http.StatusTemporaryRedirect, url)
	}
}

// HandleOAuth2Callback godoc
// @Summary      Handle OAuth2 provider callback
// @Tags         auth
// @Produce      json
// @Param        provider path  string true "OAuth provider (e.g. google)"
// @Param        code     query string true "Authorization code from provider"
// @Param        state    query string true "State token"
// @Success      200  {object}  response.JSONResponse[map[string]string]
// @Failure      400  {object}  map[string]any
// @Router       /auth/{provider}/callback [get]
func (ah *AuthHandler) HandleOAuth2Callback() gin.HandlerFunc {
	return server.Handler("AuthHandler.HandleOAuth2Callback", http.StatusOK, func(ctx *gin.Context) (any, error) {
		provider, err := ah.getProvider(ctx)
		if err != nil {
			return nil, err
		}

		request := dto.OAuthCallbackData{
			Provider: provider,
			Code:     ctx.Query("code"),
			State:    ctx.Query("state"),
		}

		tokenResp, err := ah.oAuthService.HandleOAuthCallback(ctx.Request.Context(), request)
		if err != nil {
			return nil, err
		}

		ah.setTokenCookies(ctx, tokenResp)
		return map[string]string{"message": "ok"}, nil
	})
}

// HandleVerifyRegistration godoc
// @Summary      Verify email registration
// @Tags         auth
// @Produce      json
// @Param        token query string true "Verification token"
// @Success      200  {object}  response.JSONResponse[map[string]string]
// @Failure      400  {object}  map[string]any
// @Router       /auth/verify-registration [get]
func (ah *AuthHandler) HandleVerifyRegistration() gin.HandlerFunc {
	return server.Handler("AuthHandler.HandleVerifyRegistration", http.StatusOK, func(ctx *gin.Context) (any, error) {
		token := ctx.Query("token")
		if token == "" {
			return nil, ungerr.BadRequestError("missing token")
		}

		tokenResp, err := ah.authService.VerifyRegistration(ctx.Request.Context(), token)
		if err != nil {
			return nil, err
		}

		ah.setTokenCookies(ctx, tokenResp)
		return map[string]string{"message": "ok"}, nil
	})
}

// HandleSendPasswordReset godoc
// @Summary      Send password reset email
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body dto.SendPasswordResetRequest true "Email payload"
// @Success      201  {object}  map[string]any
// @Failure      400  {object}  map[string]any
// @Router       /auth/password-reset [post]
func (ah *AuthHandler) HandleSendPasswordReset() gin.HandlerFunc {
	return server.Handler("AuthHandler.HandleSendPasswordReset", http.StatusCreated, func(ctx *gin.Context) (any, error) {
		request, err := server.BindJSON[dto.SendPasswordResetRequest](ctx)
		if err != nil {
			return nil, err
		}

		return nil, ah.authService.SendPasswordReset(ctx.Request.Context(), request.Email)
	})
}

// HandleResetPassword godoc
// @Summary      Reset password using token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body dto.ResetPasswordRequest true "Reset password payload"
// @Success      200  {object}  response.JSONResponse[map[string]string]
// @Failure      400  {object}  map[string]any
// @Router       /auth/reset-password [patch]
func (ah *AuthHandler) HandleResetPassword() gin.HandlerFunc {
	return server.Handler("AuthHandler.HandleResetPassword", http.StatusOK, func(ctx *gin.Context) (any, error) {
		request, err := server.BindJSON[dto.ResetPasswordRequest](ctx)
		if err != nil {
			return nil, err
		}

		tokenResp, err := ah.authService.ResetPassword(ctx.Request.Context(), request.Token, request.Password)
		if err != nil {
			return nil, err
		}

		ah.setTokenCookies(ctx, tokenResp)
		return map[string]string{"message": "ok"}, nil
	})
}

// HandleRefreshToken godoc
// @Summary      Refresh access token
// @Tags         auth
// @Produce      json
// @Success      200  {object}  response.JSONResponse[map[string]string]
// @Failure      401  {object}  map[string]any
// @Router       /auth/refresh [put]
func (ah *AuthHandler) HandleRefreshToken() gin.HandlerFunc {
	return server.Handler("AuthHandler.HandleRefreshToken", http.StatusOK, func(ctx *gin.Context) (any, error) {
		refreshToken, err := ctx.Cookie(cookie.RefreshTokenName)
		if err != nil {
			return nil, ungerr.UnauthorizedError("missing refresh token")
		}

		tokenResp, err := ah.sessionService.RefreshToken(ctx.Request.Context(), dto.RefreshTokenRequest{RefreshToken: refreshToken})
		if err != nil {
			return nil, err
		}

		ah.setTokenCookies(ctx, tokenResp)
		return map[string]string{"message": "ok"}, nil
	})
}

func (ah *AuthHandler) getProvider(ctx *gin.Context) (string, error) {
	provider := ctx.Param(appconstant.ContextProvider.String())
	if provider == "" {
		return "", ungerr.BadRequestError("missing oauth provider")
	}
	return provider, nil
}

// HandleLogout godoc
// @Summary      Logout current session
// @Tags         auth
// @Security     BearerAuth
// @Success      204
// @Failure      401  {object}  map[string]any
// @Router       /auth/logout [delete]
func (ah *AuthHandler) HandleLogout() gin.HandlerFunc {
	return server.Handler("AuthHandler.HandleLogout", http.StatusNoContent, func(ctx *gin.Context) (any, error) {
		sessionID, err := server.GetFromContext[uuid.UUID](ctx, appconstant.ContextSessionID.String())
		if err != nil {
			return nil, err
		}

		if err := ah.authService.Logout(ctx.Request.Context(), sessionID); err != nil {
			return nil, err
		}

		cookie.ClearTokens(ctx, ah.cookieCfg)
		return nil, nil
	})
}
