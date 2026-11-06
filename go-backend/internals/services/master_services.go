package services

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"go-backend/internals/core/domains"
	"go-backend/internals/core/dtos"
	"go-backend/internals/core/ports"
)

// GenericService implements ports.GenericUseCase[T] for standard master data entities.
type GenericService[T any] struct {
	repo       ports.GenericRepository[T]
	userRepo   ports.UserRepository
	entityName string
}

func NewGenericService[T any](repo ports.GenericRepository[T], userRepo ports.UserRepository, entityName string) *GenericService[T] {
	return &GenericService[T]{
		repo:       repo,
		userRepo:   userRepo,
		entityName: entityName,
	}
}

func (s *GenericService[T]) Search(ctx context.Context, filter *dtos.Filter) ([]T, int64, error) {
	return s.repo.FindPaged(ctx, filter)
}

func (s *GenericService[T]) GetAll(ctx context.Context) ([]T, error) {
	return s.repo.FindAll(ctx)
}

func (s *GenericService[T]) GetByID(ctx context.Context, id any) (*T, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *GenericService[T]) Create(ctx context.Context, userSession string, entity *T) error {
	err := s.repo.Create(ctx, entity)
	if err == nil && s.userRepo != nil && userSession != "" {
		s.userRepo.AddHistory(ctx, userSession, fmt.Sprintf("Add %s", s.entityName))
	}
	return err
}

func (s *GenericService[T]) Update(ctx context.Context, userSession string, id any, entity *T) error {
	err := s.repo.Update(ctx, id, entity)
	if err == nil && s.userRepo != nil && userSession != "" {
		s.userRepo.AddHistory(ctx, userSession, fmt.Sprintf("Update %s: %v", s.entityName, id))
	}
	return err
}

func (s *GenericService[T]) Delete(ctx context.Context, userSession string, id any) error {
	err := s.repo.Delete(ctx, id)
	if err == nil && s.userRepo != nil && userSession != "" {
		s.userRepo.AddHistory(ctx, userSession, fmt.Sprintf("Delete %s: %v", s.entityName, id))
	}
	return err
}

func hasPermissionFlag(ctx context.Context, userRepo ports.UserRepository, userSession string, selector func(*domains.PERMISSION) *bool) bool {
	if userRepo == nil {
		return true
	}
	user, _ := userRepo.FindByUserID(ctx, userSession)
	isAdmin := user != nil && strings.EqualFold(user.Role, "admin")
	perm, _ := userRepo.FindPermission(ctx, userSession)
	if perm == nil {
		return isAdmin
	}
	val := selector(perm)
	return (val != nil && *val) || (val == nil && isAdmin)
}

// DepartmentService enforces DepartmentModify/DepartmentDel permissions and auto-generates 5-digit IDs.
type DepartmentService struct {
	*GenericService[domains.DEPARTMENT]
	deptRepo ports.DepartmentRepository
}

func NewDepartmentService(deptRepo ports.DepartmentRepository, userRepo ports.UserRepository) *DepartmentService {
	return &DepartmentService{
		GenericService: NewGenericService[domains.DEPARTMENT](deptRepo, userRepo, "DEPARTMENT"),
		deptRepo:       deptRepo,
	}
}

func (s *DepartmentService) Create(ctx context.Context, userSession string, entity *domains.DEPARTMENT) error {
	if !hasPermissionFlag(ctx, s.userRepo, userSession, func(p *domains.PERMISSION) *bool { return p.DepartmentModify }) {
		return errors.New("you do not have permission to modify departments")
	}
	if strings.TrimSpace(entity.ID) == "" {
		nextID, err := s.deptRepo.GetNextID(ctx)
		if err != nil {
			return err
		}
		entity.ID = nextID
	}
	return s.GenericService.Create(ctx, userSession, entity)
}

func (s *DepartmentService) Update(ctx context.Context, userSession string, id any, entity *domains.DEPARTMENT) error {
	if !hasPermissionFlag(ctx, s.userRepo, userSession, func(p *domains.PERMISSION) *bool { return p.DepartmentModify }) {
		return errors.New("you do not have permission to modify departments")
	}
	return s.GenericService.Update(ctx, userSession, id, entity)
}

