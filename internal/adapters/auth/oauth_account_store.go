package authadapter

import (
	"context"

	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/domain/entity/users"
	"github.com/itsLeonB/cashback/internal/domain/service/auth"
	"github.com/itsLeonB/go-crud"
)

type oauthAccountStoreAdapter struct {
	repo crud.Repository[users.OAuthAccount]
}

func NewOAuthAccountStore(repo crud.Repository[users.OAuthAccount]) auth.OAuthAccountStore {
	return &oauthAccountStoreAdapter{repo}
}

func (a *oauthAccountStoreAdapter) FindByProvider(ctx context.Context, provider, providerID string) (auth.OAuthAccount, error) {
	spec := crud.Specification[users.OAuthAccount]{}
	spec.Model.Provider = provider
	spec.Model.ProviderID = providerID
	spec.PreloadRelations = []string{"User"}
	account, err := a.repo.FindFirst(ctx, spec)
	if err != nil {
		return auth.OAuthAccount{}, err
	}
	if account.IsZero() {
		return auth.OAuthAccount{}, auth.ErrUserNotFound
	}
	return auth.OAuthAccount{
		UserID:     account.UserID.String(),
		Provider:   account.Provider,
		ProviderID: account.ProviderID,
		Email:      account.Email,
	}, nil
}

func (a *oauthAccountStoreAdapter) Link(ctx context.Context, userID, provider, providerID, email string) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return err
	}
	_, err = a.repo.Insert(ctx, users.OAuthAccount{
		UserID:     uid,
		Provider:   provider,
		ProviderID: providerID,
		Email:      email,
	})
	return err
}
