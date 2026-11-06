package scheduler

import (
	"context"

	"go-backend/internals/core/ports"
	"go-backend/pkg/logger"

	"github.com/robfig/cron/v3"
)

type CronScheduler struct {
	cron    *cron.Cron
	useCase ports.AutoSyncUseCase
}

func NewCronScheduler(useCase ports.AutoSyncUseCase) *CronScheduler {
	return &CronScheduler{
		cron:    cron.New(),
		useCase: useCase,
	}
}

func (s *CronScheduler) Start(cronSchedule string) {
	if cronSchedule == "" {
		cronSchedule = "0 0 * * *"
	}

	_, err := s.cron.AddFunc(cronSchedule, func() {
		logger.Info("Executing automatic synchronization...")
		_ = s.useCase.SyncAll(context.Background())
	})
	if err != nil {
		logger.Errorf("Failed to schedule auto-sync cron: %v", err)
		return
	}

	s.cron.Start()
	logger.Infof("AutoSync cron scheduler started with schedule: %s", cronSchedule)
}

func (s *CronScheduler) Stop() {
	if s.cron != nil {
		s.cron.Stop()
	}
}
