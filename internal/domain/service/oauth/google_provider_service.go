package oauth

import (
	"context"
	"io"
	"net/http"

	"github.com/itsLeonB/cashback/internal/core/config"
	"github.com/itsLeonB/cashback/internal/core/logger"
	"github.com/itsLeonB/ezutil/v2"
	"github.com/itsLeonB/ungerr"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type googleUserInfo struct {
	ID      string `json:"id"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

type googleProviderService struct {
	userInfoURL string
	cfg         *oauth2.Config
	httpClient  *http.Client
}

func newGoogleProviderService(
	oauthConfig config.OAuthProvider,
	httpClient *http.Client,
) ProviderService {
	return &googleProviderService{
		userInfoURL: "https://www.googleapis.com/oauth2/v2/userinfo",
		cfg: &oauth2.Config{
			ClientID:     oauthConfig.ClientID,
			ClientSecret: oauthConfig.ClientSecret,
			RedirectURL:  oauthConfig.RedirectUrl,
			Endpoint:     google.Endpoint,
			Scopes: []string{
				"https://www.googleapis.com/auth/userinfo.email",
				"https://www.googleapis.com/auth/userinfo.profile",
			},
		},
		httpClient: httpClient,
	}
}

func (*googleProviderService) IsTrusted() bool {
	return true
}

func (gps *googleProviderService) GetAuthCodeURL(ctx context.Context, state string) (string, error) {
	url := gps.cfg.AuthCodeURL(state, oauth2.AccessTypeOffline)
	if url == "" {
		return "", ungerr.Unknownf("OAuth2 google provider returns empty string for auth code URL")
	}
	return url, nil
}

func (gps *googleProviderService) HandleCallback(ctx context.Context, code string) (UserInfo, error) {
	token, err := gps.cfg.Exchange(ctx, code)
	if err != nil {
		return UserInfo{}, ungerr.Wrap(err, "error exchange OAuth2 token at callback")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, gps.userInfoURL, nil)
	if err != nil {
		return UserInfo{}, ungerr.Wrap(err, "error creating new HTTP request")
	}
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)

	resp, err := gps.httpClient.Do(req)
	if err != nil {
		return UserInfo{}, ungerr.Wrap(err, "error making HTTP request")
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logger.Error(ungerr.Wrap(err, "error closing HTTP response body"))
		}
	}()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return UserInfo{}, ungerr.Wrap(err, "error reading response body")
	}
	if resp.StatusCode != http.StatusOK {
		return UserInfo{}, ungerr.Unknownf("error getting user info: %s", string(body))
	}

	userInfo, err := ezutil.Unmarshal[googleUserInfo](body)
	if err != nil {
		return UserInfo{}, err
	}

	return UserInfo{
		Provider:    "google",
		ProviderID:  userInfo.ID,
		Email:       userInfo.Email,
		Name:        userInfo.Name,
		Avatar:      userInfo.Picture,
		AccessToken: token.AccessToken,
	}, nil
}
