package subscriber

import (
	"context"

	"github.com/hibiken/asynq"
	"github.com/itsLeonB/cashback/internal/core/config"
	"github.com/itsLeonB/cashback/internal/core/logger"
	"github.com/itsLeonB/cashback/internal/provider"
	"github.com/itsLeonB/ungerr"
)

type Subscriber struct {
	srv *asynq.Server
	mux *asynq.ServeMux
}

func Setup(providers *provider.Providers) (*Subscriber, error) {
	queues, queuePriorities := configureQueues(providers)

	// Configure asynq server
	srv := asynq.NewServer(
		provider.RedisClientOpts(config.Global.Valkey),
		asynq.Config{
			Concurrency: 3,
			Queues:      queuePriorities,
			ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
				if err != nil {
					logger.Errorf("error processing message: %v", err)
				}
			}),
			Logger: logger.Global,
		},
	)

	// Verify connectivity
	if err := srv.Ping(); err != nil {
		return nil, ungerr.Wrap(err, "error pinging valkey")
	}

	// Register handlers
	mux := asynq.NewServeMux()
	for _, q := range queues {
		mux.Handle(q.name, q.handler)
	}

	return &Subscriber{srv, mux}, nil
}

func (w *Subscriber) Start() error {
	if err := w.srv.Start(w.mux); err != nil {
		return ungerr.Wrap(err, "error starting subscriber")
	}
	return nil
}

func (w *Subscriber) Stop() {
	w.srv.Shutdown()
}
