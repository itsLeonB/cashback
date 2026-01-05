package subscriber

import (
	"context"

	"github.com/hibiken/asynq"
	"github.com/itsLeonB/cashback/internal/core/config"
	"github.com/itsLeonB/cashback/internal/core/logger"
	"github.com/itsLeonB/cashback/internal/domain/message"
	"github.com/itsLeonB/cashback/internal/provider"
	"github.com/itsLeonB/cashback/internal/provider/datasource"
	"github.com/itsLeonB/ungerr"
)

type Subscriber struct {
	srv *asynq.Server
	mux *asynq.ServeMux
}

func Setup(providers *provider.Providers) (*Subscriber, error) {
	expenseBillUploadedQueue := message.ExpenseBillUploaded{}.Type()
	expenseBillTextExtractedQueue := message.ExpenseBillTextExtracted{}.Type()

	asynqCfg := asynq.Config{
		Concurrency: 3,
		Queues: map[string]int{
			expenseBillUploadedQueue:      3,
			expenseBillTextExtractedQueue: 3,
		},
		ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
			if err != nil {
				logger.Errorf("error processing message: %v", err)
			}
		}),
		Logger: logger.Global,
	}

	srv := asynq.NewServer(datasource.RedisClientOpts(config.Global.Valkey), asynqCfg)
	mux := asynq.NewServeMux()

	mux.Handle(expenseBillUploadedQueue, expenseBillUploadedHandler(providers.Services.ExpenseBill, providers.ExpenseBillTextExtracted))
	mux.Handle(expenseBillTextExtractedQueue, expenseBillTextExtractedHandler(providers.Services.GroupExpense))

	if err := srv.Ping(); err != nil {
		return nil, ungerr.Wrap(err, "error pinging valkey")
	}

	return &Subscriber{
		srv,
		mux,
	}, nil
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
