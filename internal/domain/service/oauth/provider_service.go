package oauth

import (
	"context"
	"net/url"

	"github.com/itsLeonB/cashback/internal/core/config"
	"github.com/itsLeonB/ungerr"
	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/google"
)

type ProviderService interface {
	IsTrusted(provider string) (bool, error)
	GetAuthCodeURL(provider string, state string) (url string, session string, err error)
	HandleCallback(ctx context.Context, provider string, code string, session string) (UserInfo, error)
}

type providerServiceImpl struct {
	providers map[string]providerEntry
}

type providerEntry struct {
	provider goth.Provider
	trusted  bool
}

func NewProviderService(cfgs config.OAuthProviders) ProviderService {
	return &providerServiceImpl{
		providers: map[string]providerEntry{
			"google": {
				provider: google.New(cfgs.Google.ClientID, cfgs.Google.ClientSecret, cfgs.Google.RedirectUrl, "email", "profile"),
				trusted:  true,
			},
		},
	}
}

func (a *providerServiceImpl) get(provider string) (providerEntry, error) {
	entry, ok := a.providers[provider]
	if !ok {
		return providerEntry{}, ungerr.BadRequestError("unsupported oauth provider: " + provider)
	}
	return entry, nil
}

func (a *providerServiceImpl) IsTrusted(provider string) (bool, error) {
	entry, err := a.get(provider)
	if err != nil {
		return false, err
	}
	return entry.trusted, nil
}

func (a *providerServiceImpl) GetAuthCodeURL(provider string, state string) (string, string, error) {
	entry, err := a.get(provider)
	if err != nil {
		return "", "", err
	}

	session, err := entry.provider.BeginAuth(state)
	if err != nil {
		return "", "", ungerr.Wrap(err, "error beginning oauth auth")
	}
	authURL, err := session.GetAuthURL()
	if err != nil {
		return "", "", ungerr.Wrap(err, "error getting oauth auth URL")
	}
	return authURL, session.Marshal(), nil
}

// TODO: goth's Authorize and FetchUser don't accept context, so these HTTP calls
// won't respect request cancellation or deadlines. This is a goth limitation.
func (a *providerServiceImpl) HandleCallback(ctx context.Context, provider string, code string, sessionStr string) (UserInfo, error) {
	entry, err := a.get(provider)
	if err != nil {
		return UserInfo{}, err
	}

	session, err := entry.provider.UnmarshalSession(sessionStr)
	if err != nil {
		return UserInfo{}, ungerr.Wrap(err, "error unmarshalling oauth session")
	}

	_, err = session.Authorize(entry.provider, url.Values{"code": {code}})
	if err != nil {
		return UserInfo{}, ungerr.Wrap(err, "error authorizing oauth session")
	}

	user, err := entry.provider.FetchUser(session)
	if err != nil {
		return UserInfo{}, ungerr.Wrap(err, "error fetching oauth user")
	}

	return UserInfo{
		Provider:    user.Provider,
		ProviderID:  user.UserID,
		Email:       user.Email,
		Name:        user.Name,
		Avatar:      user.AvatarURL,
		AccessToken: user.AccessToken,
	}, nil
}
