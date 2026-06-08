package scheduler

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"

	"github.com/itsLeonB/cashback/internal/core/logger"
	"github.com/itsLeonB/cashback/internal/core/otel"
	"github.com/itsLeonB/cashback/internal/domain/service"
	"github.com/itsLeonB/cashback/internal/domain/service/monetization"
	"github.com/itsLeonB/cashback/internal/provider"
	"github.com/itsLeonB/ungerr"
	"github.com/robfig/cron/v3"
	"go.opentelemetry.io/otel/codes"
)

type Scheduler struct {
	billSvc         service.ExpenseBillService
	subscriptionSvc monetization.SubscriptionService
	cron            *cron.Cron
}

func Setup(providers *provider.Providers) (*Scheduler, error) {
	s := &Scheduler{providers.Services.ExpenseBill, providers.Services.Subscription, cron.New()}
	schedules := s.getSchedules()

	var err error
	for _, schedule := range schedules {
		if _, e := s.cron.AddFunc(schedule.cronSpec, s.jobWrapper(schedule.jobName, schedule.jobFn)); e != nil {
			err = errors.Join(err, e)
		}
	}
	if err != nil {
		return nil, ungerr.Wrap(err, "error scheduling jobs")
	}

	return s, nil
}

func (s *Scheduler) jobWrapper(jobName string, jobFn func(context.Context) error) func() {
	return func() {
		ctx, span := otel.Tracer.Start(context.Background(), jobName)
		defer span.End()

		defer func() {
			if r := recover(); r != nil {
				stack := debug.Stack()
				err := fmt.Errorf("panic in job %s: %v\n%s", jobName, r, stack)
				span.RecordError(err)
				span.SetStatus(codes.Error, "panic recovered")
				logger.Errorf("panic in job %s: %v\n%s", jobName, r, stack)
			}
		}()

		logger.Infof("starting %s...", jobName)
		if err := jobFn(ctx); err != nil {
			logger.Errorf("%s failed: %v", jobName, err)
			return
		}

		logger.Infof("%s success", jobName)
	}
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
