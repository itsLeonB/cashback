package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"time"

	"github.com/itsLeonB/cashback/internal/core/logger"
	"github.com/itsLeonB/cashback/internal/core/otel"
	"github.com/itsLeonB/cashback/internal/core/service/store"
	"github.com/itsLeonB/cashback/internal/domain/dto"
	"github.com/itsLeonB/cashback/internal/domain/entity/users"
	"github.com/itsLeonB/cashback/internal/domain/service/oauth"
	"github.com/itsLeonB/go-crud"
	"github.com/itsLeonB/ungerr"
)

type oauthServiceImpl struct {
	transactor       crud.Transactor
	providerSvc      oauth.ProviderService
	oauthAccountRepo crud.Repository[users.OAuthAccount]
	stateStore       store.StateStore
	userSvc          UserService
	sessionSvc       SessionService
	hooks            AuthHooks
}

func NewOAuthService(
	transactor crud.Transactor,
	providerSvc oauth.ProviderService,
	oauthAccountRepo crud.Repository[users.OAuthAccount],
	stateStore store.StateStore,
	userSvc UserService,
	sessionSvc SessionService,
	hooks AuthHooks,
) OAuthService {
	return &oauthServiceImpl{
		transactor:       transactor,
		providerSvc:      providerSvc,
		oauthAccountRepo: oauthAccountRepo,
		stateStore:       stateStore,
		userSvc:          userSvc,
		sessionSvc:       sessionSvc,
		hooks:            hooks,
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
		user     users.User
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

		if !user.IsVerified() {
			if _, err = as.userSvc.Verify(ctx, user.ID, user.Email, userInfo.Name, userInfo.Avatar); err != nil {
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

func (as *oauthServiceImpl) getOrCreateUser(ctx context.Context, userInfo oauth.UserInfo) (users.User, bool, error) {
	existingOAuth, err := as.findOAuthAccount(ctx, userInfo.Provider, userInfo.ProviderID)
	if err != nil {
		return users.User{}, false, err
	}
	if !existingOAuth.IsZero() {
		return existingOAuth.User, false, nil
	}
	user, err := as.createNewUserOAuth(ctx, userInfo)
	return user, true, err
}

func (as *oauthServiceImpl) createNewUserOAuth(ctx context.Context, userInfo oauth.UserInfo) (users.User, error) {
	user, err := as.userSvc.FindByEmail(ctx, userInfo.Email)
	if err != nil {
		return users.User{}, err
	}
	if user.IsZero() {
		newUser := dto.NewUserRequest{
			Email:     userInfo.Email,
			Name:      userInfo.Name,
			Avatar:    userInfo.Avatar,
			VerifyNow: true,
		}
		user, err = as.userSvc.CreateNew(ctx, newUser)
		if err != nil {
			return users.User{}, err
		}
	}

	trusted, err := as.providerSvc.IsTrusted(userInfo.Provider)
	if err != nil {
		return users.User{}, err
	}
	if !trusted {
		return users.User{}, ungerr.Unknown("provider temporarily disabled")
	}

	newOAuthAccount := users.OAuthAccount{
		UserID:     user.ID,
		Provider:   userInfo.Provider,
		ProviderID: userInfo.ProviderID,
		Email:      userInfo.Email,
	}

	if _, err = as.oauthAccountRepo.Insert(ctx, newOAuthAccount); err != nil {
		return users.User{}, err
	}

	return user, nil
}

func (as *oauthServiceImpl) findOAuthAccount(ctx context.Context, provider, providerID string) (users.OAuthAccount, error) {
	oauthSpec := crud.Specification[users.OAuthAccount]{}
	oauthSpec.Model.Provider = provider
	oauthSpec.Model.ProviderID = providerID
	oauthSpec.PreloadRelations = []string{"User"}
	return as.oauthAccountRepo.FindFirst(ctx, oauthSpec)
}

func (as *oauthServiceImpl) generateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", ungerr.Wrap(err, "error generating random string")
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
