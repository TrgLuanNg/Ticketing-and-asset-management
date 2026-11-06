package repository

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"go-backend/internals/core/domains"

	"gorm.io/gorm"
)

type DisposalRepository struct {
	*GenericRepository[domains.DISPOSAL]
	db *gorm.DB
}

func NewDisposalRepository(db *gorm.DB) *DisposalRepository {
	return &DisposalRepository{
		GenericRepository: NewGenericRepository[domains.DISPOSAL](
			db,
			"UnitCode",
			[]string{"UnitCode", "IDDisposal", "Note", "UserID"},
			"\"StartDate\" DESC, \"IDDisposal\" DESC",
		),
		db: db,
	}
}

func (r *DisposalRepository) CountTodayBatches(ctx context.Context, todayPrefix string) (int64, error) {
	var ids []string
	err := getDB(ctx, r.db).Model(&domains.DISPOSAL{}).
		Where("\"IDDisposal\" LIKE ?", todayPrefix+"%").
		Distinct("\"IDDisposal\"").
		Pluck("\"IDDisposal\"", &ids).Error
	return int64(len(ids)), err
}

func (r *DisposalRepository) ExistsByUnitCode(ctx context.Context, unitCode string) bool {
	var count int64
	getDB(ctx, r.db).Model(&domains.DISPOSAL{}).Where("\"UnitCode\" = ?", unitCode).Count(&count)
	return count > 0
}

func (r *DisposalRepository) FindByIDDisposal(ctx context.Context, idDisposal string) ([]domains.DISPOSAL, error) {
	var list []domains.DISPOSAL
	err := getDB(ctx, r.db).Where("\"IDDisposal\" = ?", idDisposal).Find(&list).Error
	return list, err
}

func (r *DisposalRepository) UpdateDisposalItem(ctx context.Context, d *domains.DISPOSAL) error {
	return getDB(ctx, r.db).Model(&domains.DISPOSAL{}).
		Where("\"UnitCode\" = ? AND \"IDDisposal\" = ?", d.UnitCode, d.IDDisposal).
		Updates(map[string]any{
			"Status":  d.Status,
			"Note":    d.Note,
			"EndDate": d.EndDate,
		}).Error
}

type ProblemReportRepository struct {
	*GenericRepository[domains.PROBLEMREPORT]
	db *gorm.DB
}

func NewProblemReportRepository(db *gorm.DB) *ProblemReportRepository {
	return &ProblemReportRepository{
		GenericRepository: NewGenericRepository[domains.PROBLEMREPORT](
			db,
			"ID",
			[]string{"ID", "UnitCode", "UserCreated", "ProblemSummary", "ProblemNo", "JiraCode", "UserReport"},
			"\"ReportDate\" DESC, \"ID\" DESC",
		),
		db: db,
	}
}

func (r *ProblemReportRepository) GetMaxNumericID(ctx context.Context) (int64, error) {
	var ids []string
	err := getDB(ctx, r.db).Model(&domains.PROBLEMREPORT{}).Pluck("\"ID\"", &ids).Error
	if err != nil {
		return 0, err
	}
	var maxID int64
	for _, s := range ids {
		if v, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64); err == nil && v > maxID {
			maxID = v
		}
	}
	return maxID, nil
}

type AssetsLogRepository struct {
	*GenericRepository[domains.ASSETSLOG]
	db *gorm.DB
}

func NewAssetsLogRepository(db *gorm.DB) *AssetsLogRepository {
	return &AssetsLogRepository{
		GenericRepository: NewGenericRepository[domains.ASSETSLOG](
			db,
			"ID",
			[]string{"UnitCode", "UserID", "Action", "ActionLogs"},
			"\"DateModified\" DESC, \"ID\" DESC",
		),
		db: db,
	}
}

