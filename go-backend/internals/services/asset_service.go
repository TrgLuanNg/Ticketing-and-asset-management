package services

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"go-backend/internals/core/domains"
	"go-backend/internals/core/dtos"
	"go-backend/internals/core/ports"
	"go-backend/pkg/logger"
)

type AssetService struct {
	assetRepo      ports.AssetRepository
	userRepo       ports.UserRepository
	authUseCase    ports.AuthUseCase
	txManager      ports.TransactionManager
	excelProcessor ports.ExcelProcessor
}

func NewAssetService(
	assetRepo ports.AssetRepository,
	userRepo ports.UserRepository,
	authUseCase ports.AuthUseCase,
	txManager ports.TransactionManager,
	excelProcessor ports.ExcelProcessor,
) *AssetService {
	return &AssetService{
		assetRepo:      assetRepo,
		userRepo:       userRepo,
		authUseCase:    authUseCase,
		txManager:      txManager,
		excelProcessor: excelProcessor,
	}
}

func (s *AssetService) SearchGrid(ctx context.Context, filter *dtos.Filter, userSession string) ([]dtos.AssetDto, int64, error) {
	allowedAreas, seeAll := s.authUseCase.GetUserAllowedAreas(ctx, userSession)
	entities, total, err := s.assetRepo.FindPagedWithAreas(ctx, filter, allowedAreas, seeAll, userSession)
	if err != nil {
		return nil, 0, err
	}

	dtosList := make([]dtos.AssetDto, len(entities))
	for i := range entities {
		dtosList[i] = s.mapToDto(&entities[i])
	}

	return dtosList, total, nil
}

func (s *AssetService) GetAll(ctx context.Context, userSession string) ([]dtos.AssetDto, error) {
	allowedAreas, seeAll := s.authUseCase.GetUserAllowedAreas(ctx, userSession)
	entities, err := s.assetRepo.FindAllWithAreas(ctx, allowedAreas, seeAll, userSession)
	if err != nil {
		return nil, err
	}

	dtosList := make([]dtos.AssetDto, len(entities))
	for i := range entities {
		dtosList[i] = s.mapToDto(&entities[i])
	}
	return dtosList, nil
}

func (s *AssetService) GetByID(ctx context.Context, userSession, unitCode string) (*dtos.AssetDto, error) {
	asset, err := s.assetRepo.FindByID(ctx, unitCode)
	if err != nil || asset == nil {
		return nil, errors.New("asset not found")
	}

	if strings.TrimSpace(userSession) != "" {
		allowedAreas, seeAll := s.authUseCase.GetUserAllowedAreas(ctx, userSession)
		if !s.isAssetAllowed(asset, allowedAreas, seeAll, userSession) {
			return nil, errors.New("you do not have permission to view this asset")
		}
	}

	dto := s.mapToDto(asset)

	subAssets, _ := s.assetRepo.FindSubAssets(ctx, unitCode)
	for i := range subAssets {
		dto.SubAssets = append(dto.SubAssets, s.mapToDto(&subAssets[i]))
	}

	return &dto, nil
}

func (s *AssetService) isAssetAllowed(asset *domains.ASSET, allowedAreas []string, seeAll bool, userSession string) bool {
	if len(allowedAreas) > 0 {
		hasAll := false
		matchedArea := false
		for _, a := range allowedAreas {
			if strings.EqualFold(a, "ALL") || strings.EqualFold(a, "VN") {
				hasAll = true
				break
			}
			if strings.EqualFold(a, strings.TrimSpace(asset.Area)) {
				matchedArea = true
			}
		}
		if !hasAll && !matchedArea {
			return false
		}
	}
	if !seeAll && strings.TrimSpace(userSession) != "" {
		if !strings.EqualFold(strings.TrimSpace(asset.HRCode), strings.TrimSpace(userSession)) {
			return false
		}
	}
	return true
}

