package scheduler

import (
	"context"
	"errors"

	"github.com/itsLeonB/cashback/internal/core/logger"
	"github.com/itsLeonB/cashback/internal/domain/service"
	"github.com/itsLeonB/cashback/internal/domain/service/monetization"
	"github.com/itsLeonB/cashback/internal/provider"
	"github.com/itsLeonB/ungerr"
	"github.com/robfig/cron/v3"
)

type Scheduler struct {
	billSvc         service.ExpenseBillService
	subscriptionSvc monetization.SubscriptionService
	cron            *cron.Cron
}

func Setup(providers *provider.Providers) (*Scheduler, error) {
	s := &Scheduler{providers.Services.ExpenseBill, providers.Services.Subscription, cron.New()}

	var err error
	if _, e := s.cron.AddFunc("0 21 * * *", s.doCleanup); e != nil {
		err = errors.Join(err, e)
	}

	if _, e := s.cron.AddFunc("0 21 * * *", s.doPastDueUpdates); e != nil {
		err = errors.Join(err, e)
	}

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

func (s *Scheduler) doPastDueUpdates() {
	logger.Info("starting daily past-due subscription updates...")
	if err := s.subscriptionSvc.UpdatePastDues(context.Background()); err != nil {
		logger.Errorf("past-due subscription updates failed: %v", err)
		return
	}
	logger.Info("past-due subscription updates success")
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