func (s *DepartmentService) Delete(ctx context.Context, userSession string, id any) error {
	if !hasPermissionFlag(ctx, s.userRepo, userSession, func(p *domains.PERMISSION) *bool { return p.DepartmentDel }) {
		return errors.New("you do not have permission to delete departments")
	}
	return s.GenericService.Delete(ctx, userSession, id)
}

// TypeService enforces TypeModify/TypeDel permissions.
type TypeService struct {
	*GenericService[domains.TYPE]
}

func NewTypeService(repo ports.GenericRepository[domains.TYPE], userRepo ports.UserRepository) *TypeService {
	return &TypeService{
		GenericService: NewGenericService[domains.TYPE](repo, userRepo, "TYPE"),
	}
}

func (s *TypeService) Create(ctx context.Context, userSession string, entity *domains.TYPE) error {
	if !hasPermissionFlag(ctx, s.userRepo, userSession, func(p *domains.PERMISSION) *bool { return p.TypeModify }) {
		return errors.New("you do not have permission to modify asset types")
	}
	return s.GenericService.Create(ctx, userSession, entity)
}

func (s *TypeService) Update(ctx context.Context, userSession string, id any, entity *domains.TYPE) error {
	if !hasPermissionFlag(ctx, s.userRepo, userSession, func(p *domains.PERMISSION) *bool { return p.TypeModify }) {
		return errors.New("you do not have permission to modify asset types")
	}
	return s.GenericService.Update(ctx, userSession, id, entity)
}

func (s *TypeService) Delete(ctx context.Context, userSession string, id any) error {
	if !hasPermissionFlag(ctx, s.userRepo, userSession, func(p *domains.PERMISSION) *bool { return p.TypeDel }) {
		return errors.New("you do not have permission to delete asset types")
	}
	return s.GenericService.Delete(ctx, userSession, id)
}

// PermissionService enforces PermissionApproval, supports upsert, and evicts area cache.
type PermissionService struct {
	*GenericService[domains.PERMISSION]
	cache ports.CacheProvider
}

func NewPermissionService(repo ports.GenericRepository[domains.PERMISSION], userRepo ports.UserRepository, cache ports.CacheProvider) *PermissionService {
	return &PermissionService{
		GenericService: NewGenericService[domains.PERMISSION](repo, userRepo, "PERMISSION"),
		cache:          cache,
	}
}

func (s *PermissionService) Create(ctx context.Context, userSession string, entity *domains.PERMISSION) error {
	if !hasPermissionFlag(ctx, s.userRepo, userSession, func(p *domains.PERMISSION) *bool { return p.PermissionApproval }) {
		return errors.New("you do not have permission to approve permissions")
	}
	err := s.userRepo.UpsertPermission(ctx, entity)
	if s.cache != nil {
		s.cache.Remove(fmt.Sprintf("user_areas_%s", entity.UserID))
	}
	if err == nil && s.userRepo != nil && userSession != "" {
		s.userRepo.AddHistory(ctx, userSession, fmt.Sprintf("Upsert PERMISSION: %s", entity.UserID))
	}
	return err
}

func (s *PermissionService) Update(ctx context.Context, userSession string, id any, entity *domains.PERMISSION) error {
	if !hasPermissionFlag(ctx, s.userRepo, userSession, func(p *domains.PERMISSION) *bool { return p.PermissionApproval }) {
		return errors.New("you do not have permission to approve permissions")
	}
	if strings.TrimSpace(entity.UserID) == "" {
		entity.UserID = fmt.Sprintf("%v", id)
	}
	err := s.userRepo.UpsertPermission(ctx, entity)
	if s.cache != nil {
		s.cache.Remove(fmt.Sprintf("user_areas_%s", entity.UserID))
	}
	if err == nil && s.userRepo != nil && userSession != "" {
		s.userRepo.AddHistory(ctx, userSession, fmt.Sprintf("Update PERMISSION: %s", entity.UserID))
	}
	return err
}

func (s *PermissionService) Delete(ctx context.Context, userSession string, id any) error {
	if !hasPermissionFlag(ctx, s.userRepo, userSession, func(p *domains.PERMISSION) *bool { return p.PermissionApproval }) {
		return errors.New("you do not have permission to delete permissions")
	}
	err := s.GenericService.Delete(ctx, userSession, id)
	if s.cache != nil {
		s.cache.Remove(fmt.Sprintf("user_areas_%v", id))
	}
	return err
}

