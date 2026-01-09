package provider

import (
	"crypto/tls"

	"github.com/hibiken/asynq"
	adapters "github.com/itsLeonB/cashback/internal/adapters/queue"
	"github.com/itsLeonB/cashback/internal/core/config"
	"github.com/itsLeonB/cashback/internal/core/service/queue"
)

type Queues struct {
	taskQueue queue.TaskQueue
}

func ProvideQueues(cfg config.Valkey) (*Queues, error) {
	taskQueue, err := adapters.NewTaskQueue(RedisClientOpts(cfg))
	if err != nil {
		return nil, err
	}

	return &Queues{
		taskQueue,
	}, nil
}

func RedisClientOpts(cfg config.Valkey) asynq.RedisClientOpt {
	opt := asynq.RedisClientOpt{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.Db,
	}

	if cfg.EnableTls {
		opt.TLSConfig = &tls.Config{MinVersion: tls.VersionTLS12}
	}

	return opt
}

func (q *Queues) Shutdown() error {
	return q.taskQueue.Shutdown()
}
