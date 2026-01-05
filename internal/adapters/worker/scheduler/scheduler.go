package scheduler

import (
	"context"

	"github.com/itsLeonB/cashback/internal/core/logger"
	"github.com/itsLeonB/cashback/internal/domain/service"
	"github.com/itsLeonB/cashback/internal/provider"
	"github.com/itsLeonB/ungerr"
	"github.com/robfig/cron/v3"
)

type Scheduler struct {
	billSvc service.ExpenseBillService
	cron    *cron.Cron
}

func Setup(providers *provider.Providers) (*Scheduler, error) {
	s := &Scheduler{providers.Services.ExpenseBill, cron.New()}

	_, err := s.cron.AddFunc("0 4 * * *", s.doCleanup)
	if err != nil {
		return nil, ungerr.Wrap(err, "error setting up cleanup job")
	}

	return s, nil
}

func (s *Scheduler) doCleanup() {
	logger.Info("starting daily expense bill cleanup...")
	if err := s.billSvc.Cleanup(context.Background()); err != nil {
		logger.Errorf("expense bill cleanup failed: %v", err)
		return
	}
	logger.Info("expense bill cleanup success")
}

func (s *Scheduler) Start() {
	s.cron.Start()
	logger.Info("scheduler started")
}

func (s *Scheduler) Stop() {
	logger.Info("stopping scheduler...")
	<-s.cron.Stop().Done()
	logger.Info("scheduler stopped")
}