// ItemMasterDataService auto-generates 4-digit zero-padded IDs on Create.
type ItemMasterDataService struct {
	*GenericService[domains.ITEMMASTERDATA]
	itemRepo ports.ItemMasterDataRepository
}

func NewItemMasterDataService(itemRepo ports.ItemMasterDataRepository, userRepo ports.UserRepository) *ItemMasterDataService {
	return &ItemMasterDataService{
		GenericService: NewGenericService[domains.ITEMMASTERDATA](itemRepo, userRepo, "ITEMMASTERDATA"),
		itemRepo:       itemRepo,
	}
}

func (s *ItemMasterDataService) Create(ctx context.Context, userSession string, entity *domains.ITEMMASTERDATA) error {
	if strings.TrimSpace(entity.ID) == "" {
		nextID, err := s.itemRepo.GetNextID(ctx)
		if err != nil {
			return err
		}
		entity.ID = nextID
	}
	return s.GenericService.Create(ctx, userSession, entity)
}

// DisposalService implements ports.DisposalUseCase with full batch WFD / confirm / check / import workflows.
type DisposalService struct {
	*GenericService[domains.DISPOSAL]
	disposalRepo   ports.DisposalRepository
	assetRepo      ports.AssetRepository
	authUseCase    ports.AuthUseCase
	txManager      ports.TransactionManager
	excelProcessor ports.ExcelProcessor
}

func NewDisposalService(
	disposalRepo ports.DisposalRepository,
	assetRepo ports.AssetRepository,
	userRepo ports.UserRepository,
	authUseCase ports.AuthUseCase,
	txManager ports.TransactionManager,
	excelProcessor ports.ExcelProcessor,
) *DisposalService {
	return &DisposalService{
		GenericService: NewGenericService[domains.DISPOSAL](disposalRepo, userRepo, "DISPOSAL"),
		disposalRepo:   disposalRepo,
		assetRepo:      assetRepo,
		authUseCase:    authUseCase,
		txManager:      txManager,
		excelProcessor: excelProcessor,
	}
}

func (s *DisposalService) hasDisposalPermission(ctx context.Context, userSession string) bool {
	return hasPermissionFlag(ctx, s.userRepo, userSession, func(p *domains.PERMISSION) *bool { return p.Disposal })
}

func (s *DisposalService) TransferToWFD(ctx context.Context, userSession string, req *dtos.TransferWfdRequestDto) (int, string, error) {
	if !s.hasDisposalPermission(ctx, userSession) {
		return 0, "", errors.New("you do not have permission to manage disposals")
	}
	if req == nil || len(req.UnitCodes) == 0 {
		return 0, "", errors.New("unitCodes list cannot be empty")
	}

	todayPrefix := fmt.Sprintf("DIS-%s-", time.Now().Format("20060102"))
	batchCnt, _ := s.disposalRepo.CountTodayBatches(ctx, todayPrefix)
	idDisposal := fmt.Sprintf("%s%02d", todayPrefix, batchCnt+1)
	today := domains.NowDateOnly()
	nowTime := time.Now()
	transferred := 0

	err := s.txManager.RunInTransaction(ctx, func(txCtx context.Context) error {
		for _, rawCode := range req.UnitCodes {
			code := strings.TrimSpace(rawCode)
			if code == "" {
				continue
			}
			asset, err := s.assetRepo.FindByID(txCtx, code)
			if err != nil || asset == nil {
				return fmt.Errorf("asset %s not found", code)
			}
			if strings.EqualFold(asset.Status, "In-use") || strings.EqualFold(asset.Status, "Borrowed") {
				return fmt.Errorf("asset %s is currently %s and cannot be transferred to WFD", code, asset.Status)
			}
			if s.disposalRepo.ExistsByUnitCode(txCtx, code) {
				return fmt.Errorf("asset %s is already in a disposal batch", code)
			}

			asset.Status = "Waiting for Disposal"
			if err := s.assetRepo.Update(txCtx, asset.UnitCode, asset); err != nil {
				return err
			}

			disp := &domains.DISPOSAL{
				IDDisposal: idDisposal,
				UnitCode:   asset.UnitCode,
				StartDate:  today,
				Status:     false,
				UserID:     userSession,
				Note:       req.Note,
			}
			if err := s.disposalRepo.Create(txCtx, disp); err != nil {
				return err
			}

			_ = s.assetRepo.AddAssetLog(txCtx, &domains.ASSETSLOG{
				UnitCode:     asset.UnitCode,
				UserID:       userSession,
				DateModified: nowTime,
				Action:       "Transfer to WFD",
				ActionLogs:   fmt.Sprintf("Batch: %s | Note: %s", idDisposal, req.Note),
			})
			transferred++
		}
		return nil
	})

	if err != nil {
		return 0, "", err
	}

	s.userRepo.AddHistory(ctx, userSession, fmt.Sprintf("Transferred %d assets to WFD (%s)", transferred, idDisposal))
	return transferred, idDisposal, nil
}

