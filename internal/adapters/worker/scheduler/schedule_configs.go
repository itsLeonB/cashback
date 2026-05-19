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
	}
}
