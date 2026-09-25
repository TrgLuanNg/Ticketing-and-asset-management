package repository

import (
	"context"
	"strings"

	"go-backend/internals/core/domains"
	"go-backend/internals/core/dtos"

	"gorm.io/gorm"
)

type AssetRepository struct {
	db          *gorm.DB
	searchProps []string
}

func NewAssetRepository(db *gorm.DB) *AssetRepository {
	return &AssetRepository{
		db: db,
		searchProps: []string{
			"UnitCode", "Type", "UserName", "Brand", "Model", "Serial",
			"Department", "Office", "Company", "Area", "Status", "State",
			"Vendor", "Notes", "HRCode", "Email", "Detail1", "Detail2",
		},
	}
}

func (r *AssetRepository) DB() *gorm.DB {
	return r.db
}

func (r *AssetRepository) FindPagedWithAreas(ctx context.Context, filter *dtos.Filter, allowedAreas []string, seeAll bool) ([]domains.ASSET, int64, error) {
	var assets []domains.ASSET
	var total int64

	query := r.db.WithContext(ctx).Model(&domains.ASSET{})

	// Area filtering based on user permissions
	if !seeAll {
		if len(allowedAreas) > 0 {
			query = query.Where("\"Area\" IN (?)", allowedAreas)
		} else {
			// If no allowed areas and cannot see all, return empty
			return []domains.ASSET{}, 0, nil
		}
	}

	if filter != nil && !filter.SkipCount {
		if err := query.Count(&total).Error; err != nil {
			return nil, 0, err
		}
	}

	if filter != nil && filter.PageSize > 0 {
		pageIndex := filter.PageIndex
		if pageIndex <= 0 {
			pageIndex = 1
		}
		offset := (pageIndex - 1) * filter.PageSize
		query = query.Offset(offset).Limit(filter.PageSize)
	}

	if err := query.Find(&assets).Error; err != nil {
		return nil, 0, err
	}

	return assets, total, nil
}

func (r *AssetRepository) FindByID(ctx context.Context, unitCode string) (*domains.ASSET, error) {
	var asset domains.ASSET
	err := r.db.WithContext(ctx).Where("\"UnitCode\" = ?", unitCode).First(&asset).Error
	if err != nil {
		return nil, err
	}
	return &asset, nil
}

func (r *AssetRepository) FindSubAssets(ctx context.Context, mainUnitCode string) ([]domains.ASSET, error) {
	var subAssets []domains.ASSET
	err := r.db.WithContext(ctx).Where("\"MainAsset\" = ?", mainUnitCode).Find(&subAssets).Error
	return subAssets, err
}

func (r *AssetRepository) FindUnassignedSubAssets(ctx context.Context, filter *dtos.Filter) ([]dtos.UnassignedSubAssetDto, int64, error) {
	var dtosList []dtos.UnassignedSubAssetDto
	var total int64

	query := r.db.WithContext(ctx).Model(&domains.ASSET{}).
		Where("(\"MainAsset\" IS NULL OR \"MainAsset\" = '')").
		Where("LOWER(\"Type\") IN (?)", []string{"monitor", "mouse", "keyboard", "headset", "ram", "ssd", "hdd"})

	if filter != nil && strings.TrimSpace(filter.SearchText) != "" {
		likePattern := "%" + filter.SearchText + "%"
		query = query.Where("LOWER(\"UnitCode\") LIKE LOWER(?) OR LOWER(\"Model\") LIKE LOWER(?) OR LOWER(\"Serial\") LIKE LOWER(?)",
			likePattern, likePattern, likePattern)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if filter != nil && filter.PageSize > 0 {
		pageIndex := filter.PageIndex
		if pageIndex <= 0 {
			pageIndex = 1
		}
		offset := (pageIndex - 1) * filter.PageSize
		query = query.Offset(offset).Limit(filter.PageSize)
	}

	err := query.Select("\"UnitCode\", \"Type\", \"Model\", \"Serial\", \"Brand\", \"Status\", \"State\", \"Notes\", \"UsageDate\"").
		Scan(&dtosList).Error

	return dtosList, total, err
}

func (r *AssetRepository) UpdateMainAsset(ctx context.Context, unitCode string, mainAsset *string) error {
	return r.db.WithContext(ctx).Model(&domains.ASSET{}).
		Where("\"UnitCode\" = ?", unitCode).
		Update("MainAsset", mainAsset).Error
}

func (r *AssetRepository) GetDashboardMetrics(ctx context.Context, allowedAreas []string, seeAll bool, fromDate string) (*dtos.DashboardMetricsDto, error) {
	metrics := &dtos.DashboardMetricsDto{
		AssetsByType:    make(map[string]int),
		AssetsByDept:    make(map[string]int),
		AssetsByCompany: make(map[string]int),
		AssetsByArea:    make(map[string]int),
	}

	baseQuery := r.db.WithContext(ctx).Model(&domains.ASSET{})
	if !seeAll && len(allowedAreas) > 0 {
		baseQuery = baseQuery.Where("\"Area\" IN (?)", allowedAreas)
	}
	if fromDate != "" {
		baseQuery = baseQuery.Where("\"PurchasedDate\" >= ?", fromDate)
	}

	_ = baseQuery.Count(&metrics.TotalAssets).Error

	// By Status
	type statusCount struct {
		Status string
		Count  int64
	}
	var statuses []statusCount
	_ = r.db.WithContext(ctx).Model(&domains.ASSET{}).
		Select("\"Status\", COUNT(*) as count").
		Group("\"Status\"").Scan(&statuses).Error

	for _, s := range statuses {
		st := strings.ToLower(s.Status)
		switch {
		case strings.Contains(st, "in-use") || strings.Contains(st, "in use"):
			metrics.TotalInUse += s.Count
		case strings.Contains(st, "spare"):
			metrics.TotalSpare += s.Count
		case strings.Contains(st, "disposal"):
			metrics.TotalDisposed += s.Count
		case strings.Contains(st, "broken"):
			metrics.TotalBroken += s.Count
		}
	}

	// By Type
	type groupCount struct {
		Name  string
		Count int
	}
	var typeCounts []groupCount
	_ = r.db.WithContext(ctx).Model(&domains.ASSET{}).
		Select("\"Type\" as name, COUNT(*) as count").
		Group("\"Type\"").Scan(&typeCounts).Error
	for _, tc := range typeCounts {
		if tc.Name != "" {
			metrics.AssetsByType[tc.Name] = tc.Count
		}
	}

	// By Department
	var deptCounts []groupCount
	_ = r.db.WithContext(ctx).Model(&domains.ASSET{}).
		Select("\"Department\" as name, COUNT(*) as count").
		Group("\"Department\"").Scan(&deptCounts).Error
	for _, dc := range deptCounts {
		if dc.Name != "" {
			metrics.AssetsByDept[dc.Name] = dc.Count
		}
	}

	return metrics, nil
}

func (r *AssetRepository) GetRunningCountForPrefix(ctx context.Context, prefix string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domains.ASSET{}).
		Where("\"UnitCode\" LIKE ?", prefix+"%").
		Count(&count).Error
	return count, err
}
