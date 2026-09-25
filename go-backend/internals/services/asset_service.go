package services

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"go-backend/internals/core/domains"
	"go-backend/internals/core/dtos"
	"go-backend/internals/repository"
	"go-backend/pkg/logger"

	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

type AssetService struct {
	assetRepo   *repository.AssetRepository
	userRepo    *repository.UserRepository
	authService *AuthService
}

func NewAssetService(assetRepo *repository.AssetRepository, userRepo *repository.UserRepository, authService *AuthService) *AssetService {
	return &AssetService{
		assetRepo:   assetRepo,
		userRepo:    userRepo,
		authService: authService,
	}
}

func (s *AssetService) SearchGrid(ctx context.Context, filter *dtos.Filter, userSession string) ([]dtos.AssetDto, int64, error) {
	allowedAreas, seeAll := s.authService.GetUserAllowedAreas(ctx, userSession)
	entities, total, err := s.assetRepo.FindPagedWithAreas(ctx, filter, allowedAreas, seeAll)
	if err != nil {
		return nil, 0, err
	}

	dtosList := make([]dtos.AssetDto, len(entities))
	for i, e := range entities {
		dtosList[i] = s.mapToDto(&e)
	}

	return dtosList, total, nil
}

func (s *AssetService) GetByID(ctx context.Context, unitCode string) (*dtos.AssetDto, error) {
	asset, err := s.assetRepo.FindByID(ctx, unitCode)
	if err != nil {
		return nil, err
	}

	dto := s.mapToDto(asset)

	// Fetch sub-assets
	subAssets, _ := s.assetRepo.FindSubAssets(ctx, unitCode)
	for _, sub := range subAssets {
		dto.SubAssets = append(dto.SubAssets, s.mapToDto(&sub))
	}

	return &dto, nil
}