func (s *DisposalService) ConfirmDispose(ctx context.Context, userSession string, req *dtos.ConfirmDisposeRequestDto) (int, error) {
	if !s.hasDisposalPermission(ctx, userSession) {
		return 0, errors.New("you do not have permission to manage disposals")
	}
	if req == nil || strings.TrimSpace(req.IDDisposal) == "" {
		return 0, errors.New("idDisposal is required")
	}

	items, err := s.disposalRepo.FindByIDDisposal(ctx, req.IDDisposal)
	if err != nil || len(items) == 0 {
		return 0, errors.New("disposal batch not found")
	}

	dispDate := domains.NowDateOnly()
	nowTime := time.Now()
	confirmed := 0

	err = s.txManager.RunInTransaction(ctx, func(txCtx context.Context) error {
		for i := range items {
			item := items[i]
			if asset, errA := s.assetRepo.FindByID(txCtx, item.UnitCode); errA == nil && asset != nil {
				asset.Status = "Disposed"
				if errU := s.assetRepo.Update(txCtx, asset.UnitCode, asset); errU != nil {
					return errU
				}
			}

			item.EndDate = dispDate
			item.Status = true
			if strings.TrimSpace(req.Note) != "" {
				item.Note = req.Note
			}
			if errD := s.disposalRepo.UpdateDisposalItem(txCtx, &item); errD != nil {
				return errD
			}

			_ = s.assetRepo.AddAssetLog(txCtx, &domains.ASSETSLOG{
				UnitCode:     item.UnitCode,
				UserID:       userSession,
				DateModified: nowTime,
				Action:       "Confirm Dispose",
				ActionLogs:   fmt.Sprintf("Batch: %s | Note: %s", req.IDDisposal, item.Note),
			})
			confirmed++
		}
		return nil
	})

	if err != nil {
		return 0, err
	}

	s.userRepo.AddHistory(ctx, userSession, fmt.Sprintf("Confirmed disposal batch %s (%d assets)", req.IDDisposal, confirmed))
	return confirmed, nil
}

func (s *DisposalService) GetBatches(ctx context.Context, userSession string) ([]dtos.DisposalBatchDto, error) {
	allowedAreas, seeAll := s.authUseCase.GetUserAllowedAreas(ctx, userSession)
	allAssets, err := s.assetRepo.FindAllWithAreas(ctx, allowedAreas, seeAll, userSession)
	if err != nil {
		return nil, err
	}
	assetMap := make(map[string]domains.ASSET, len(allAssets))
	for _, a := range allAssets {
		assetMap[a.UnitCode] = a
	}

	allDisposals, err := s.disposalRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	type batchGroup struct {
		idDisposal  string
		startDate   domains.DateOnly
		endDate     domains.DateOnly
		userID      string
		note        string
		total       int
		disposedCnt int
	}

	groups := make(map[string]*batchGroup)
	for _, d := range allDisposals {
		asset, ok := assetMap[d.UnitCode]
		if !ok {
			continue
		}
		g, exists := groups[d.IDDisposal]
		if !exists {
			g = &batchGroup{
				idDisposal: d.IDDisposal,
				startDate:  d.StartDate,
				endDate:    d.EndDate,
				userID:     d.UserID,
				note:       d.Note,
			}
			groups[d.IDDisposal] = g
		}
		g.total++
		if d.Status || strings.EqualFold(asset.Status, "Disposed") {
			g.disposedCnt++
		}
	}

	var result []dtos.DisposalBatchDto
	for _, g := range groups {
		status := g.total > 0 && g.disposedCnt == g.total
		result = append(result, dtos.DisposalBatchDto{
			IDDisposal: g.idDisposal,
			StartDate:  g.startDate,
			EndDate:    g.endDate,
			Status:     status,
			Note:       g.note,
			UserID:     g.userID,
			AssetCount: g.total,
		})
	}

	sort.Slice(result, func(i, j int) bool {
		if !result[i].StartDate.Equal(result[j].StartDate.Time) {
			return result[i].StartDate.After(result[j].StartDate.Time)
		}
		return result[i].IDDisposal > result[j].IDDisposal
	})

	return result, nil
}

