package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"github.com/itsLeonB/cashback/internal/core/logger"
	"github.com/itsLeonB/cashback/internal/core/otel"
	"github.com/itsLeonB/cashback/internal/domain/dto"
	"github.com/itsLeonB/cashback/internal/domain/service/auth"
	"github.com/itsLeonB/cashback/internal/domain/service/oauth"
	"github.com/itsLeonB/ungerr"
)

type oauthServiceImpl struct {
	transactor    auth.Transactor
	providerSvc   oauth.ProviderService
	oauthAccounts auth.OAuthAccountStore
	stateStore    auth.StateStore
	users         auth.UserStore
	sessionSvc    SessionService
	hooks         AuthHooks
}

func NewOAuthService(
	transactor auth.Transactor,
	providerSvc oauth.ProviderService,
	oauthAccounts auth.OAuthAccountStore,
	stateStore auth.StateStore,
	users auth.UserStore,
	sessionSvc SessionService,
	hooks AuthHooks,
) OAuthService {
	return &oauthServiceImpl{
		transactor:    transactor,
		providerSvc:   providerSvc,
		oauthAccounts: oauthAccounts,
		stateStore:    stateStore,
		users:         users,
		sessionSvc:    sessionSvc,
		hooks:         hooks,
	}
}

func (as *oauthServiceImpl) GetOAuthURL(ctx context.Context, provider string) (string, error) {
	ctx, span := otel.Tracer.Start(ctx, "OAuthService.GetOAuthURL")
	defer span.End()

	state, err := as.generateState()
	if err != nil {
		return "", err
	}

	url, sessionStr, err := as.providerSvc.GetAuthCodeURL(provider, state)
	if err != nil {
		return "", err
	}

	if err = as.stateStore.Store(ctx, state, sessionStr, 5*time.Minute); err != nil {
		return "", err
	}

	return url, nil
}

func (as *oauthServiceImpl) HandleOAuthCallback(ctx context.Context, data dto.OAuthCallbackData) (dto.TokenResponse, error) {
	ctx, span := otel.Tracer.Start(ctx, "OAuthService.HandleOAuthCallback")
	defer span.End()

	// Preserve the parent context for hooks that run outside the transaction.
	parentCtx := ctx

	var (
		response dto.TokenResponse
		user     auth.User
		isNew    bool
	)
	err := as.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		sessionStr, err := as.stateStore.VerifyAndDelete(ctx, data.State)
		if err != nil {
			return err
		}

		userInfo, err := as.providerSvc.HandleCallback(ctx, data.Provider, data.Code, sessionStr)
		if err != nil {
			return err
		}

		user, isNew, err = as.getOrCreateUser(ctx, userInfo)
		if err != nil {
			return err
		}

		if !user.Verified {
			var err error
			user, err = as.users.SetVerified(ctx, user.ID, userInfo.Name, userInfo.Avatar)
			if err != nil {
				return err
			}
		}

		response, err = as.sessionSvc.CreateTokenAndSession(ctx, user)
		return err
	})
	if err != nil {
		return dto.TokenResponse{}, err
	}

	if hookErr := as.hooks.CallAfterOAuthLogin(parentCtx, user.ID, data.Provider, isNew); hookErr != nil {
		logger.Error(hookErr)
	}

	return response, nil
}

func (as *oauthServiceImpl) getOrCreateUser(ctx context.Context, userInfo oauth.UserInfo) (auth.User, bool, error) {
	existingOAuth, err := as.oauthAccounts.FindByProvider(ctx, userInfo.Provider, userInfo.ProviderID)
	if err != nil && !errors.Is(err, auth.ErrUserNotFound) {
		return auth.User{}, false, err
	}
	if !existingOAuth.IsZero() {
		found, err := as.users.FindByEmail(ctx, existingOAuth.Email)
		if err != nil {
			return auth.User{}, false, err
		}
		return found, false, nil
	}
	user, err := as.createNewUserOAuth(ctx, userInfo)
	return user, true, err
}

func (as *oauthServiceImpl) createNewUserOAuth(ctx context.Context, userInfo oauth.UserInfo) (auth.User, error) {
	user, err := as.users.FindByEmail(ctx, userInfo.Email)
	if err != nil && !errors.Is(err, auth.ErrUserNotFound) {
		return auth.User{}, err
	}
	if user.IsZero() {
		user, err = as.users.CreateOAuth(ctx, userInfo.Email, userInfo.Name, userInfo.Avatar)
		if err != nil {
			return auth.User{}, err
		}
	}

	trusted, err := as.providerSvc.IsTrusted(userInfo.Provider)
	if err != nil {
		return auth.User{}, err
	}
	if !trusted {
		return auth.User{}, ungerr.Unknown("provider temporarily disabled")
	}

	if err = as.oauthAccounts.Link(ctx, user.ID, userInfo.Provider, userInfo.ProviderID, userInfo.Email); err != nil {
		return auth.User{}, err
	}

	return user, nil
}

func (as *oauthServiceImpl) generateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", ungerr.Wrap(err, "error generating random string")
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