func (s *AssetService) Insert(ctx context.Context, userSession string, dto *dtos.AssetDto) (*dtos.AssetDto, error) {
	if !s.HasModifyPermission(ctx, userSession) {
		return nil, errors.New("you do not have permission to modify assets")
	}

	entity := s.mapToEntity(dto)

	// Auto generate UnitCode if empty
	if entity.UnitCode == "" {
		purchasedTime := time.Now()
		if entity.PurchasedDate != nil && !entity.PurchasedDate.IsZero() {
			purchasedTime = entity.PurchasedDate.Time
		}
		entity.UnitCode = s.GenerateUnitCode(ctx, entity.Type, purchasedTime, entity.Company, entity.Area)
	}

	db := s.assetRepo.DB()
	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(entity).Error; err != nil {
			return err
		}

		// Insert SubAssets if present
		for _, subDto := range dto.SubAssets {
			subEntity := s.mapToEntity(&subDto)
			subEntity.MainAsset = entity.UnitCode
			subEntity.UserName = entity.UserName
			subEntity.Department = entity.Department
			subEntity.Office = entity.Office
			subEntity.Company = entity.Company
			subEntity.Area = entity.Area
			subEntity.Email = entity.Email
			subEntity.HRCode = entity.HRCode
			subEntity.State = entity.State
			subEntity.PurchasedDate = entity.PurchasedDate
			subEntity.Vendor = entity.Vendor
			subEntity.TypeEXP = entity.TypeEXP

			if subEntity.UnitCode == "" {
				subPurchasedTime := time.Now()
				if subEntity.PurchasedDate != nil && !subEntity.PurchasedDate.IsZero() {
					subPurchasedTime = subEntity.PurchasedDate.Time
				}
				subEntity.UnitCode = s.GenerateUnitCode(ctx, subEntity.Type, subPurchasedTime, subEntity.Company, subEntity.Area)
			}

			if err := tx.Create(subEntity).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	createdDto := s.mapToDto(entity)
	return &createdDto, nil
}

func (s *AssetService) Update(ctx context.Context, userSession string, unitCode string, dto *dtos.AssetDto) error {
	if !s.HasModifyPermission(ctx, userSession) {
		return errors.New("you do not have permission to modify assets")
	}

	entity := s.mapToEntity(dto)
	entity.UnitCode = unitCode

	return s.assetRepo.DB().WithContext(ctx).Model(&domains.ASSET{}).
		Where("\"UnitCode\" = ?", unitCode).
		Updates(entity).Error
}

func (s *AssetService) Delete(ctx context.Context, userSession string, unitCode string) error {
	if !s.HasDeletePermission(ctx, userSession) {
		return errors.New("you do not have permission to delete assets")
	}

	return s.assetRepo.DB().WithContext(ctx).
		Where("\"UnitCode\" = ?", unitCode).
		Delete(&domains.ASSET{}).Error
}

func (s *AssetService) GenerateUnitCode(ctx context.Context, assetType string, purchasedDate time.Time, company, area string) string {
	typeUpper := strings.ToUpper(strings.TrimSpace(assetType))
	prefix := "OT"
	switch typeUpper {
	case "LAPTOP":
		prefix = "LT"
	case "DESKTOP", "PC":
		prefix = "PC"
	case "MONITOR":
		prefix = "MO"
	case "SERVER":
		prefix = "SV"
	case "PRINTER":
		prefix = "PR"
	case "NETWORK", "SWITCH", "ROUTER":
		prefix = "NW"
	}

	yearStr := purchasedDate.Format("06") // 2-digit year e.g. 26
	codePrefix := fmt.Sprintf("%s%s", prefix, yearStr)

	count, err := s.assetRepo.GetRunningCountForPrefix(ctx, codePrefix)
	if err != nil {
		count = 0
	}

	return fmt.Sprintf("%s%04d", codePrefix, count+1)
}

func (s *AssetService) GetDashboardMetrics(ctx context.Context, userSession, fromDate string) (*dtos.DashboardMetricsDto, error) {
	allowedAreas, seeAll := s.authService.GetUserAllowedAreas(ctx, userSession)
	return s.assetRepo.GetDashboardMetrics(ctx, allowedAreas, seeAll, fromDate)
}

func (s *AssetService) GetUnassignedSubAssets(ctx context.Context, filter *dtos.Filter) ([]dtos.UnassignedSubAssetDto, int64, error) {
	return s.assetRepo.FindUnassignedSubAssets(ctx, filter)
}

func (s *AssetService) UpdateMainAsset(ctx context.Context, unitCode string, mainAsset *string) bool {
	err := s.assetRepo.UpdateMainAsset(ctx, unitCode, mainAsset)
	return err == nil
}

func (s *AssetService) HasModifyPermission(ctx context.Context, userSession string) bool {
	user, _ := s.userRepo.FindByUserID(ctx, userSession)
	isAdmin := user != nil && strings.EqualFold(user.Role, "admin")

	perm, _ := s.userRepo.FindPermission(ctx, userSession)
	if perm == nil {
		return isAdmin
	}
	return (perm.AssetsModify != nil && *perm.AssetsModify) || (perm.AssetsModify == nil && isAdmin)
}

func (s *AssetService) HasDeletePermission(ctx context.Context, userSession string) bool {
	user, _ := s.userRepo.FindByUserID(ctx, userSession)
	isAdmin := user != nil && strings.EqualFold(user.Role, "admin")

	perm, _ := s.userRepo.FindPermission(ctx, userSession)
	if perm == nil {
		return isAdmin
	}
	return (perm.AssetsDel != nil && *perm.AssetsDel) || (perm.AssetsDel == nil && isAdmin)
}

// ExportExcel exports assets to .xlsx byte slice using excelize
func (s *AssetService) ExportExcel(ctx context.Context, filter *dtos.Filter, userSession string) ([]byte, error) {
	allowedAreas, seeAll := s.authService.GetUserAllowedAreas(ctx, userSession)
	filterCopy := *filter
	filterCopy.PageSize = 100000 // Export all matching
	filterCopy.SkipCount = true

	assets, _, err := s.assetRepo.FindPagedWithAreas(ctx, &filterCopy, allowedAreas, seeAll)
	if err != nil || len(assets) == 0 {
		return nil, errors.New("no data to export")
	}

	f := excelize.NewFile()
	defer f.Close()
	sheet := "Assets"
	f.SetSheetName("Sheet1", sheet)

	headers := []string{
		"Unit Code", "Type", "User Name", "Department", "Office", "Company",
		"Area", "Brand", "Model", "Serial", "Status", "State", "Purchased Date", "Notes",
	}

	for colIdx, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(colIdx+1, 1)
		_ = f.SetCellValue(sheet, cell, h)
	}

	for rowIdx, a := range assets {
		row := rowIdx + 2
		_ = f.SetCellValue(sheet, fmt.Sprintf("A%d", row), a.UnitCode)
		_ = f.SetCellValue(sheet, fmt.Sprintf("B%d", row), a.Type)
		_ = f.SetCellValue(sheet, fmt.Sprintf("C%d", row), a.UserName)
		_ = f.SetCellValue(sheet, fmt.Sprintf("D%d", row), a.Department)
		_ = f.SetCellValue(sheet, fmt.Sprintf("E%d", row), a.Office)
		_ = f.SetCellValue(sheet, fmt.Sprintf("F%d", row), a.Company)
		_ = f.SetCellValue(sheet, fmt.Sprintf("G%d", row), a.Area)
		_ = f.SetCellValue(sheet, fmt.Sprintf("H%d", row), a.Brand)
		_ = f.SetCellValue(sheet, fmt.Sprintf("I%d", row), a.Model)
		_ = f.SetCellValue(sheet, fmt.Sprintf("J%d", row), a.Serial)
		_ = f.SetCellValue(sheet, fmt.Sprintf("K%d", row), a.Status)
		_ = f.SetCellValue(sheet, fmt.Sprintf("L%d", row), a.State)
		if a.PurchasedDate != nil && !a.PurchasedDate.IsZero() {
			_ = f.SetCellValue(sheet, fmt.Sprintf("M%d", row), a.PurchasedDate.Format("2006-01-02"))
		}
		_ = f.SetCellValue(sheet, fmt.Sprintf("N%d", row), a.Notes)
	}

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// ImportExcel processes an uploaded .xlsx file
func (s *AssetService) ImportExcel(ctx context.Context, reader io.Reader, userSession string) (*dtos.ImportResultDto, error) {
	f, err := excelize.OpenReader(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to open excel reader: %w", err)
	}
	defer f.Close()

	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, errors.New("excel file contains no sheets")
	}

	rows, err := f.GetRows(sheets[0])
	if err != nil {
		return nil, err
	}

	if len(rows) <= 1 {
		return nil, errors.New("excel file is empty")
	}

	importedCount := 0
	failedCount := 0

	for i := 1; i < len(rows); i++ {
		row := rows[i]
		if len(row) < 2 || strings.TrimSpace(row[0]) == "" {
			continue
		}

		unitCode := strings.TrimSpace(row[0])
		assetType := ""
		if len(row) > 1 {
			assetType = strings.TrimSpace(row[1])
		}

		asset := domains.ASSET{
			UnitCode: unitCode,
			Type:     assetType,
			Status:   "In-use",
		}

		if len(row) > 2 {
			asset.UserName = strings.TrimSpace(row[2])
		}
		if len(row) > 3 {
			asset.Department = strings.TrimSpace(row[3])
		}

		if err := s.assetRepo.DB().WithContext(ctx).Save(&asset).Error; err != nil {
			logger.Errorf("Failed to import asset %s: %v", unitCode, err)
			failedCount++
		} else {
			importedCount++
		}
	}

	return &dtos.ImportResultDto{
		ImportedCount: importedCount,
		FailedCount:   failedCount,
		Message:       fmt.Sprintf("Successfully imported %d assets (%d failed)", importedCount, failedCount),
	}, nil
}