func (s *DisposalService) GetBatchDetails(ctx context.Context, userSession, idDisposal string) ([]dtos.DisposalBatchDetailDto, error) {
	allowedAreas, seeAll := s.authUseCase.GetUserAllowedAreas(ctx, userSession)
	allAssets, err := s.assetRepo.FindAllWithAreas(ctx, allowedAreas, seeAll, userSession)
	if err != nil {
		return nil, err
	}
	assetMap := make(map[string]domains.ASSET, len(allAssets))
	for _, a := range allAssets {
		assetMap[a.UnitCode] = a
	}

	items, err := s.disposalRepo.FindByIDDisposal(ctx, idDisposal)
	if err != nil {
		return nil, err
	}

	var details []dtos.DisposalBatchDetailDto
	for _, d := range items {
		a, ok := assetMap[d.UnitCode]
		if !ok {
			continue
		}
		details = append(details, dtos.DisposalBatchDetailDto{
			UnitCode:    d.UnitCode,
			IDDisposal:  d.IDDisposal,
			StartDate:   d.StartDate,
			EndDate:     d.EndDate,
			Status:      d.Status,
			Note:        d.Note,
			UserID:      d.UserID,
			AssetType:   a.Type,
			UserName:    a.UserName,
			StatusAsset: a.Status,
			State:       a.State,
		})
	}
	return details, nil
}

func (s *DisposalService) CheckUnitCodes(ctx context.Context, userSession string, req *dtos.CheckUnitCodesRequestDto) ([]dtos.UnitCodeCheckResultDto, error) {
	if req == nil {
		return []dtos.UnitCodeCheckResultDto{}, nil
	}
	allowedAreas, seeAll := s.authUseCase.GetUserAllowedAreas(ctx, userSession)
	allAssets, _ := s.assetRepo.FindAllWithAreas(ctx, allowedAreas, seeAll, userSession)
	allowedMap := make(map[string]domains.ASSET, len(allAssets))
	for _, a := range allAssets {
		allowedMap[strings.ToUpper(strings.TrimSpace(a.UnitCode))] = a
	}

	var results []dtos.UnitCodeCheckResultDto
	for _, rawCode := range req.UnitCodes {
		code := strings.TrimSpace(rawCode)
		if code == "" {
			continue
		}
		asset, ok := allowedMap[strings.ToUpper(code)]
		if !ok {
			results = append(results, dtos.UnitCodeCheckResultDto{
				UnitCode: code,
				Valid:    false,
				Error:    "Asset not found or outside allowed area",
			})
			continue
		}

		res := dtos.UnitCodeCheckResultDto{
			UnitCode: asset.UnitCode,
			Type:     asset.Type,
			UserName: asset.UserName,
			Status:   asset.Status,
			State:    asset.State,
		}

		if strings.EqualFold(asset.Status, "In-use") || strings.EqualFold(asset.Status, "Borrowed") {
			res.Valid = false
			res.Error = fmt.Sprintf("Asset is currently %s", asset.Status)
		} else if strings.EqualFold(asset.Status, "Waiting for Disposal") || strings.EqualFold(asset.Status, "Disposed") || s.disposalRepo.ExistsByUnitCode(ctx, asset.UnitCode) {
			res.Valid = false
			res.Error = "Asset is already in disposal process"
		} else {
			res.Valid = true
		}
		results = append(results, res)
	}

	return results, nil
}

func (s *DisposalService) ImportUnitCodes(ctx context.Context, userSession string, reader io.Reader) ([]string, error) {
	if s.excelProcessor == nil {
		return nil, errors.New("excel processor not configured")
	}
	return s.excelProcessor.ParseDisposalUnitCodesExcel(reader)
}

// ProblemReportService implements ports.ProblemReportUseCase with 9-digit ID generation, ASSETSLOG logging, and PDF export.
type ProblemReportService struct {
	*GenericService[domains.PROBLEMREPORT]
	problemRepo ports.ProblemReportRepository
	assetRepo   ports.AssetRepository
	pdfGen      ports.PDFGenerator
}

