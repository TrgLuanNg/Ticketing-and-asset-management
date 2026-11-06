package ports

import (
	"context"
	"time"

	"go-backend/internals/core/domains"
	"go-backend/internals/core/dtos"
)

// TransactionManager defines a unit-of-work transaction boundary so UseCases remain decoupled from GORM.
type TransactionManager interface {
	RunInTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

// GenericRepository defines standard CRUD and paged search operations for any domain entity T.
type GenericRepository[T any] interface {
	FindPaged(ctx context.Context, filter *dtos.Filter) ([]T, int64, error)
	FindAll(ctx context.Context) ([]T, error)
	FindByID(ctx context.Context, id any) (*T, error)
	Create(ctx context.Context, entity *T) error
	Update(ctx context.Context, id any, entity *T) error
	Delete(ctx context.Context, id any) error
}

type AssetRepository interface {
	FindPagedWithAreas(ctx context.Context, filter *dtos.Filter, allowedAreas []string, seeAll bool, userSession string) ([]domains.ASSET, int64, error)
	FindAllWithAreas(ctx context.Context, allowedAreas []string, seeAll bool, userSession string) ([]domains.ASSET, error)
	FindByID(ctx context.Context, unitCode string) (*domains.ASSET, error)
	FindSubAssets(ctx context.Context, mainUnitCode string) ([]domains.ASSET, error)
	FindUnassignedSubAssets(ctx context.Context, filter *dtos.Filter) ([]dtos.UnassignedSubAssetDto, int64, error)
	Create(ctx context.Context, asset *domains.ASSET) error
	Update(ctx context.Context, unitCode string, asset *domains.ASSET) error
	Delete(ctx context.Context, unitCode string) error
	UpdateMainAsset(ctx context.Context, unitCode string, mainAsset *string) error
	UnitCodeExists(ctx context.Context, unitCode string) bool
	SerialExists(ctx context.Context, serial string) bool
	GetMatchingUnitCodes(ctx context.Context, prefix string) ([]string, error)
	GetDashboardMetrics(ctx context.Context, allowedAreas []string, seeAll bool, userSession, fromDate string) (*dtos.DashboardMetricsDto, error)
	GetSuggestions(ctx context.Context, filter *dtos.Filter, allowedAreas []string, seeAll bool, userSession string) ([]dtos.AssetSuggestionDto, error)
	AddAssetLog(ctx context.Context, log *domains.ASSETSLOG) error
	GetAllDepartments(ctx context.Context) ([]domains.DEPARTMENT, error)
	GetAllStates(ctx context.Context) ([]domains.STATEMASTERDATA, error)
	GetAllDeviceConfigs(ctx context.Context) ([]domains.DEVICECONFIG, error)
	GetAllItemMasterData(ctx context.Context) ([]domains.ITEMMASTERDATA, error)
	GetAllTypes(ctx context.Context) ([]domains.TYPE, error)
}

type UserRepository interface {
	FindPaged(ctx context.Context, filter *dtos.Filter) ([]domains.USERINFO, int64, error)
	FindAll(ctx context.Context) ([]domains.USERINFO, error)
	FindByUserID(ctx context.Context, userID string) (*domains.USERINFO, error)
	CreateUser(ctx context.Context, user *domains.USERINFO) error
	UpdateUser(ctx context.Context, user *domains.USERINFO) error
	UpdateUserFields(ctx context.Context, id string, fields map[string]any) error
	DeleteUser(ctx context.Context, userID string) error
	FindPermission(ctx context.Context, userID string) (*domains.PERMISSION, error)
	CreatePermission(ctx context.Context, perm *domains.PERMISSION) error
	UpsertPermission(ctx context.Context, perm *domains.PERMISSION) error
	CreateRefreshToken(ctx context.Context, token *domains.REFRESH_TOKEN) error
	FindRefreshToken(ctx context.Context, tokenStr string) (*domains.REFRESH_TOKEN, error)
	UpdateRefreshToken(ctx context.Context, token *domains.REFRESH_TOKEN) error
	CleanupExpiredTokens(ctx context.Context)
	GetUserHistory(ctx context.Context, userID string) ([]domains.HISTORY, error)
	AddHistory(ctx context.Context, userID, action string)
}

type HandoverRepository interface {
	GetPagedHandovers(ctx context.Context, unitCode, userSession, userName string, isViewer bool, filter *dtos.Filter) ([]domains.HANDOVEROTHER, int64, error)
	FindAll(ctx context.Context) ([]domains.HANDOVER, error)
	FindByID(ctx context.Context, id string) (*domains.HANDOVER, error)
	FindOtherByID(ctx context.Context, id, unitCode string) (*domains.HANDOVEROTHER, error)
	CreateHandover(ctx context.Context, h *domains.HANDOVER) error
	UpdateHandover(ctx context.Context, id string, h *domains.HANDOVER) error
	DeleteHandover(ctx context.Context, id string) error
	CreateHandoverOther(ctx context.Context, ho *domains.HANDOVEROTHER) error
	GetMaxHandoverOtherID(ctx context.Context) (int64, error)
	CountHandoverOthersByPrefix(ctx context.Context, prefix string) (int64, error)
	UpsertUpdateEmployeeCode(ctx context.Context, uec *domains.UpdateEmployeeCode) error
	GetAccessoriesByUnitCode(ctx context.Context, unitCode string) ([]domains.ACCESSORY, error)
	GetSoftwaresByUnitCode(ctx context.Context, unitCode string) ([]domains.SOFTWARE, error)
}

type DisposalRepository interface {
	GenericRepository[domains.DISPOSAL]
	CountTodayBatches(ctx context.Context, todayPrefix string) (int64, error)
	ExistsByUnitCode(ctx context.Context, unitCode string) bool
	FindByIDDisposal(ctx context.Context, idDisposal string) ([]domains.DISPOSAL, error)
	UpdateDisposalItem(ctx context.Context, d *domains.DISPOSAL) error
}

type ProblemReportRepository interface {
	GenericRepository[domains.PROBLEMREPORT]
	GetMaxNumericID(ctx context.Context) (int64, error)
}

type AssetsLogRepository interface {
	GenericRepository[domains.ASSETSLOG]
	FindForExport(ctx context.Context, unitCode string, fromDate, toDate *time.Time) ([]domains.ASSETSLOG, error)
}

type DepartmentRepository interface {
	GenericRepository[domains.DEPARTMENT]
	GetNextID(ctx context.Context) (string, error)
}

type ItemMasterDataRepository interface {
	GenericRepository[domains.ITEMMASTERDATA]
	GetNextID(ctx context.Context) (string, error)
}

type AutoSyncRepository interface {
	GetADAssets(ctx context.Context) ([]domains.ASSETSAD, error)
	GetTMAssets(ctx context.Context) ([]domains.ASSETSTM, error)
	GetMEAssets(ctx context.Context) ([]domains.ASSETSME, error)
	GetMEDetails(ctx context.Context) ([]domains.ASSETSEMPATCHDETAIL, error)
}
