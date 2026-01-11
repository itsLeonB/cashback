package store

import (
	"context"
	"time"

	"github.com/itsLeonB/cashback/internal/adapters/core/service/store"
	"github.com/itsLeonB/cashback/internal/core/config"
	"github.com/itsLeonB/ungerr"
)

type StateStore interface {
	Store(ctx context.Context, state string, expiry time.Duration) error
	VerifyAndDelete(ctx context.Context, state string) error
	Shutdown() error
}

func NewStateStore() (StateStore, error) {
	switch config.Global.StateStore {
	case "inmemory":
		return store.NewInMemoryStateStore(), nil
	case "valkey":
		return store.NewValkeyStateStore(config.Global.Valkey)
	default:
		return nil, ungerr.Unknownf("unimplemented state store: %s", config.Global.StateStore)
	}
}