func NewProblemReportService(
	problemRepo ports.ProblemReportRepository,
	assetRepo ports.AssetRepository,
	userRepo ports.UserRepository,
	pdfGen ports.PDFGenerator,
) *ProblemReportService {
	return &ProblemReportService{
		GenericService: NewGenericService[domains.PROBLEMREPORT](problemRepo, userRepo, "PROBLEMREPORT"),
		problemRepo:    problemRepo,
		assetRepo:      assetRepo,
		pdfGen:         pdfGen,
	}
}

func (s *ProblemReportService) Create(ctx context.Context, userSession string, entity *domains.PROBLEMREPORT) error {
	if strings.TrimSpace(entity.ID) == "" {
		maxID, err := s.problemRepo.GetMaxNumericID(ctx)
		if err != nil {
			return err
		}
		entity.ID = fmt.Sprintf("%09d", maxID+1)
	}
	if entity.ReportDate.IsZero() {
		entity.ReportDate = domains.NowDateOnly()
	}
	if strings.TrimSpace(entity.UserCreated) == "" {
		entity.UserCreated = userSession
	}

	if err := s.problemRepo.Create(ctx, entity); err != nil {
		return err
	}

	nowTime := time.Now()
	if s.assetRepo != nil && strings.TrimSpace(entity.UnitCode) != "" {
		_ = s.assetRepo.AddAssetLog(ctx, &domains.ASSETSLOG{
			UnitCode:     entity.UnitCode,
			UserID:       userSession,
			DateModified: nowTime,
			Action:       "Report Issue",
			ActionLogs:   fmt.Sprintf("Problem ID: %s | Summary: %s", entity.ID, entity.ProblemSummary),
		})
	}
	if s.userRepo != nil && userSession != "" {
		s.userRepo.AddHistory(ctx, userSession, fmt.Sprintf("Add PROBLEMREPORT: %s", entity.ID))
	}
	return nil
}

func (s *ProblemReportService) Update(ctx context.Context, userSession string, id any, entity *domains.PROBLEMREPORT) error {
	if err := s.problemRepo.Update(ctx, id, entity); err != nil {
		return err
	}

	nowTime := time.Now()
	if s.assetRepo != nil && strings.TrimSpace(entity.UnitCode) != "" {
		_ = s.assetRepo.AddAssetLog(ctx, &domains.ASSETSLOG{
			UnitCode:     entity.UnitCode,
			UserID:       userSession,
			DateModified: nowTime,
			Action:       "Modify Problem",
			ActionLogs:   fmt.Sprintf("Problem ID: %v | Summary: %s", id, entity.ProblemSummary),
		})
	}
	if s.userRepo != nil && userSession != "" {
		s.userRepo.AddHistory(ctx, userSession, fmt.Sprintf("Update PROBLEMREPORT: %v", id))
	}
	return nil
}

