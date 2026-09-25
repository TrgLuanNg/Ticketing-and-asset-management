package services

import (
	"context"

	"go-backend/internals/core/domains"
	"go-backend/internals/core/dtos"
	"go-backend/internals/repository"

	"gorm.io/gorm"
)

type MasterService struct {
	DeptRepo       *repository.GenericRepository[domains.DEPARTMENT]
	TypeRepo       *repository.GenericRepository[domains.TYPE]
	DeviceRepo     *repository.GenericRepository[domains.DEVICECONFIG]
	DisposalRepo   *repository.GenericRepository[domains.DISPOSAL]
	SoftwareRepo   *repository.GenericRepository[domains.SOFTWARE]
	TechSpecRepo   *repository.GenericRepository[domains.KEYTECHSPEC]
	ItemDataRepo   *repository.GenericRepository[domains.ITEMMASTERDATA]
	StateDataRepo  *repository.GenericRepository[domains.STATEMASTERDATA]
	TitleEmpRepo   *repository.GenericRepository[domains.EmployeeTitle]
	ProblemRepo    *repository.GenericRepository[domains.PROBLEMREPORT]
	AssetsLogRepo  *repository.GenericRepository[domains.ASSETSLOG]
	PermissionRepo *repository.GenericRepository[domains.PERMISSION]
}

func NewMasterService(db *gorm.DB) *MasterService {
	return &MasterService{
		DeptRepo: repository.NewGenericRepository[domains.DEPARTMENT](db, "ID", []string{"ID", "Name", "Description", "ParentId"}),
		TypeRepo: repository.NewGenericRepository[domains.TYPE](db, "Type", []string{"Type", "TypeName", "VN"}),
		DeviceRepo: repository.NewGenericRepository[domains.DEVICECONFIG](db, "Device", []string{"Device", "Detail1", "Detail2", "Detail3"}),
		DisposalRepo: repository.NewGenericRepository[domains.DISPOSAL](db, "UnitCode", []string{"UnitCode", "IDDisposal", "Note", "UserID"}),
		SoftwareRepo: repository.NewGenericRepository[domains.SOFTWARE](db, "ID", []string{"UnitCode", "Software", "Version"}),
		TechSpecRepo: repository.NewGenericRepository[domains.KEYTECHSPEC](db, "Name", []string{"Name", "DescriptionTechSpecs"}),
		ItemDataRepo: repository.NewGenericRepository[domains.ITEMMASTERDATA](db, "ID", []string{"ID", "Item", "TechSpecs"}),
		StateDataRepo: repository.NewGenericRepository[domains.STATEMASTERDATA](db, "ID", []string{"ID", "State", "Description"}),
		TitleEmpRepo: repository.NewGenericRepository[domains.EmployeeTitle](db, "EmployeeCode", []string{"EmployeeCode", "Title"}),
		ProblemRepo: repository.NewGenericRepository[domains.PROBLEMREPORT](db, "ID", []string{"ID", "UnitCode", "UserCreated", "ProblemSummary", "ProblemNo", "JiraCode"}),
		AssetsLogRepo: repository.NewGenericRepository[domains.ASSETSLOG](db, "ID", []string{"UnitCode", "UserID", "Action", "ActionLogs"}),
		PermissionRepo: repository.NewGenericRepository[domains.PERMISSION](db, "UserID", []string{"UserID"}),
	}
}

// Master CRUD Helpers
func SearchGeneric[T any](ctx context.Context, repo *repository.GenericRepository[T], filter *dtos.Filter) ([]T, int64, error) {
	return repo.FindPaged(ctx, filter)
}

func GetAllGeneric[T any](ctx context.Context, repo *repository.GenericRepository[T]) ([]T, error) {
	return repo.FindAll(ctx)
}

func GetByIDGeneric[T any](ctx context.Context, repo *repository.GenericRepository[T], id any) (*T, error) {
	return repo.FindByID(ctx, id)
}

func CreateGeneric[T any](ctx context.Context, repo *repository.GenericRepository[T], entity *T) error {
	return repo.Create(ctx, entity)
}

func UpdateGeneric[T any](ctx context.Context, repo *repository.GenericRepository[T], id any, entity *T) error {
	return repo.Update(ctx, id, entity)
}

func DeleteGeneric[T any](ctx context.Context, repo *repository.GenericRepository[T], id any) error {
	return repo.Delete(ctx, id)
}
