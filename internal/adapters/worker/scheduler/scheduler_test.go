package scheduler

import (
	"context"
	"os"
	"testing"

	"github.com/itsLeonB/cashback/internal/core/logger"
	"github.com/robfig/cron/v3"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	logger.Init("test")
	os.Exit(m.Run())
}

func TestJobWrapper_PanicRecovery(t *testing.T) {
	s := &Scheduler{cron: cron.New()}
	wrapped := s.jobWrapper("test-panic-job", func(ctx context.Context) error {
		panic("test panic")
	})

	assert.NotPanics(t, wrapped)
}
