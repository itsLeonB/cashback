package oauth

import (
	"context"
	"net/http"

	"github.com/itsLeonB/cashback/internal/core/config"
)

type ProviderService interface {
	IsTrusted() bool
	GetAuthCodeURL(ctx context.Context, state string) (string, error)
	HandleCallback(ctx context.Context, code string) (UserInfo, error)
}

func NewOAuthProviderServices(
	cfgs config.OAuthProviders,
	httpClient *http.Client,
) map[string]ProviderService {
	return map[string]ProviderService{
		"google": newGoogleProviderService(cfgs.Google, httpClient),
	}
}
