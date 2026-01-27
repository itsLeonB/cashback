package subscriber

import (
	"context"
	"encoding/json"

	"github.com/hibiken/asynq"
	"github.com/itsLeonB/cashback/internal/core/logger"
	"github.com/itsLeonB/cashback/internal/core/service/queue"
	"github.com/itsLeonB/ungerr"
)

func withLogging[T queue.TaskMessage](taskType string, handler func(context.Context, T) error) asynq.Handler {
	return asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) error {
		logger.Infof("received new task %s", taskType)

		var msg T
		if err := json.Unmarshal(t.Payload(), &msg); err != nil {
			logger.Errorf("error processing %s task: %v", taskType, err)
			return ungerr.Wrapf(err, "error unmarshaling payload to: %T", msg)
		}

		if err := handler(ctx, msg); err != nil {
			logger.Errorf("error processing %s task: %v", taskType, err)
			return err
		}

		logger.Infof("success processing %s task", taskType)
		return nil
	})
}
