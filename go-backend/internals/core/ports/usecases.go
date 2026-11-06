package ports

import (
	"context"
	"io"
	"time"

	"go-backend/internals/core/domains"
	"go-backend/internals/core/dtos"
)

type GenericUseCase[T any] interface {
	Search(ctx context.Context, filter *dtos.Filter) ([]T, int64, error)
	GetAll(ctx context.Context) ([]T, error)
	GetByID(ctx context.Context, id any) (*T, error)
	Create(ctx context.Context, userSession string, entity *T) error
	Update(ctx context.Context, userSession string, id any, entity *T) error
	Delete(ctx context.Context, userSession string, id any) error
}

type AuthUseCase interface {
	Login(ctx context.Context, username, password string) (*dtos.LoginResponseDto, error)
	GetResource(ctx context.Context, authCode string) (*dtos.LoginResponseDto, error)
	RefreshSession(ctx context.Context, refreshTokenStr string) (*dtos.LoginResponseDto, error)
	GetUserAllowedAreas(ctx context.Context, userID string) ([]string, bool)
}

type UserInfoUseCase interface {
	Search(ctx context.Context, filter *dtos.Filter) ([]dtos.UserInfoDto, int64, error)
	GetAll(ctx context.Context) ([]dtos.UserInfoDto, error)
	GetByID(ctx context.Context, userID string) (*dtos.UserInfoDto, error)
	GetHistory(ctx context.Context, userID string) ([]dtos.HistoryDto, error)
	ChangePassword(ctx context.Context, requesterId string, dto *dtos.ChangePasswordDto) error
	Insert(ctx context.Context, requesterId string, dto *dtos.UserInfoDto) (*dtos.UserInfoDto, error)
	Update(ctx context.Context, requesterId string, id string, dto *dtos.UserInfoDto) error
	Delete(ctx context.Context, requesterId string, id string) error
	Logout(ctx context.Context, userSession string)
}

type AssetUseCase interface {
	SearchGrid(ctx context.Context, filter *dtos.Filter, userSession string) ([]dtos.AssetDto, int64, error)
	GetAll(ctx context.Context, userSession string) ([]dtos.AssetDto, error)
	GetByID(ctx context.Context, userSession, unitCode string) (*dtos.AssetDto, error)
	Insert(ctx context.Context, userSession string, dto *dtos.AssetDto) (*dtos.AssetDto, error)
	Update(ctx context.Context, userSession string, unitCode string, dto *dtos.AssetDto) error
	Delete(ctx context.Context, userSession string, unitCode string) error
	GenerateUnitCode(ctx context.Context, assetType string, purchasedDate time.Time, company, area string) string
	GetDashboardMetrics(ctx context.Context, userSession, fromDate string) (*dtos.DashboardMetricsDto, error)
	GetSuggestions(ctx context.Context, filter *dtos.Filter, userSession string) ([]dtos.AssetSuggestionDto, error)
	GetUnassignedSubAssets(ctx context.Context, filter *dtos.Filter) ([]dtos.UnassignedSubAssetDto, int64, error)
	UpdateMainAsset(ctx context.Context, unitCode string, mainAsset *string) bool
	ExportExcel(ctx context.Context, filter *dtos.Filter, userSession string) ([]byte, error)
	ImportExcel(ctx context.Context, reader io.Reader, userSession string) (*dtos.ImportResultDto, error)
}

type HandoverUseCase interface {
	GetPagedHandovers(ctx context.Context, unitCode, userSession string, filter *dtos.Filter) ([]dtos.HandoverDto, int64, error)
	GetAll(ctx context.Context) ([]domains.HANDOVER, error)
	GetByID(ctx context.Context, id string) (*domains.HANDOVER, error)
	Insert(ctx context.Context, userSession string, dto *dtos.HandoverDto) (*domains.HANDOVER, error)
	Update(ctx context.Context, userSession, id string, entity *domains.HANDOVER) error
	Delete(ctx context.Context, userSession, id string) error
	ExportHandoverForm(ctx context.Context, handoverID, unitCode string) ([]byte, error)
}

type DisposalUseCase interface {
	GenericUseCase[domains.DISPOSAL]
	TransferToWFD(ctx context.Context, userSession string, req *dtos.TransferWfdRequestDto) (int, string, error)
	ConfirmDispose(ctx context.Context, userSession string, req *dtos.ConfirmDisposeRequestDto) (int, error)
	GetBatches(ctx context.Context, userSession string) ([]dtos.DisposalBatchDto, error)
	GetBatchDetails(ctx context.Context, userSession, idDisposal string) ([]dtos.DisposalBatchDetailDto, error)
	CheckUnitCodes(ctx context.Context, userSession string, req *dtos.CheckUnitCodesRequestDto) ([]dtos.UnitCodeCheckResultDto, error)
	ImportUnitCodes(ctx context.Context, userSession string, reader io.Reader) ([]string, error)
}

type ProblemReportUseCase interface {
	GenericUseCase[domains.PROBLEMREPORT]
	ExportProblemReport(ctx context.Context, id string) ([]byte, error)
}

type AssetsLogUseCase interface {
	GenericUseCase[domains.ASSETSLOG]
	ExportExcel(ctx context.Context, unitCode, fromDateStr, toDateStr string) ([]byte, error)
}

type AutoSyncUseCase interface {
	SyncAD(ctx context.Context) (bool, error)
	SyncPatchManager(ctx context.Context) (bool, error)
	SyncTrendMicro(ctx context.Context) (bool, error)
	SyncMEDetail(ctx context.Context) (bool, error)
	SyncAll(ctx context.Context) map[string]bool
	GetADAssets(ctx context.Context) ([]domains.ASSETSAD, error)
	GetTMAssets(ctx context.Context) ([]domains.ASSETSTM, error)
	GetMEAssets(ctx context.Context) ([]domains.ASSETSME, error)
	GetMEDetails(ctx context.Context) ([]domains.ASSETSEMPATCHDETAIL, error)
}