func (r *AssetsLogRepository) FindForExport(ctx context.Context, unitCode string, fromDate, toDate *time.Time) ([]domains.ASSETSLOG, error) {
	var logs []domains.ASSETSLOG
	q := getDB(ctx, r.db).Model(&domains.ASSETSLOG{}).Where("\"UnitCode\" = ?", unitCode)
	if fromDate != nil {
		q = q.Where("\"DateModified\" >= ?", *fromDate)
	}
	if toDate != nil {
		endOfDay := time.Date(toDate.Year(), toDate.Month(), toDate.Day(), 23, 59, 59, 999999999, toDate.Location())
		q = q.Where("\"DateModified\" <= ?", endOfDay)
	}
	err := q.Order("\"DateModified\" DESC, \"ID\" DESC").Find(&logs).Error
	return logs, err
}

type DepartmentRepository struct {
	*GenericRepository[domains.DEPARTMENT]
	db *gorm.DB
}

func NewDepartmentRepository(db *gorm.DB) *DepartmentRepository {
	return &DepartmentRepository{
		GenericRepository: NewGenericRepository[domains.DEPARTMENT](
			db,
			"ID",
			[]string{"ID", "Name", "Description", "ParentId"},
			"\"ID\" ASC",
		),
		db: db,
	}
}

func (r *DepartmentRepository) GetNextID(ctx context.Context) (string, error) {
	var ids []string
	if err := getDB(ctx, r.db).Model(&domains.DEPARTMENT{}).Pluck("\"ID\"", &ids).Error; err != nil {
		return "", err
	}
	maxID := 0
	for _, s := range ids {
		if v, err := strconv.Atoi(strings.TrimSpace(s)); err == nil && v > maxID {
			maxID = v
		}
	}
	return fmt.Sprintf("%05d", maxID+1), nil
}

type ItemMasterDataRepository struct {
	*GenericRepository[domains.ITEMMASTERDATA]
	db *gorm.DB
}

func NewItemMasterDataRepository(db *gorm.DB) *ItemMasterDataRepository {
	return &ItemMasterDataRepository{
		GenericRepository: NewGenericRepository[domains.ITEMMASTERDATA](
			db,
			"ID",
			[]string{"ID", "Item", "TechSpecs"},
			"\"ID\" ASC",
		),
		db: db,
	}
}

func (r *ItemMasterDataRepository) GetNextID(ctx context.Context) (string, error) {
	var ids []string
	if err := getDB(ctx, r.db).Model(&domains.ITEMMASTERDATA{}).Pluck("\"ID\"", &ids).Error; err != nil {
		return "", err
	}
	maxID := 0
	for _, s := range ids {
		if v, err := strconv.Atoi(strings.TrimSpace(s)); err == nil && v > maxID {
			maxID = v
		}
	}
	return strconv.Itoa(maxID + 1), nil
}

type AutoSyncRepository struct {
	db *gorm.DB
}

func NewAutoSyncRepository(db *gorm.DB) *AutoSyncRepository {
	return &AutoSyncRepository{db: db}
}

func (r *AutoSyncRepository) GetADAssets(ctx context.Context) ([]domains.ASSETSAD, error) {
	var list []domains.ASSETSAD
	err := getDB(ctx, r.db).Find(&list).Error
	return list, err
}

func (r *AutoSyncRepository) GetTMAssets(ctx context.Context) ([]domains.ASSETSTM, error) {
	var list []domains.ASSETSTM
	err := getDB(ctx, r.db).Find(&list).Error
	return list, err
}

func (r *AutoSyncRepository) GetMEAssets(ctx context.Context) ([]domains.ASSETSME, error) {
	var list []domains.ASSETSME
	err := getDB(ctx, r.db).Find(&list).Error
	return list, err
}

func (r *AutoSyncRepository) GetMEDetails(ctx context.Context) ([]domains.ASSETSEMPATCHDETAIL, error) {
	var list []domains.ASSETSEMPATCHDETAIL
	err := getDB(ctx, r.db).Find(&list).Error
	return list, err
}
