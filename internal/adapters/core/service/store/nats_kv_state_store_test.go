package store

import (
	"context"
	"testing"
	"time"

	"github.com/nats-io/nats.go/jetstream"
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
	if _, ok := m.entries[key]; !ok {
		return nil, jetstream.ErrKeyNotFound
	}
	return &mockEntry{revision: 1}, nil
}

func (m *mockKV) Delete(ctx context.Context, key string, opts ...jetstream.KVDeleteOpt) error {
	delete(m.entries, key)
	return nil
}

type mockEntry struct {
	jetstream.KeyValueEntry
	revision uint64
}

func (e *mockEntry) Revision() uint64 { return e.revision }

func TestNATSKVStateStore_Store(t *testing.T) {
	kv := newMockKV()
	s := NewNATSKVStateStore(kv)

	err := s.Store(context.Background(), "abc123", 5*time.Minute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, ok := kv.entries["state.abc123"]; !ok {
		t.Fatal("expected key to be stored")
	}
}

func TestNATSKVStateStore_Store_Duplicate(t *testing.T) {
	kv := newMockKV()
	s := NewNATSKVStateStore(kv)

	_ = s.Store(context.Background(), "abc123", 5*time.Minute)
	err := s.Store(context.Background(), "abc123", 5*time.Minute)
	if err == nil {
		t.Fatal("expected error on duplicate store")
	}
}

func TestNATSKVStateStore_VerifyAndDelete(t *testing.T) {
	kv := newMockKV()
	s := NewNATSKVStateStore(kv)

	_ = s.Store(context.Background(), "abc123", 5*time.Minute)

	err := s.VerifyAndDelete(context.Background(), "abc123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, ok := kv.entries["state.abc123"]; ok {
		t.Fatal("expected key to be deleted")
	}
}

func TestNATSKVStateStore_VerifyAndDelete_NotFound(t *testing.T) {
	kv := newMockKV()
	s := NewNATSKVStateStore(kv)

	err := s.VerifyAndDelete(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent state")
	}
}

func TestNATSKVStateStore_Shutdown(t *testing.T) {
	kv := newMockKV()
	s := NewNATSKVStateStore(kv)

	if err := s.Shutdown(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
