package scheduler

import "context"

type scheduleConfig struct {
	cronSpec string
	jobFn    func(context.Context) error
	jobName  string
}

func (s *Scheduler) getSchedules() []scheduleConfig {
	return []scheduleConfig{
		{
			cronSpec: "0 4 * * *",
			jobFn:    s.billSvc.Cleanup,
			jobName:  "expense bill cleanup",
		},
		{
			cronSpec: "0 12 * * *",
			jobFn:    s.subscriptionSvc.UpdatePastDues,
			jobName:  "past-due subscription updates",
		},
		{
			cronSpec: "0 12 * * *",
			jobFn:    s.subscriptionSvc.PublishSubscriptionDueNotifications,
			jobName:  "publish nearing due-date subscription notification",
		},
	}
}
