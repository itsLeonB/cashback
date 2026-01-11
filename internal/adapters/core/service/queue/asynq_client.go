package queue

import (
	"context"
	"encoding/json"

	"github.com/hibiken/asynq"
	"github.com/itsLeonB/cashback/internal/core/logger"
	"github.com/itsLeonB/cashback/internal/core/service/queue"
	"github.com/itsLeonB/ungerr"
)

type asynqClient struct {
	client *asynq.Client
}

func NewTaskQueue(opts asynq.RedisConnOpt) (*asynqClient, error) {
	client := asynq.NewClient(opts)
	if err := client.Ping(); err != nil {
		return nil, ungerr.Wrap(err, "error pinging asynq client")
	}

	return &asynqClient{client}, nil
}

func (ac *asynqClient) Enqueue(ctx context.Context, message queue.TaskMessage) error {
	payload, err := json.Marshal(message)
	if err != nil {
		return ungerr.Wrap(err, "error marshaling message to JSON")
	}

	task := asynq.NewTask(message.Type(), payload)

	info, err := ac.client.EnqueueContext(ctx, task, asynq.Queue(message.Type()))
	if err != nil {
		return ungerr.Wrap(err, "error enqueuing task")
	}

	logger.Infof("enqueued task: ID=%s, Queue=%s", info.ID, info.Queue)
	return nil
}

func (ac *asynqClient) Shutdown() error {
	if err := ac.client.Close(); err != nil {
		return ungerr.Wrap(err, "error closing asynq client")
	}
	return nil
}
