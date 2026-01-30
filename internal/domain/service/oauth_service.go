package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"time"

	"github.com/itsLeonB/cashback/internal/core/config"
	"github.com/itsLeonB/cashback/internal/core/service/store"
	"github.com/itsLeonB/cashback/internal/domain/dto"
	"github.com/itsLeonB/cashback/internal/domain/entity/users"
	"github.com/itsLeonB/cashback/internal/domain/mapper"
	"github.com/itsLeonB/cashback/internal/domain/service/oauth"
	"github.com/itsLeonB/go-crud"
	"github.com/itsLeonB/sekure"
	"github.com/itsLeonB/ungerr"
)

type oauthServiceImpl struct {
	jwtService       sekure.JWTService
	transactor       crud.Transactor
	oauthProviders   map[string]oauth.ProviderService
	oauthAccountRepo crud.Repository[users.OAuthAccount]
	stateStore       store.StateStore
	userSvc          UserService
}

func NewOAuthService(
	transactor crud.Transactor,
	oauthAccountRepo crud.Repository[users.OAuthAccount],
	stateStore store.StateStore,
	userSvc UserService,
	httpClient *http.Client,
	jwtSvc sekure.JWTService,
) OAuthService {
	return &oauthServiceImpl{
		jwtSvc,
		transactor,
		oauth.NewOAuthProviderServices(config.Global.OAuthProviders, httpClient),
		oauthAccountRepo,
		stateStore,
		userSvc,
	}
}

func (as *oauthServiceImpl) GetOAuthURL(ctx context.Context, provider string) (string, error) {
	oauthProvider, ok := as.oauthProviders[provider]
	if !ok {
		return "", ungerr.Unknownf("unsupported oauth provider: %s", provider)
	}

	state, err := as.generateState()
	if err != nil {
		return "", err
	}

	url, err := oauthProvider.GetAuthCodeURL(ctx, state)
	if err != nil {
		return "", err
	}

	if err = as.stateStore.Store(ctx, state, 5*time.Minute); err != nil {
		return "", err
	}

	return url, nil
}

func (as *oauthServiceImpl) HandleOAuthCallback(ctx context.Context, data dto.OAuthCallbackData) (dto.LoginResponse, error) {
	var response dto.LoginResponse
	err := as.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		oauthProvider, ok := as.oauthProviders[data.Provider]
		if !ok {
			return ungerr.Unknownf("unsupported oauth provider: %s", data.Provider)
		}

		if err := as.stateStore.VerifyAndDelete(ctx, data.State); err != nil {
			return err
		}

		userInfo, err := oauthProvider.HandleCallback(ctx, data.Code)
		if err != nil {
			return err
		}

		user, err := as.getOrCreateUser(ctx, userInfo)
		if err != nil {
			return err
		}

		if !user.IsVerified() {
			if _, err = as.userSvc.Verify(ctx, user.ID, user.Email, userInfo.Name, userInfo.Avatar); err != nil {
				return err
			}
		}

		response, err = as.CreateLoginResponse(user, users.Session{})
		return err
	})

	return response, err
}

func (as *oauthServiceImpl) getOrCreateUser(ctx context.Context, userInfo oauth.UserInfo) (users.User, error) {
	existingOAuth, err := as.findOAuthAccount(ctx, userInfo.Provider, userInfo.ProviderID)
	if err != nil {
		return users.User{}, err
	}
	if !existingOAuth.IsZero() {
		return existingOAuth.User, nil
	}
	return as.createNewUserOAuth(ctx, userInfo)
}

func (as *oauthServiceImpl) createNewUserOAuth(ctx context.Context, userInfo oauth.UserInfo) (users.User, error) {
	user, err := as.userSvc.FindByEmail(ctx, userInfo.Email)
	if err != nil {
		return users.User{}, err
	}
	if user.IsZero() {
		// New user
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

	if !as.oauthProviders[userInfo.Provider].IsTrusted() {
		return users.User{}, ungerr.Unknown("provider temporarily disabled")
	}

	// New oauth method
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

func (as *oauthServiceImpl) CreateLoginResponse(user users.User, session users.Session) (dto.LoginResponse, error) {
	authData := mapper.UserToAuthData(user, session)

	token, err := as.jwtService.CreateToken(authData)
	if err != nil {
		return dto.LoginResponse{}, err
	}

	return dto.NewBearerTokenResp(token), nil
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
