package store

import (
	"context"
	"testing"
	"time"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/assert"
)

type mockKV struct {
	entries map[string][]byte
	jetstream.KeyValue
}

func newMockKV() *mockKV {
	return &mockKV{entries: make(map[string][]byte)}
}

func (m *mockKV) Create(ctx context.Context, key string, value []byte, opts ...jetstream.KVCreateOpt) (uint64, error) {
	if _, ok := m.entries[key]; ok {
		return 0, jetstream.ErrKeyExists
	}
	m.entries[key] = value
	return 1, nil
}

func (m *mockKV) Get(ctx context.Context, key string) (jetstream.KeyValueEntry, error) {
	v, ok := m.entries[key]
	if !ok {
		return nil, jetstream.ErrKeyNotFound
	}
	return &mockEntry{revision: 1, value: v}, nil
}

func (m *mockKV) Delete(ctx context.Context, key string, opts ...jetstream.KVDeleteOpt) error {
	delete(m.entries, key)
	return nil
}

type mockEntry struct {
	jetstream.KeyValueEntry
	revision uint64
	value    []byte
}

func (e *mockEntry) Revision() uint64 { return e.revision }
func (e *mockEntry) Value() []byte    { return e.value }

func TestNATSKVStateStore_Store(t *testing.T) {
	kv := newMockKV()
	s := NewNATSKVStateStore(kv)

	err := s.Store(context.Background(), "abc123", "session-data", 5*time.Minute)
	assert.NoError(t, err)

	assert.Contains(t, kv.entries, "state.abc123")
}

func TestNATSKVStateStore_Store_Duplicate(t *testing.T) {
	kv := newMockKV()
	s := NewNATSKVStateStore(kv)

	_ = s.Store(context.Background(), "abc123", "session-data", 5*time.Minute)
	err := s.Store(context.Background(), "abc123", "session-data", 5*time.Minute)
	assert.Error(t, err)
}

func TestNATSKVStateStore_VerifyAndDelete(t *testing.T) {
	kv := newMockKV()
	s := NewNATSKVStateStore(kv)

	_ = s.Store(context.Background(), "abc123", "session-data", 5*time.Minute)

	value, err := s.VerifyAndDelete(context.Background(), "abc123")
	assert.NoError(t, err)
	assert.Equal(t, "session-data", value)

	assert.NotContains(t, kv.entries, "state.abc123")
}

func TestNATSKVStateStore_VerifyAndDelete_NotFound(t *testing.T) {
	kv := newMockKV()
	s := NewNATSKVStateStore(kv)

	_, err := s.VerifyAndDelete(context.Background(), "nonexistent")
	assert.Error(t, err)
}

func TestNATSKVStateStore_Shutdown(t *testing.T) {
	kv := newMockKV()
	s := NewNATSKVStateStore(kv)

	assert.NoError(t, s.Shutdown())
}
