package subscriber

import (
	"context"
	"os"
	"testing"

	"github.com/itsLeonB/cashback/internal/core/logger"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	logger.Init("test")
	os.Exit(m.Run())
}

type mockMsg struct {
	jetstream.Msg
	nakCalled bool
}

func (m *mockMsg) Data() []byte {
	return []byte(`{"type":"test"}`)
}

func (m *mockMsg) Nak() error {
	m.nakCalled = true
	return nil
}

type testTask struct {
	TaskType string `json:"type"`
}

func (t testTask) Type() string {
	return t.TaskType
}

func TestWithLogging_PanicRecovery(t *testing.T) {
	msg := &mockMsg{}
	handler := withLogging[testTask]("test-task", func(ctx context.Context, task testTask) error {
		panic("test panic")
	})

	assert.NotPanics(t, func() { handler(msg) })
	assert.True(t, msg.nakCalled)
}
