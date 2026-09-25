package services

import (
	"context"

	"go-backend/pkg/logger"

	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

type AutoSyncService struct {
	db   *gorm.DB
	cron *cron.Cron
}

func NewAutoSyncService(db *gorm.DB) *AutoSyncService {
	return &AutoSyncService{
		db:   db,
		cron: cron.New(),
	}
}

func (s *AutoSyncService) StartScheduler(cronSchedule string) {
	if cronSchedule == "" {
		cronSchedule = "0 0 * * *" // Midnight daily
	}

	_, err := s.cron.AddFunc(cronSchedule, func() {
		logger.Info("Executing automatic synchronization...")
		_ = s.SyncAll(context.Background())
	})

	if err != nil {
		logger.Errorf("Failed to schedule auto-sync cron: %v", err)
		return
	}

	s.cron.Start()
	logger.Infof("AutoSync cron scheduler started with schedule: %s", cronSchedule)
}

func (s *AutoSyncService) StopScheduler() {
	if s.cron != nil {
		s.cron.Stop()
	}
}

func (s *AutoSyncService) SyncAD(ctx context.Context) (bool, error) {
	logger.Info("Syncing Active Directory...")
	// Logic for AD sync
	return true, nil
}

func (s *AutoSyncService) SyncPatchManager(ctx context.Context) (bool, error) {
	logger.Info("Syncing Patch Manager Plus...")
	// Logic for PatchManager sync
	return true, nil
}

func (s *AutoSyncService) SyncTrendMicro(ctx context.Context) (bool, error) {
	logger.Info("Syncing Trend Micro Apex Central...")
	// Logic for TrendMicro sync
	return true, nil
}

func (s *AutoSyncService) SyncAll(ctx context.Context) map[string]bool {
	results := make(map[string]bool)
	ok, _ := s.SyncAD(ctx)
	results["AD"] = ok
	ok, _ = s.SyncPatchManager(ctx)
	results["PatchManager"] = ok
	ok, _ = s.SyncTrendMicro(ctx)
	results["TrendMicro"] = ok
	return results
}