func (s *AssetService) Insert(ctx context.Context, userSession string, dto *dtos.AssetDto) (*dtos.AssetDto, error) {
	if !s.HasModifyPermission(ctx, userSession) {
		return nil, errors.New("you do not have permission to modify assets")
	}

	entity := s.mapToEntity(dto)
	entity.MainAsset = ""
	now := domains.NowDateOnly()
	if entity.RecordDate == nil {
		entity.RecordDate = &now
	}
	allocated := make(map[string]bool)

	err := s.txManager.RunInTransaction(ctx, func(txCtx context.Context) error {
		if entity.UnitCode == "" || s.assetRepo.UnitCodeExists(txCtx, entity.UnitCode) {
			purchasedTime := time.Now()
			if entity.PurchasedDate != nil && !entity.PurchasedDate.IsZero() {
				purchasedTime = entity.PurchasedDate.Time
			}
			entity.UnitCode = s.generateUnitCode(txCtx, entity.Type, purchasedTime, entity.Company, entity.Area, allocated)
		}
		allocated[entity.UnitCode] = true

		if err := s.assetRepo.Create(txCtx, entity); err != nil {
			return err
		}

		nowTime := time.Now()
		_ = s.assetRepo.AddAssetLog(txCtx, &domains.ASSETSLOG{
			UnitCode:     entity.UnitCode,
			UserID:       userSession,
			DateModified: nowTime,
			Action:       "Create",
			ActionLogs:   fmt.Sprintf("UnitCode : %s", entity.UnitCode),
		})

		for _, subDto := range dto.SubAssets {
			subEntity := s.mapToEntity(&subDto)
			subEntity.MainAsset = entity.UnitCode
			subEntity.UserName = entity.UserName
			subEntity.Department = entity.Department
			subEntity.GroupDep = entity.GroupDep
			subEntity.Division = entity.Division
			subEntity.Office = entity.Office
			subEntity.Company = entity.Company
			subEntity.Area = entity.Area
			subEntity.Email = entity.Email
			subEntity.HRCode = entity.HRCode
			subEntity.State = entity.State
			subEntity.PurchasedDate = entity.PurchasedDate
			subEntity.Vendor = entity.Vendor
			subEntity.TypeEXP = entity.TypeEXP
			if subEntity.RecordDate == nil {
				subEntity.RecordDate = &now
			}

			if subEntity.UnitCode == "" || allocated[subEntity.UnitCode] || s.assetRepo.UnitCodeExists(txCtx, subEntity.UnitCode) {
				subPurchasedTime := time.Now()
				if subEntity.PurchasedDate != nil && !subEntity.PurchasedDate.IsZero() {
					subPurchasedTime = subEntity.PurchasedDate.Time
				}
				subEntity.UnitCode = s.generateUnitCode(txCtx, subEntity.Type, subPurchasedTime, subEntity.Company, subEntity.Area, allocated)
			}
			allocated[subEntity.UnitCode] = true

			if err := s.assetRepo.Create(txCtx, subEntity); err != nil {
				return err
			}

			_ = s.assetRepo.AddAssetLog(txCtx, &domains.ASSETSLOG{
				UnitCode:     subEntity.UnitCode,
				UserID:       userSession,
				DateModified: nowTime,
				Action:       "Create",
				ActionLogs:   fmt.Sprintf("UnitCode : %s", subEntity.UnitCode),
			})
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	s.userRepo.AddHistory(ctx, userSession, fmt.Sprintf("Add ASSET: %s", entity.UnitCode))
	createdDto := s.mapToDto(entity)
	return &createdDto, nil
}

func (s *AssetService) Update(ctx context.Context, userSession string, unitCode string, dto *dtos.AssetDto) error {
	if !s.HasModifyPermission(ctx, userSession) {
		return errors.New("you do not have permission to modify assets")
	}

	oldAsset, err := s.assetRepo.FindByID(ctx, unitCode)
	if err != nil || oldAsset == nil {
		return errors.New("asset not found")
	}

	entity := s.mapToEntity(dto)
	if strings.TrimSpace(entity.UnitCode) == "" {
		entity.UnitCode = unitCode
	}

	changeDetails := s.buildChangeDetails(ctx, oldAsset, entity)

	err = s.txManager.RunInTransaction(ctx, func(txCtx context.Context) error {
		if !strings.EqualFold(unitCode, entity.UnitCode) {
			if s.assetRepo.UnitCodeExists(txCtx, entity.UnitCode) {
				return fmt.Errorf("asset with UnitCode %s already exists", entity.UnitCode)
			}
			if err := s.assetRepo.Create(txCtx, entity); err != nil {
				return err
			}
			if err := s.assetRepo.Delete(txCtx, unitCode); err != nil {
				return err
			}
		} else {
			if err := s.assetRepo.Update(txCtx, unitCode, entity); err != nil {
				return err
			}
		}

		if changeDetails != "" {
			nowTime := time.Now()
			_ = s.assetRepo.AddAssetLog(txCtx, &domains.ASSETSLOG{
				UnitCode:     entity.UnitCode,
				UserID:       userSession,
				DateModified: nowTime,
				Action:       "Update",
				ActionLogs:   changeDetails,
			})
		}
		return nil
	})

	if err == nil {
		s.userRepo.AddHistory(ctx, userSession, fmt.Sprintf("Update ASSET: %s", entity.UnitCode))
	}
	return err
}

func (s *AssetService) buildChangeDetails(ctx context.Context, oldAsset, newAsset *domains.ASSET) string {
	depts, _ := s.assetRepo.GetAllDepartments(ctx)
	deptMap := make(map[string]string, len(depts))
	for _, d := range depts {
		deptMap[d.ID] = d.Name
	}

	states, _ := s.assetRepo.GetAllStates(ctx)
	stateMap := make(map[string]string, len(states))
	for _, st := range states {
		stateMap[st.ID] = st.State
	}

	configs, _ := s.assetRepo.GetAllDeviceConfigs(ctx)
	var devCfg *domains.DEVICECONFIG
	for i := range configs {
		if strings.EqualFold(configs[i].Device, newAsset.Type) {
			devCfg = &configs[i]
			break
		}
	}

	items, _ := s.assetRepo.GetAllItemMasterData(ctx)
	itemMap := make(map[string]string, len(items))
	for _, it := range items {
		itemMap[it.ID] = it.TechSpecs
	}

	resolveDept := func(id string) string {
		if v, ok := deptMap[id]; ok && v != "" {
			return v
		}
		return id
	}
	resolveState := func(id string) string {
		if v, ok := stateMap[id]; ok && v != "" {
			return v
		}
		return id
	}
	resolveItem := func(id string) string {
		if v, ok := itemMap[id]; ok && v != "" {
			return v
		}
		return id
	}
	formatDate := func(d *domains.DateOnly) string {
		if d == nil || d.IsZero() {
			return ""
		}
		return d.Format("2006-01-02")
	}

	var changes []string
	addDiff := func(label, oldVal, newVal string) {
		if strings.TrimSpace(oldVal) != strings.TrimSpace(newVal) {
			changes = append(changes, fmt.Sprintf("%s: %s -> %s", label, oldVal, newVal))
		}
	}

	addDiff("UnitCode", oldAsset.UnitCode, newAsset.UnitCode)
	addDiff("Type", oldAsset.Type, newAsset.Type)
	addDiff("UserName", oldAsset.UserName, newAsset.UserName)
	addDiff("HRCode", oldAsset.HRCode, newAsset.HRCode)
	addDiff("Email", oldAsset.Email, newAsset.Email)
	addDiff("Brand", oldAsset.Brand, newAsset.Brand)
	addDiff("Model", oldAsset.Model, newAsset.Model)
	addDiff("Serial", oldAsset.Serial, newAsset.Serial)
	addDiff("IDTag", oldAsset.IDTag, newAsset.IDTag)
	addDiff("Status", oldAsset.Status, newAsset.Status)
	addDiff("State", resolveState(oldAsset.State), resolveState(newAsset.State))
	addDiff("Company", resolveDept(oldAsset.Company), resolveDept(newAsset.Company))
	addDiff("Office", resolveDept(oldAsset.Office), resolveDept(newAsset.Office))
	addDiff("Department", resolveDept(oldAsset.Department), resolveDept(newAsset.Department))
	addDiff("GroupDep", resolveDept(oldAsset.GroupDep), resolveDept(newAsset.GroupDep))
	addDiff("Division", resolveDept(oldAsset.Division), resolveDept(newAsset.Division))
	addDiff("Vendor", resolveDept(oldAsset.Vendor), resolveDept(newAsset.Vendor))
	addDiff("Area", oldAsset.Area, newAsset.Area)
	addDiff("PurchasedDate", formatDate(oldAsset.PurchasedDate), formatDate(newAsset.PurchasedDate))
	addDiff("UsageDate", formatDate(oldAsset.UsageDate), formatDate(newAsset.UsageDate))
	addDiff("Notes", oldAsset.Notes, newAsset.Notes)
	addDiff("MainAsset", oldAsset.MainAsset, newAsset.MainAsset)

	detailLabel := func(idx int, fallback string) string {
		if devCfg == nil {
			return fallback
		}
		labels := []string{devCfg.Detail1, devCfg.Detail2, devCfg.Detail3, devCfg.Detail4, devCfg.Detail5, devCfg.Detail6, devCfg.Detail7, devCfg.Detail8, devCfg.Detail9}
		if idx >= 0 && idx < len(labels) && strings.TrimSpace(labels[idx]) != "" {
			return labels[idx]
		}
		return fallback
	}

	addDiff(detailLabel(0, "Detail1"), resolveItem(oldAsset.Detail1), resolveItem(newAsset.Detail1))
	addDiff(detailLabel(1, "Detail2"), resolveItem(oldAsset.Detail2), resolveItem(newAsset.Detail2))
	addDiff(detailLabel(2, "Detail3"), resolveItem(oldAsset.Detail3), resolveItem(newAsset.Detail3))
	addDiff(detailLabel(3, "Detail4"), resolveItem(oldAsset.Detail4), resolveItem(newAsset.Detail4))
	addDiff(detailLabel(4, "Detail5"), resolveItem(oldAsset.Detail5), resolveItem(newAsset.Detail5))
	addDiff(detailLabel(5, "Detail6"), resolveItem(oldAsset.Detail6), resolveItem(newAsset.Detail6))
	addDiff(detailLabel(6, "Detail7"), resolveItem(oldAsset.Detail7), resolveItem(newAsset.Detail7))
	addDiff(detailLabel(7, "Detail8"), resolveItem(oldAsset.Detail8), resolveItem(newAsset.Detail8))
	addDiff(detailLabel(8, "Detail9"), resolveItem(oldAsset.Detail9), resolveItem(newAsset.Detail9))

	return strings.Join(changes, " ; ")
}

func (s *AssetService) Delete(ctx context.Context, userSession string, unitCode string) error {
	if !s.HasDeletePermission(ctx, userSession) {
		return errors.New("you do not have permission to delete assets")
	}

	err := s.assetRepo.Delete(ctx, unitCode)
	if err == nil {
		s.userRepo.AddHistory(ctx, userSession, fmt.Sprintf("Delete ASSET: %s", unitCode))
	}
	return err
}

func (s *AssetService) GenerateUnitCode(ctx context.Context, assetType string, purchasedDate time.Time, company, area string) string {
	return s.generateUnitCode(ctx, assetType, purchasedDate, company, area, nil)
}

func (s *AssetService) generatePrefix(assetType string, purchasedDate time.Time, company, area string) string {
	id1 := "YLVN"
	if company == "00060" || strings.EqualFold(company, "NYKV") {
		id1 = "NYKV"
	}
	id2 := "N"
	if strings.EqualFold(area, "South") || strings.EqualFold(area, "S") {
		id2 = "S"
	}
	typeTrimmed := strings.TrimSpace(assetType)
	id3 := "XXX"
	if len(typeTrimmed) >= 3 {
		id3 = typeTrimmed[:3]
	} else if len(typeTrimmed) > 0 {
		id3 = typeTrimmed + strings.Repeat("X", 3-len(typeTrimmed))
	}
	if purchasedDate.IsZero() {
		purchasedDate = time.Now()
	}
	id4 := purchasedDate.Format("0601")
	return strings.ToUpper(fmt.Sprintf("%s%s%s%s", id1, id2, id3, id4))
}

func (s *AssetService) generateUnitCode(ctx context.Context, assetType string, purchasedDate time.Time, company, area string, allocated map[string]bool) string {
	prefix := s.generatePrefix(assetType, purchasedDate, company, area)

	matchingCodes, _ := s.assetRepo.GetMatchingUnitCodes(ctx, prefix)

	maxSuffix := 0
	for _, code := range matchingCodes {
		if len(code) > len(prefix) {
			suffixStr := code[len(prefix):]
			if num, err := strconv.Atoi(suffixStr); err == nil {
				if num > maxSuffix {
					maxSuffix = num
				}
			}
		}
	}

	suffix := maxSuffix + 1
	candidate := fmt.Sprintf("%s%03d", prefix, suffix)
	for (allocated != nil && allocated[candidate]) || s.assetRepo.UnitCodeExists(ctx, candidate) {
		suffix++
		candidate = fmt.Sprintf("%s%03d", prefix, suffix)
	}
	if allocated != nil {
		allocated[candidate] = true
	}
	return candidate
}

func (s *AssetService) GetDashboardMetrics(ctx context.Context, userSession, fromDate string) (*dtos.DashboardMetricsDto, error) {
	allowedAreas, seeAll := s.authUseCase.GetUserAllowedAreas(ctx, userSession)
	return s.assetRepo.GetDashboardMetrics(ctx, allowedAreas, seeAll, userSession, fromDate)
}

func (s *AssetService) GetSuggestions(ctx context.Context, filter *dtos.Filter, userSession string) ([]dtos.AssetSuggestionDto, error) {
	allowedAreas, seeAll := s.authUseCase.GetUserAllowedAreas(ctx, userSession)
	return s.assetRepo.GetSuggestions(ctx, filter, allowedAreas, seeAll, userSession)
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

func (s *AssetService) ExportExcel(ctx context.Context, filter *dtos.Filter, userSession string) ([]byte, error) {
	if s.excelProcessor == nil {
		return nil, errors.New("excel processor not configured")
	}

	allowedAreas, seeAll := s.authUseCase.GetUserAllowedAreas(ctx, userSession)
	var filterCopy dtos.Filter
	if filter != nil {
		filterCopy = *filter
	}
	filterCopy.PageIndex = 1
	filterCopy.PageSize = 100000
	filterCopy.SkipCount = true

	assets, _, err := s.assetRepo.FindPagedWithAreas(ctx, &filterCopy, allowedAreas, seeAll, userSession)
	if err != nil {
		return nil, err
	}

	depts, _ := s.assetRepo.GetAllDepartments(ctx)
	deptMap := make(map[string]string, len(depts))
	for _, d := range depts {
		deptMap[d.ID] = d.Name
	}

	configs, _ := s.assetRepo.GetAllDeviceConfigs(ctx)
	deviceConfigMap := make(map[string]domains.DEVICECONFIG, len(configs))
	for _, c := range configs {
		deviceConfigMap[strings.ToUpper(strings.TrimSpace(c.Device))] = c
	}

	items, _ := s.assetRepo.GetAllItemMasterData(ctx)
	itemMap := make(map[string]string, len(items))
	for _, it := range items {
		itemMap[it.ID] = it.TechSpecs
	}

	return s.excelProcessor.ExportAssetsExcel(assets, deptMap, deviceConfigMap, itemMap)
}

func (s *AssetService) ImportExcel(ctx context.Context, reader io.Reader, userSession string) (*dtos.ImportResultDto, error) {
	if !s.HasModifyPermission(ctx, userSession) {
		return nil, errors.New("you do not have permission to import assets")
	}
	if s.excelProcessor == nil {
		return nil, errors.New("excel processor not configured")
	}

	rows, err := s.excelProcessor.ParseAssetsImportExcel(reader)
	if err != nil {
		return nil, err
	}

	depts, _ := s.assetRepo.GetAllDepartments(ctx)
	deptByName := make(map[string]domains.DEPARTMENT, len(depts))
	deptByID := make(map[string]domains.DEPARTMENT, len(depts))
	for _, d := range depts {
		deptByID[d.ID] = d
		deptByName[strings.ToUpper(strings.TrimSpace(d.Name))] = d
	}

	states, _ := s.assetRepo.GetAllStates(ctx)
	stateByName := make(map[string]string, len(states))
	for _, st := range states {
		stateByName[strings.ToUpper(strings.TrimSpace(st.State))] = st.ID
		stateByName[strings.ToUpper(strings.TrimSpace(st.ID))] = st.ID
	}

	configs, _ := s.assetRepo.GetAllDeviceConfigs(ctx)
	configByDevice := make(map[string]domains.DEVICECONFIG, len(configs))
	for _, c := range configs {
		configByDevice[strings.ToUpper(strings.TrimSpace(c.Device))] = c
	}

	items, _ := s.assetRepo.GetAllItemMasterData(ctx)

	allocatedCodes := make(map[string]bool)
	seenSerials := make(map[string]bool)

	var failedRows []ports.AssetImportErrorRow
	importedCount := 0
	now := domains.NowDateOnly()
	nowTime := time.Now()

	resolveDeptID := func(val string) (domains.DEPARTMENT, bool) {
		trimmed := strings.TrimSpace(val)
		if trimmed == "" {
			return domains.DEPARTMENT{}, false
		}
		if d, ok := deptByID[trimmed]; ok {
			return d, true
		}
		if d, ok := deptByName[strings.ToUpper(trimmed)]; ok {
			return d, true
		}
		return domains.DEPARTMENT{}, false
	}

	for _, r := range rows {
		var rowErrs []string

		if r.Type == "" {
			rowErrs = append(rowErrs, "Type is required")
		}
		if r.Brand == "" {
			rowErrs = append(rowErrs, "Brand is required")
		}
		if r.Model == "" {
			rowErrs = append(rowErrs, "Model is required")
		}
		if r.Serial == "" {
			rowErrs = append(rowErrs, "Serial is required")
		} else {
			serialUpper := strings.ToUpper(r.Serial)
			if seenSerials[serialUpper] || s.assetRepo.SerialExists(ctx, r.Serial) {
				rowErrs = append(rowErrs, fmt.Sprintf("Duplicate Serial '%s'", r.Serial))
			}
		}
		if r.PurchasedDate == nil || r.PurchasedDate.IsZero() {
			rowErrs = append(rowErrs, "Valid Purchased Date is required")
		}
		if r.Status == "" {
			rowErrs = append(rowErrs, "Status is required")
		}
		if r.Company == "" {
			rowErrs = append(rowErrs, "Company is required")
		}
		if r.Office == "" {
			rowErrs = append(rowErrs, "Office is required")
		}
		if r.Area == "" {
			rowErrs = append(rowErrs, "Area is required")
		}

		companyID := r.Company
		if d, ok := resolveDeptID(r.Company); ok {
			companyID = d.ID
		}
		officeID := r.Office
		if d, ok := resolveDeptID(r.Office); ok {
			officeID = d.ID
		}
		vendorID := r.Vendor
		if d, ok := resolveDeptID(r.Vendor); ok {
			vendorID = d.ID
		}

		deptID := r.Department
		divID := r.Division
		groupDepID := r.GroupDep
		if d, ok := resolveDeptID(r.Department); ok {
			deptID = d.ID
			if divID == "" && d.ParentId != "" {
				if parentDiv, okDiv := deptByID[d.ParentId]; okDiv {
					divID = parentDiv.ID
					if groupDepID == "" && parentDiv.ParentId != "" {
						groupDepID = parentDiv.ParentId
					}
				}
			}
		}
		if divID != "" {
			if d, ok := resolveDeptID(divID); ok {
				divID = d.ID
				if groupDepID == "" && d.ParentId != "" {
					groupDepID = d.ParentId
				}
			}
		}
		if groupDepID != "" {
			if d, ok := resolveDeptID(groupDepID); ok {
				groupDepID = d.ID
			}
		}

		stateID := "00001"
		if r.State != "" {
			if stID, ok := stateByName[strings.ToUpper(strings.TrimSpace(r.State))]; ok {
				stateID = stID
			} else {
				stateID = r.State
			}
		}

		if len(rowErrs) > 0 {
			failedRows = append(failedRows, ports.AssetImportErrorRow{
				Row:         r,
				ErrorReason: strings.Join(rowErrs, "; "),
			})
			continue
		}

		details := s.parseTechSpec(r.Type, r.Configuration, configByDevice, items)

		pDate := r.PurchasedDate.Time
		unitCode := s.generateUnitCode(ctx, r.Type, pDate, companyID, r.Area, allocatedCodes)

		asset := domains.ASSET{
			UnitCode:      unitCode,
			Type:          r.Type,
			Brand:         r.Brand,
			Model:         r.Model,
			Serial:        r.Serial,
			PurchasedDate: r.PurchasedDate,
			RecordDate:    &now,
			Vendor:        vendorID,
			Status:        r.Status,
			UserName:      r.UserName,
			HRCode:        r.HRCode,
			Department:    deptID,
			Division:      divID,
			GroupDep:      groupDepID,
			Office:        officeID,
			Company:       companyID,
			Area:          r.Area,
			IDTag:         r.IDTag,
			Notes:         r.Notes,
			State:         stateID,
			Detail1:       details[0],
			Detail2:       details[1],
			Detail3:       details[2],
			Detail4:       details[3],
			Detail5:       details[4],
			Detail6:       details[5],
			Detail7:       details[6],
			Detail8:       details[7],
			Detail9:       details[8],
		}

		if err := s.assetRepo.Create(ctx, &asset); err != nil {
			logger.Errorf("Failed to import asset row %d: %v", r.RowNumber, err)
			failedRows = append(failedRows, ports.AssetImportErrorRow{
				Row:         r,
				ErrorReason: err.Error(),
			})
			continue
		}

		seenSerials[strings.ToUpper(r.Serial)] = true
		importedCount++

		_ = s.assetRepo.AddAssetLog(ctx, &domains.ASSETSLOG{
			UnitCode:     asset.UnitCode,
			UserID:       userSession,
			DateModified: nowTime,
			Action:       "Create",
			ActionLogs:   fmt.Sprintf("UnitCode : %s (Excel Import)", asset.UnitCode),
		})
	}

	var errBytes []byte
	if len(failedRows) > 0 {
		errBytes, _ = s.excelProcessor.GenerateAssetsErrorExcel(failedRows)
	}

	if importedCount > 0 {
		s.userRepo.AddHistory(ctx, userSession, fmt.Sprintf("Imported %d assets from Excel", importedCount))
	}

	return &dtos.ImportResultDto{
		ImportedCount:  importedCount,
		FailedCount:    len(failedRows),
		Message:        fmt.Sprintf("Successfully imported %d assets (%d failed)", importedCount, len(failedRows)),
		ErrorFileBytes: errBytes,
	}, nil
}

func (s *AssetService) parseTechSpec(
	assetType, configStr string,
	configByDevice map[string]domains.DEVICECONFIG,
	items []domains.ITEMMASTERDATA,
) [9]string {
	var res [9]string
	if strings.TrimSpace(configStr) == "" {
		return res
	}

	devCfg, ok := configByDevice[strings.ToUpper(strings.TrimSpace(assetType))]
	if !ok {
		return res
	}

	slotKeys := [9]string{
		devCfg.Detail1, devCfg.Detail2, devCfg.Detail3,
		devCfg.Detail4, devCfg.Detail5, devCfg.Detail6,
		devCfg.Detail7, devCfg.Detail8, devCfg.Detail9,
	}

	segments := strings.FieldsFunc(configStr, func(r rune) bool {
		return r == '\n' || r == '\r' || r == ';'
	})
	if len(segments) == 1 && strings.Contains(configStr, " / ") {
		segments = strings.Split(configStr, " / ")
	}

	findItemID := func(slotKey, val string) string {
		valTrim := strings.TrimSpace(val)
		if valTrim == "" {
			return ""
		}
		for _, it := range items {
			if strings.EqualFold(strings.TrimSpace(it.Item), strings.TrimSpace(slotKey)) &&
				strings.EqualFold(strings.TrimSpace(it.TechSpecs), valTrim) {
				return it.ID
			}
		}
		for _, it := range items {
			if strings.EqualFold(strings.TrimSpace(it.TechSpecs), valTrim) {
				return it.ID
			}
		}
		return valTrim
	}

	for _, seg := range segments {
		seg = strings.TrimSpace(seg)
		if seg == "" {
			continue
		}
		if colonIdx := strings.Index(seg, ":"); colonIdx > 0 {
			k := strings.TrimSpace(seg[:colonIdx])
			v := strings.TrimSpace(seg[colonIdx+1:])
			for idx, slotKey := range slotKeys {
				if slotKey != "" && strings.EqualFold(slotKey, k) {
					res[idx] = findItemID(slotKey, v)
					break
				}
			}
		}
	}

	return res
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
		GroupDepName:       e.GroupDepName,
		AD:                 e.AD,
		ME:                 e.ME,
		TM:                 e.TM,
		AdOS:               e.AdOS,
		MeTime:             e.MeTime,
		TmTime:             e.TmTime,
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
