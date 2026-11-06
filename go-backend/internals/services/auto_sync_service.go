package services

import (
	"context"

	"go-backend/internals/core/domains"
	"go-backend/internals/core/ports"
	"go-backend/pkg/logger"
)

type AutoSyncService struct {
	syncRepo ports.AutoSyncRepository
}

func NewAutoSyncService(syncRepo ports.AutoSyncRepository) *AutoSyncService {
	return &AutoSyncService{
		syncRepo: syncRepo,
	}
}

func (s *AutoSyncService) SyncAD(ctx context.Context) (bool, error) {
	logger.Info("SyncAD invoked (local database mode)")
	return true, nil
}

func (s *AutoSyncService) SyncPatchManager(ctx context.Context) (bool, error) {
	logger.Info("SyncPatchManager invoked (local database mode)")
	return true, nil
}

func (s *AutoSyncService) SyncTrendMicro(ctx context.Context) (bool, error) {
	logger.Info("SyncTrendMicro invoked (local database mode)")
	return true, nil
}

func (s *AutoSyncService) SyncMEDetail(ctx context.Context) (bool, error) {
	logger.Info("SyncMEDetail invoked (local database mode)")
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
	ok, _ = s.SyncMEDetail(ctx)
	results["MEDetail"] = ok
	return results
}

func (s *AutoSyncService) GetADAssets(ctx context.Context) ([]domains.ASSETSAD, error) {
	return s.syncRepo.GetADAssets(ctx)
}

func (s *AutoSyncService) GetTMAssets(ctx context.Context) ([]domains.ASSETSTM, error) {
	return s.syncRepo.GetTMAssets(ctx)
}

func (s *AutoSyncService) GetMEAssets(ctx context.Context) ([]domains.ASSETSME, error) {
	return s.syncRepo.GetMEAssets(ctx)
}

func (s *AutoSyncService) GetMEDetails(ctx context.Context) ([]domains.ASSETSEMPATCHDETAIL, error) {
	return s.syncRepo.GetMEDetails(ctx)
}
