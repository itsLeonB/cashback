package authadapter

import (
	"context"
	"time"

	"github.com/itsLeonB/cashback/internal/core/service/store"
	"github.com/itsLeonB/cashback/internal/domain/service/auth"
)

type stateStoreAdapter struct {
	inner store.StateStore
}

func NewStateStore(inner store.StateStore) auth.StateStore {
	return &stateStoreAdapter{inner}
}

func (a *stateStoreAdapter) Store(ctx context.Context, state, value string, expiry time.Duration) error {
	return a.inner.Store(ctx, state, value, expiry)
}

func (a *stateStoreAdapter) VerifyAndDelete(ctx context.Context, state string) (string, error) {
	return a.inner.VerifyAndDelete(ctx, state)
}