func (s *AssetService) mapToDto(e *domains.ASSET) dtos.AssetDto {
	return dtos.AssetDto{
		UnitCode:           e.UnitCode,
		Type:               e.Type,
		UserName:           e.UserName,
		Detail1:            e.Detail1,
		Detail2:            e.Detail2,
		Detail3:            e.Detail3,
		Detail4:            e.Detail4,
		Detail5:            e.Detail5,
		Detail6:            e.Detail6,
		Detail7:            e.Detail7,
		Detail8:            e.Detail8,
		Detail9:            e.Detail9,
		Model:              e.Model,
		IDTag:              e.IDTag,
		PurchasedDate:      e.PurchasedDate,
		UsageDate:          e.UsageDate,
		Brand:              e.Brand,
		WarrantyExpiration: e.WarrantyExpiration,
		PurchasedValue:     e.PurchasedValue,
		Vendor:             e.Vendor,
		TypeEXP:            e.TypeEXP,
		Notes:              e.Notes,
		Department:         e.Department,
		Office:             e.Office,
		Area:               e.Area,
		Company:            e.Company,
		Status:             e.Status,
		Serial:             e.Serial,
		RecordDate:         e.RecordDate,
		HRCode:             e.HRCode,
		Depreciation:       e.Depreciation,
		Value:              e.Value,
		Accessories:        e.Accessories,
		MainAsset:          e.MainAsset,
		Email:              e.Email,
		State:              e.State,
		Watt:               e.Watt,
		GroupDep:           e.GroupDep,
		Division:           e.Division,
		DepartmentName:     e.DepartmentName,
		OfficeName:         e.OfficeName,
		CompanyName:        e.CompanyName,
		VendorName:         e.VendorName,
		StateName:          e.StateName,
		DivisionName:       e.DivisionName,
	}
}

func (s *AssetService) mapToEntity(d *dtos.AssetDto) *domains.ASSET {
	return &domains.ASSET{
		UnitCode:           d.UnitCode,
		Type:               d.Type,
		UserName:           d.UserName,
		Detail1:            d.Detail1,
		Detail2:            d.Detail2,
		Detail3:            d.Detail3,
		Detail4:            d.Detail4,
		Detail5:            d.Detail5,
		Detail6:            d.Detail6,
		Detail7:            d.Detail7,
		Detail8:            d.Detail8,
		Detail9:            d.Detail9,
		Model:              d.Model,
		IDTag:              d.IDTag,
		PurchasedDate:      d.PurchasedDate,
		UsageDate:          d.UsageDate,
		Brand:              d.Brand,
		WarrantyExpiration: d.WarrantyExpiration,
		PurchasedValue:     d.PurchasedValue,
		Vendor:             d.Vendor,
		TypeEXP:            d.TypeEXP,
		Notes:              d.Notes,
		Department:         d.Department,
		Office:             d.Office,
		Area:               d.Area,
		Company:            d.Company,
		Status:             d.Status,
		Serial:             d.Serial,
		RecordDate:         d.RecordDate,
		HRCode:             d.HRCode,
		Depreciation:       d.Depreciation,
		Value:              d.Value,
		Accessories:        d.Accessories,
		MainAsset:          d.MainAsset,
		Email:              d.Email,
		State:              d.State,
		Watt:               d.Watt,
		GroupDep:           d.GroupDep,
		Division:           d.Division,
	}
}