func (s *ProblemReportService) ExportProblemReport(ctx context.Context, id string) ([]byte, error) {
	if s.pdfGen == nil {
		return nil, errors.New("PDF generator not configured")
	}

	report, err := s.problemRepo.FindByID(ctx, id)
	if err != nil || report == nil {
		return nil, errors.New("problem report not found")
	}

	var asset domains.ASSET
	if found, errA := s.assetRepo.FindByID(ctx, report.UnitCode); errA == nil && found != nil {
		asset = *found
	} else {
		asset.UnitCode = report.UnitCode
		asset.Type = report.Type
	}

	userCreatedFull := report.UserCreated
	if uC, errU := s.userRepo.FindByUserID(ctx, report.UserCreated); errU == nil && uC != nil && uC.UserName != "" {
		userCreatedFull = uC.UserName
	}

	userReportFull := report.UserReport
	if uR, errU := s.userRepo.FindByUserID(ctx, report.UserReport); errU == nil && uR != nil && uR.UserName != "" {
		userReportFull = uR.UserName
	}

	depts, _ := s.assetRepo.GetAllDepartments(ctx)
	deptMap := make(map[string]string, len(depts))
	for _, d := range depts {
		deptMap[d.ID] = d.Name
	}
	resolveDept := func(code string) string {
		if n, ok := deptMap[code]; ok && n != "" {
			return n
		}
		return code
	}

	configs, _ := s.assetRepo.GetAllDeviceConfigs(ctx)
	var devCfg *domains.DEVICECONFIG
	devType := firstNonEmpty(report.Type, asset.Type)
	for i := range configs {
		if strings.EqualFold(strings.TrimSpace(configs[i].Device), strings.TrimSpace(devType)) {
			devCfg = &configs[i]
			break
		}
	}

	items, _ := s.assetRepo.GetAllItemMasterData(ctx)
	itemMap := make(map[string]string, len(items))
	for _, it := range items {
		itemMap[it.ID] = it.TechSpecs
	}

	details := [9]string{
		firstNonEmpty(report.Detail1, asset.Detail1),
		firstNonEmpty(report.Detail2, asset.Detail2),
		firstNonEmpty(report.Detail3, asset.Detail3),
		firstNonEmpty(report.Detail4, asset.Detail4),
		firstNonEmpty(report.Detail5, asset.Detail5),
		firstNonEmpty(report.Detail6, asset.Detail6),
		firstNonEmpty(report.Detail7, asset.Detail7),
		firstNonEmpty(report.Detail8, asset.Detail8),
		firstNonEmpty(report.Detail9, asset.Detail9),
	}

	var labels [9]string
	if devCfg != nil {
		labels = [9]string{
			devCfg.Detail1, devCfg.Detail2, devCfg.Detail3,
			devCfg.Detail4, devCfg.Detail5, devCfg.Detail6,
			devCfg.Detail7, devCfg.Detail8, devCfg.Detail9,
		}
	}

	var specLines []string
	for i := 0; i < 9; i++ {
		val := strings.TrimSpace(details[i])
		if val == "" {
			continue
		}
		if resolved, ok := itemMap[val]; ok && resolved != "" {
			val = resolved
		}
		label := strings.TrimSpace(labels[i])
		if label == "" {
			label = fmt.Sprintf("Detail %d", i+1)
		}
		specLines = append(specLines, fmt.Sprintf("%s: %s", label, val))
	}

	pdfData := &ports.ProblemReportPDFData{
		Report:          *report,
		Asset:           asset,
		UserCreatedFull: userCreatedFull,
		UserReportFull:  userReportFull,
		DeptName:        resolveDept(firstNonEmpty(report.DepartmentReport, asset.Department)),
		OfficeName:      resolveDept(firstNonEmpty(report.OfficeReport, asset.Office)),
		CompanyName:     resolveDept(firstNonEmpty(report.CompanyReport, asset.Company)),
		SpecLines:       specLines,
	}

	return s.pdfGen.GenerateProblemReportPDF(pdfData)
}

// AssetsLogService implements ports.AssetsLogUseCase with Excel export support.
type AssetsLogService struct {
	*GenericService[domains.ASSETSLOG]
	logRepo        ports.AssetsLogRepository
	excelProcessor ports.ExcelProcessor
}

func NewAssetsLogService(
	logRepo ports.AssetsLogRepository,
	userRepo ports.UserRepository,
	excelProcessor ports.ExcelProcessor,
) *AssetsLogService {
	return &AssetsLogService{
		GenericService: NewGenericService[domains.ASSETSLOG](logRepo, userRepo, "ASSETSLOG"),
		logRepo:        logRepo,
		excelProcessor: excelProcessor,
	}
}

func (s *AssetsLogService) ExportExcel(ctx context.Context, unitCode, fromDateStr, toDateStr string) ([]byte, error) {
	if s.excelProcessor == nil {
		return nil, errors.New("excel processor not configured")
	}

	parseOptDate := func(val string, endOfDay bool) *time.Time {
		trimmed := strings.TrimSpace(val)
		if trimmed == "" {
			return nil
		}
		for _, layout := range []string{"2006-01-02", time.RFC3339, "02/01/2006"} {
			if t, err := time.Parse(layout, trimmed); err == nil {
				if endOfDay {
					t = time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 0, t.Location())
				}
				return &t
			}
		}
		return nil
	}

	fromDate := parseOptDate(fromDateStr, false)
	toDate := parseOptDate(toDateStr, true)

	logs, err := s.logRepo.FindForExport(ctx, unitCode, fromDate, toDate)
	if err != nil {
		return nil, err
	}

	return s.excelProcessor.ExportAssetsLogExcel(logs)
}
