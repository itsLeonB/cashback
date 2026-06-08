package store

import (
	"context"
	"time"

	"github.com/itsLeonB/cashback/internal/adapters/core/service/store"
)

type StateStore interface {
	Store(ctx context.Context, state string, expiry time.Duration) error
	VerifyAndDelete(ctx context.Context, state string) error
	Shutdown() error
}

func NewStateStore() StateStore {
	return store.NewInMemoryStateStore()
}
