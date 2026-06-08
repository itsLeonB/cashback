package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/itsLeonB/ungerr"
	"github.com/nats-io/nats.go/jetstream"
)

type natsKVStateStore struct {
	kv jetstream.KeyValue
}

func NewNATSKVStateStore(kv jetstream.KeyValue) *natsKVStateStore {
	return &natsKVStateStore{kv: kv}
}

func (s *natsKVStateStore) Store(ctx context.Context, state string, expiry time.Duration) error {
	_, err := s.kv.Create(ctx, s.constructKey(state), []byte(state), jetstream.KeyTTL(expiry))
	if err != nil {
		return ungerr.Wrap(err, "error storing state in NATS KV")
	}
	return nil
}

func (s *natsKVStateStore) VerifyAndDelete(ctx context.Context, state string) error {
	key := s.constructKey(state)
	entry, err := s.kv.Get(ctx, key)
	if err != nil {
		if errors.Is(err, jetstream.ErrKeyNotFound) {
			return ungerr.BadRequestError("invalid state")
		}
		return ungerr.Wrap(err, "error verifying state in NATS KV")
	}

	if err := s.kv.Delete(ctx, key, jetstream.LastRevision(entry.Revision())); err != nil {
		return ungerr.BadRequestError("invalid state")
	}
	return nil
}

func (s *natsKVStateStore) Shutdown() error {
	return nil
}

func (s *natsKVStateStore) constructKey(state string) string {
	return fmt.Sprintf("state.%s", state)
}
