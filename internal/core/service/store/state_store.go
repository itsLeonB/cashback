package store

import (
	"context"
	"crypto/tls"
	"fmt"
	"sync"
	"time"

	"github.com/itsLeonB/cashback/internal/core/config"
	"github.com/itsLeonB/cashback/internal/core/logger"
	"github.com/itsLeonB/ungerr"
	"github.com/valkey-io/valkey-go"
)

type StateStore interface {
	Store(ctx context.Context, state string, expiry time.Duration) error
	VerifyAndDelete(ctx context.Context, state string) error
	Shutdown() error
}

type valkeyStateStore struct {
	client valkey.Client
	mu     sync.Mutex
}

func NewStateStore(cfg config.Valkey) (StateStore, error) {
	opts := valkey.ClientOption{
		InitAddress: []string{cfg.Addr},
		Password:    cfg.Password,
		SelectDB:    cfg.Db,
	}

	if cfg.EnableTls {
		opts.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
	}

	client, err := valkey.NewClient(opts)
	if err != nil {
		return nil, ungerr.Wrap(err, "error initializing valkey client")
	}

	return &valkeyStateStore{
		client: client,
		mu:     sync.Mutex{},
	}, nil
}

func (vss *valkeyStateStore) Store(ctx context.Context, state string, expiry time.Duration) error {
	key := vss.constructKey(state)

	cmd := vss.client.
		B().
		Set().
		Key(key).
		Value("1").
		ExSeconds(int64(expiry.Seconds())).
		Build()

	if err := vss.client.Do(ctx, cmd).Error(); err != nil {
		return ungerr.Wrap(err, "error storing state in valkey")
	}

	return nil
}

func (vss *valkeyStateStore) VerifyAndDelete(ctx context.Context, state string) error {
	vss.mu.Lock()
	defer vss.mu.Unlock()

	key := vss.constructKey(state)

	getCmd := vss.client.
		B().
		Get().
		Key(key).
		Build()

	if err := vss.client.Do(ctx, getCmd).Error(); err != nil {
		if valkey.IsValkeyNil(err) {
			return ungerr.BadRequestError("invalid state")
		}
		return ungerr.Wrap(err, "failed to get state in valkey")
	}

	delCmd := vss.client.
		B().
		Del().
		Key(key).
		Build()

	if err := vss.client.Do(ctx, delCmd).Error(); err != nil {
		logger.Error(ungerr.Wrap(err, "error deleting key from state store"))
	}

	return nil
}

func (vss *valkeyStateStore) Shutdown() error {
	vss.client.Close()
	return nil
}

func (vss *valkeyStateStore) constructKey(state string) string {
	return fmt.Sprintf("state:%s", state)
}
