package store

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/itsLeonB/ungerr"
)

type inMemoryStateStore struct {
	data *sync.Map
}

func NewInMemoryStateStore() *inMemoryStateStore {
	return &inMemoryStateStore{new(sync.Map)}
}

func (vss *inMemoryStateStore) Store(ctx context.Context, state string, expiry time.Duration) error {
	vss.data.Store(vss.constructKey(state), state)
	return nil
}

func (vss *inMemoryStateStore) VerifyAndDelete(ctx context.Context, state string) error {
	if _, loaded := vss.data.LoadAndDelete(vss.constructKey(state)); !loaded {
		return ungerr.BadRequestError("invalid state")
	}

	return nil
}

func (vss *inMemoryStateStore) Shutdown() error {
	return nil
}

func (vss *inMemoryStateStore) constructKey(state string) string {
	return fmt.Sprintf("state:%s", state)
}
