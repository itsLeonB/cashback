package authadapter

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/domain/entity/users"
	"github.com/itsLeonB/cashback/internal/domain/service/auth"
	"github.com/itsLeonB/go-crud"
)

type resetTokenStoreAdapter struct {
	repo crud.Repository[users.PasswordResetToken]
}

func NewResetTokenStore(repo crud.Repository[users.PasswordResetToken]) auth.ResetTokenStore {
	return &resetTokenStoreAdapter{repo}
}

func (a *resetTokenStoreAdapter) Create(ctx context.Context, userID, selector, verifierHash string, expiresAt time.Time) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return err
	}

	_, err = a.repo.Insert(ctx, users.PasswordResetToken{
		UserID:       uid,
		Selector:     selector,
		VerifierHash: verifierHash,
		ExpiresAt:    expiresAt,
	})
	return err
}

func (a *resetTokenStoreAdapter) FindBySelector(ctx context.Context, selector string) (auth.ResetToken, error) {
	spec := crud.Specification[users.PasswordResetToken]{}
	spec.Model.Selector = selector
	rt, err := a.repo.FindFirst(ctx, spec)
	if err != nil {
		return auth.ResetToken{}, err
	}
	if rt.IsZero() {
		return auth.ResetToken{}, auth.ErrTokenNotFound
	}
	return auth.ResetToken{
		UserID:       rt.UserID.String(),
		Selector:     rt.Selector,
		VerifierHash: rt.VerifierHash,
		ExpiresAt:    rt.ExpiresAt,
	}, nil
}

func (a *resetTokenStoreAdapter) DeleteByUser(ctx context.Context, userID string) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return err
	}

	spec := crud.Specification[users.PasswordResetToken]{}
	spec.Model.UserID = uid
	tokens, err := a.repo.FindAll(ctx, spec)
	if err != nil {
		return err
	}
	if len(tokens) == 0 {
		return nil
	}
	return a.repo.DeleteMany(ctx, tokens)
}
