package worker

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/itsLeonB/cashback/internal/adapters/worker/scheduler"
	"github.com/itsLeonB/cashback/internal/adapters/worker/subscriber"
	"github.com/itsLeonB/cashback/internal/core/logger"
	"github.com/itsLeonB/cashback/internal/provider"
)

type Worker struct {
	*subscriber.Subscriber
	*scheduler.Scheduler
	shutdownFunc func() error
}

func Setup() (*Worker, error) {
	providers, err := provider.All()
	if err != nil {
		return nil, err
	}

	subs, err := subscriber.Setup(providers)
	if err != nil {
		if e := providers.Shutdown(); e != nil {
			logger.Errorf("error shutting down resources: %v", e)
		}
		return nil, err
	}

	sched, err := scheduler.Setup(providers)
	if err != nil {
		if e := providers.Shutdown(); e != nil {
			logger.Errorf("error shutting down resources: %v", e)
		}
		return nil, err
	}

	return &Worker{subs, sched, providers.Shutdown}, nil
}

func (w *Worker) Run() {
	logger.Info("starting worker...")
	if err := w.Subscriber.Start(); err != nil {
		logger.Fatal(err)
	}

	w.Scheduler.Start()
	logger.Info("worker started")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig

	logger.Info("stopping worker...")
	w.Subscriber.Stop()
	w.Scheduler.Stop()
	logger.Info("worker stopped")

	if err := w.shutdownFunc(); err != nil {
		logger.Error(err)
	}
}
