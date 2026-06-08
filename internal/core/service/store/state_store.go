package store

import (
	"context"
	"time"

	"github.com/itsLeonB/cashback/internal/adapters/core/service/store"
	"github.com/itsLeonB/cashback/internal/core/config"
	"github.com/itsLeonB/ungerr"
	"github.com/nats-io/nats.go/jetstream"
)

type StateStore interface {
	Store(ctx context.Context, state string, expiry time.Duration) error
	VerifyAndDelete(ctx context.Context, state string) error
	Shutdown() error
}

func NewStateStore(js jetstream.JetStream) (StateStore, error) {
	switch config.Global.StateStore {
	case "nats":
		ctx := context.Background()
		kv, err := js.CreateOrUpdateKeyValue(ctx, jetstream.KeyValueConfig{
			Bucket:         config.Global.StateStoreBucket,
			History:        1,
			LimitMarkerTTL: 10 * time.Minute,
		})
		if err != nil {
			return nil, ungerr.Wrap(err, "error creating NATS KV state store bucket")
		}
		return store.NewNATSKVStateStore(kv), nil
	case "inmemory":
		return store.NewInMemoryStateStore(), nil
	default:
		return nil, ungerr.Unknownf("unsupported AUTH_STATE_STORE value: %q", config.Global.StateStore)
	}
}
