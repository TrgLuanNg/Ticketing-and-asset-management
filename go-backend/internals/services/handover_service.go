package services

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"go-backend/internals/core/domains"
	"go-backend/internals/core/dtos"
	"go-backend/internals/core/ports"
)

type HandoverService struct {
	handoverRepo ports.HandoverRepository
	assetRepo    ports.AssetRepository
	userRepo     ports.UserRepository
	txManager    ports.TransactionManager
	pdfGen       ports.PDFGenerator
}

func NewHandoverService(
	handoverRepo ports.HandoverRepository,
	assetRepo ports.AssetRepository,
	userRepo ports.UserRepository,
	txManager ports.TransactionManager,
	pdfGen ports.PDFGenerator,
) *HandoverService {
	return &HandoverService{
		handoverRepo: handoverRepo,
		assetRepo:    assetRepo,
		userRepo:     userRepo,
		txManager:    txManager,
		pdfGen:       pdfGen,
	}
}

func (s *HandoverService) GetPagedHandovers(ctx context.Context, unitCode, userSession string, filter *dtos.Filter) ([]dtos.HandoverDto, int64, error) {
	userInfo, _ := s.userRepo.FindByUserID(ctx, userSession)
	isViewer := userInfo != nil && strings.EqualFold(userInfo.Role, "viewer")
	userName := ""
	if userInfo != nil {
		userName = userInfo.UserName
	}

	entities, total, err := s.handoverRepo.GetPagedHandovers(ctx, unitCode, userSession, userName, isViewer, filter)
	if err != nil {
		return nil, 0, err
	}

	dtosList := make([]dtos.HandoverDto, len(entities))
	for i, h := range entities {
		dtosList[i] = dtos.HandoverDto{
			ID:                 h.ID,
			UnitCode:           h.UnitCode,
			HandoverDate:       h.HandoverDate,
			UserCreated:        h.UserCreated,
			Form:               h.Form,
			UserTransfer:       h.UserB,
			DepartmentTransfer: h.DepartmentB,
			OfficeTransfer:     h.OfficeB,
			CompanyTransfer:    h.CompanyB,
			AreaTransfer:       h.AreaB,
			Type:               h.Type,
			JiraCode:           h.JiraCode,
			Detail1:            h.Detail1,
			Detail2:            h.Detail2,
			Detail3:            h.Detail3,
			Detail4:            h.Detail4,
			Detail5:            h.Detail5,
			Detail6:            h.Detail6,
			Detail7:            h.Detail7,
			Detail8:            h.Detail8,
			Detail9:            h.Detail9,
		}
	}

	return dtosList, total, nil
}

func (s *HandoverService) GetAll(ctx context.Context) ([]domains.HANDOVER, error) {
	return s.handoverRepo.FindAll(ctx)
}

func (s *HandoverService) GetByID(ctx context.Context, id string) (*domains.HANDOVER, error) {
	return s.handoverRepo.FindByID(ctx, id)
}

func (s *HandoverService) Update(ctx context.Context, userSession, id string, entity *domains.HANDOVER) error {
	err := s.handoverRepo.UpdateHandover(ctx, id, entity)
	if err == nil {
		s.userRepo.AddHistory(ctx, userSession, fmt.Sprintf("Update HANDOVER: %s", id))
	}
	return err
}

func (s *HandoverService) Delete(ctx context.Context, userSession, id string) error {
	err := s.handoverRepo.DeleteHandover(ctx, id)
	if err == nil {
		s.userRepo.AddHistory(ctx, userSession, fmt.Sprintf("Delete HANDOVER: %s", id))
	}
	return err
}

func (s *HandoverService) Insert(ctx context.Context, userSession string, dto *dtos.HandoverDto) (*domains.HANDOVER, error) {
	if dto == nil || strings.TrimSpace(dto.UnitCode) == "" {
		return nil, errors.New("unitCode is required")
	}

	asset, err := s.assetRepo.FindByID(ctx, dto.UnitCode)
	if err != nil || asset == nil {
		return nil, fmt.Errorf("asset %s not found", dto.UnitCode)
	}

	subAssets, _ := s.assetRepo.FindSubAssets(ctx, dto.UnitCode)

	formUpper := strings.ToUpper(strings.TrimSpace(dto.Form))
	if formUpper == "" {
		formUpper = "ASS"
	}

	userCreated := strings.TrimSpace(dto.UserCreated)
	if userCreated == "" {
		userCreated = userSession
	}

	handoverDate := domains.NowDateOnly()
	if !dto.HandoverDate.IsZero() {
		handoverDate = dto.HandoverDate
	}

	var (
		userA, deptA, officeA, companyA, areaA string
		userB, deptB, officeB, companyB, areaB string
		shouldInsertHandover                   bool
	)

	origHRCode := asset.HRCode

	switch formUpper {
	case "ASS":
		userA = "ITD"
		deptA = "ITD"
		officeA = asset.Office
		companyA = asset.Company
		areaA = asset.Area

		userB = dto.UserTransfer
		deptB = dto.DepartmentTransfer
		officeB = firstNonEmpty(dto.OfficeTransfer, asset.Office)
		companyB = firstNonEmpty(dto.CompanyTransfer, asset.Company)
		areaB = firstNonEmpty(dto.AreaTransfer, asset.Area)

		asset.Status = "In-use"
		asset.UserName = userB
		asset.HRCode = dto.HRCode
		asset.Department = deptB
		asset.GroupDep = dto.GroupDepTransfer
		asset.Division = dto.DivisionTransfer
		asset.Office = officeB
		asset.Company = companyB
		asset.Area = areaB
		if dto.Email != "" {
			asset.Email = dto.Email
		}
		asset.UsageDate = &handoverDate
		if dto.State != "" {
			asset.State = dto.State
		}
		shouldInsertHandover = true

	case "LOA":
		userA = "ITD"
		deptA = "ITD"
		officeA = asset.Office
		companyA = asset.Company
		areaA = asset.Area

		userB = dto.UserTransfer
		deptB = dto.DepartmentTransfer
		officeB = firstNonEmpty(dto.OfficeTransfer, asset.Office)
		companyB = firstNonEmpty(dto.CompanyTransfer, asset.Company)
		areaB = firstNonEmpty(dto.AreaTransfer, asset.Area)

		asset.Status = "Borrowed"
		asset.UserName = userB
		asset.HRCode = dto.HRCode
		asset.Department = deptB
		asset.GroupDep = dto.GroupDepTransfer
		asset.Division = dto.DivisionTransfer
		asset.Office = officeB
		asset.Company = companyB
		asset.Area = areaB
		if dto.Email != "" {
			asset.Email = dto.Email
		}
		asset.UsageDate = &handoverDate
		if dto.State != "" {
			asset.State = dto.State
		}
		shouldInsertHandover = false

	case "RETURN":
		userA = asset.UserName
		deptA = asset.Department
		officeA = asset.Office
		companyA = asset.Company
		areaA = asset.Area

		userB = "ITD"
		deptB = "ITD"
		officeB = firstNonEmpty(dto.OfficeTransfer, asset.Office)
		companyB = firstNonEmpty(dto.CompanyTransfer, asset.Company)
		areaB = firstNonEmpty(dto.AreaTransfer, asset.Area)

		asset.Status = "In-stock"
		asset.UserName = ""
		asset.HRCode = ""
		asset.Department = ""
		asset.GroupDep = ""
		asset.Division = ""
		asset.Email = ""
		asset.Office = officeB
		asset.Company = companyB
		asset.Area = areaB
		if dto.State != "" {
			asset.State = dto.State
		}
		shouldInsertHandover = true

	case "TRF":
		userA = asset.UserName
		deptA = asset.Department
		officeA = asset.Office
		companyA = asset.Company
		areaA = asset.Area

		userB = dto.UserTransfer
		deptB = dto.DepartmentTransfer
		officeB = firstNonEmpty(dto.OfficeTransfer, asset.Office)
		companyB = firstNonEmpty(dto.CompanyTransfer, asset.Company)
		areaB = firstNonEmpty(dto.AreaTransfer, asset.Area)

		asset.Status = "In-use"
		asset.UserName = userB
		asset.HRCode = dto.HRCode
		asset.Department = deptB
		asset.GroupDep = dto.GroupDepTransfer
		asset.Division = dto.DivisionTransfer
		asset.Office = officeB
		asset.Company = companyB
		asset.Area = areaB
		if dto.Email != "" {
			asset.Email = dto.Email
		}
		asset.UsageDate = &handoverDate
		if dto.State != "" {
			asset.State = dto.State
		}
		shouldInsertHandover = false

	case "PRB":
		userA = asset.UserName
		deptA = asset.Department
		officeA = asset.Office
		companyA = asset.Company
		areaA = asset.Area

		retrieveToITD := dto.RetrieveToITD != nil && *dto.RetrieveToITD
		if retrieveToITD {
			userB = "ITD"
			deptB = "ITD"
			officeB = firstNonEmpty(dto.OfficeTransfer, asset.Office)
			companyB = firstNonEmpty(dto.CompanyTransfer, asset.Company)
			areaB = firstNonEmpty(dto.AreaTransfer, asset.Area)

			asset.Status = "In-stock"
			asset.UserName = ""
			asset.HRCode = ""
			asset.Department = ""
			asset.GroupDep = ""
			asset.Division = ""
			asset.Email = ""
			asset.Office = officeB
			asset.Company = companyB
			asset.Area = areaB
			asset.State = firstNonEmpty(dto.State, "00002")
			shouldInsertHandover = true
		} else {
			userB = asset.UserName
			deptB = asset.Department
			officeB = asset.Office
			companyB = asset.Company
			areaB = asset.Area
			asset.State = firstNonEmpty(dto.State, "00002")
			shouldInsertHandover = false
		}

	case "LOST":
		userA = asset.UserName
		deptA = asset.Department
		officeA = asset.Office
		companyA = asset.Company
		areaA = asset.Area

		userB = "ITD"
		deptB = "ITD"
		officeB = asset.Office
		companyB = asset.Company
		areaB = asset.Area

		asset.Status = "Lost"
		asset.State = firstNonEmpty(dto.State, "00004")
		shouldInsertHandover = false

	default:
		userA = asset.UserName
		deptA = asset.Department
		officeA = asset.Office
		companyA = asset.Company
		areaA = asset.Area

		userB = dto.UserTransfer
		deptB = dto.DepartmentTransfer
		officeB = firstNonEmpty(dto.OfficeTransfer, asset.Office)
		companyB = firstNonEmpty(dto.CompanyTransfer, asset.Company)
		areaB = firstNonEmpty(dto.AreaTransfer, asset.Area)
	}

	var resultHandover *domains.HANDOVER

	err = s.txManager.RunInTransaction(ctx, func(txCtx context.Context) error {
		if err := s.assetRepo.Update(txCtx, asset.UnitCode, asset); err != nil {
			return err
		}

		// Update sub-assets according to form rules
		transferSubs := dto.TransferSubAssets == nil || *dto.TransferSubAssets
		retrieveSubs := dto.RetrieveToITD == nil || *dto.RetrieveToITD

		for i := range subAssets {
			sub := subAssets[i]
			switch formUpper {
			case "ASS", "LOA", "TRF":
				if transferSubs {
					sub.Status = asset.Status
					sub.UserName = asset.UserName
					sub.HRCode = asset.HRCode
					sub.Department = asset.Department
					sub.GroupDep = asset.GroupDep
					sub.Division = asset.Division
					sub.Office = asset.Office
					sub.Company = asset.Company
					sub.Area = asset.Area
					sub.Email = asset.Email
					sub.UsageDate = asset.UsageDate
				} else {
					sub.MainAsset = ""
				}
			case "RETURN":
				if retrieveSubs {
					sub.Status = "In-stock"
					sub.UserName = ""
					sub.HRCode = ""
					sub.Department = ""
					sub.GroupDep = ""
					sub.Division = ""
					sub.Email = ""
					sub.Office = asset.Office
					sub.Company = asset.Company
					sub.Area = asset.Area
				} else {
					sub.MainAsset = ""
				}
			case "PRB":
				if dto.RetrieveToITD != nil && *dto.RetrieveToITD {
					sub.Status = "In-stock"
					sub.UserName = ""
					sub.HRCode = ""
					sub.Department = ""
					sub.GroupDep = ""
					sub.Division = ""
					sub.Email = ""
				}
			case "LOST":
				sub.MainAsset = ""
			}
			_ = s.assetRepo.Update(txCtx, sub.UnitCode, &sub)
		}

		// Upsert UpdateEmployeeCode if HRCode provided on ASS / LOA / TRF
		if (formUpper == "ASS" || formUpper == "LOA" || formUpper == "TRF") && strings.TrimSpace(dto.HRCode) != "" {
			_ = s.handoverRepo.UpsertUpdateEmployeeCode(txCtx, &domains.UpdateEmployeeCode{
				UnitCode: asset.UnitCode,
				HRCode:   strings.TrimSpace(dto.HRCode),
			})
		}

		// Generate 9-digit ID (yyMMdd001 or max+1)
		datePrefix := time.Now().Format("060102")
		minID, _ := strconv.ParseInt(datePrefix+"001", 10, 64)
		maxID, _ := s.handoverRepo.GetMaxHandoverOtherID(txCtx)
		nextID := minID
		if maxID >= minID {
			nextID = maxID + 1
		}
		idStr := fmt.Sprintf("%09d", nextID)

		// Generate HandoverNo: {hrCode5}.{MMddyy}.{seq2}
		hrCode := firstNonEmpty(strings.TrimSpace(dto.HRCode), strings.TrimSpace(origHRCode), "00000")
		if len(hrCode) > 5 {
			hrCode = hrCode[len(hrCode)-5:]
		} else if len(hrCode) < 5 {
			hrCode = strings.Repeat("0", 5-len(hrCode)) + hrCode
		}
		mmddyy := time.Now().Format("010206")
		hoPrefix := fmt.Sprintf("%s.%s.", hrCode, mmddyy)
		seqCount, _ := s.handoverRepo.CountHandoverOthersByPrefix(txCtx, hoPrefix)
		handoverNo := fmt.Sprintf("%s%02d", hoPrefix, seqCount+1)

		ho := &domains.HANDOVEROTHER{
			ID:           idStr,
			UnitCode:     asset.UnitCode,
			HandoverDate: handoverDate,
			UserCreated:  userCreated,
			Form:         formUpper,
			UserA:        userA,
			DepartmentA:  deptA,
			OfficeA:      officeA,
			CompanyA:     companyA,
			AreaA:        areaA,
			UserB:        userB,
			DepartmentB:  deptB,
			OfficeB:      officeB,
			CompanyB:     companyB,
			AreaB:        areaB,
			Type:         firstNonEmpty(dto.Type, asset.Type),
			JiraCode:     dto.JiraCode,
			Detail1:      firstNonEmpty(dto.Detail1, asset.Detail1),
			Detail2:      firstNonEmpty(dto.Detail2, asset.Detail2),
			Detail3:      firstNonEmpty(dto.Detail3, asset.Detail3),
			Detail4:      firstNonEmpty(dto.Detail4, asset.Detail4),
			Detail5:      firstNonEmpty(dto.Detail5, asset.Detail5),
			Detail6:      firstNonEmpty(dto.Detail6, asset.Detail6),
			Detail7:      firstNonEmpty(dto.Detail7, asset.Detail7),
			Detail8:      firstNonEmpty(dto.Detail8, asset.Detail8),
			Detail9:      firstNonEmpty(dto.Detail9, asset.Detail9),
			HandoverNo:   handoverNo,
		}

		if err := s.handoverRepo.CreateHandoverOther(txCtx, ho); err != nil {
			return err
		}

		h := &domains.HANDOVER{
			ID:                 idStr,
			UnitCode:           asset.UnitCode,
			HandoverDate:       handoverDate,
			UserCreated:        userCreated,
			Form:               formUpper,
			UserTransfer:       userB,
			DepartmentTransfer: deptB,
			OfficeTransfer:     officeB,
			CompanyTransfer:    companyB,
			AreaTransfer:       areaB,
			Type:               ho.Type,
			JiraCode:           dto.JiraCode,
			Detail1:            ho.Detail1,
			Detail2:            ho.Detail2,
			Detail3:            ho.Detail3,
			Detail4:            ho.Detail4,
			Detail5:            ho.Detail5,
			Detail6:            ho.Detail6,
			Detail7:            ho.Detail7,
			Detail8:            ho.Detail8,
			Detail9:            ho.Detail9,
		}

		if shouldInsertHandover {
			if err := s.handoverRepo.CreateHandover(txCtx, h); err != nil {
				return err
			}
		}
		resultHandover = h

		nowTime := time.Now()
		_ = s.assetRepo.AddAssetLog(txCtx, &domains.ASSETSLOG{
			UnitCode:     asset.UnitCode,
			UserID:       userSession,
			DateModified: nowTime,
			Action:       fmt.Sprintf("Handover (%s)", formUpper),
			ActionLogs:   fmt.Sprintf("Form: %s | From: %s -> To: %s | HandoverNo: %s", formUpper, userA, userB, handoverNo),
		})

		return nil
	})

	if err != nil {
		return nil, err
	}

	s.userRepo.AddHistory(ctx, userSession, fmt.Sprintf("Create HANDOVER (%s) for %s", formUpper, asset.UnitCode))
	return resultHandover, nil
}

func (s *HandoverService) ExportHandoverForm(ctx context.Context, handoverID, unitCode string) ([]byte, error) {
	if s.pdfGen == nil {
		return nil, errors.New("PDF generator not configured")
	}

	handover, err := s.handoverRepo.FindOtherByID(ctx, handoverID, unitCode)
	if err != nil || handover == nil {
		return nil, errors.New("handover record not found")
	}

	targetUnitCode := firstNonEmpty(unitCode, handover.UnitCode)
	var asset domains.ASSET
	if foundAsset, errA := s.assetRepo.FindByID(ctx, targetUnitCode); errA == nil && foundAsset != nil {
		asset = *foundAsset
	} else {
		asset.UnitCode = targetUnitCode
		asset.Type = handover.Type
	}

	userAFull := handover.UserA
	if uA, errU := s.userRepo.FindByUserID(ctx, handover.UserA); errU == nil && uA != nil && uA.UserName != "" {
		userAFull = uA.UserName
	}
	userBFull := handover.UserB
	if uB, errU := s.userRepo.FindByUserID(ctx, handover.UserB); errU == nil && uB != nil && uB.UserName != "" {
		userBFull = uB.UserName
	}

	depts, _ := s.assetRepo.GetAllDepartments(ctx)
	deptMap := make(map[string]string, len(depts))
	for _, d := range depts {
		deptMap[d.ID] = d.Name
	}
	resolveDept := func(id string) string {
		if name, ok := deptMap[id]; ok && name != "" {
			return name
		}
		return id
	}

	specLines := s.buildSpecLines(ctx, firstNonEmpty(handover.Type, asset.Type), [9]string{
		firstNonEmpty(handover.Detail1, asset.Detail1),
		firstNonEmpty(handover.Detail2, asset.Detail2),
		firstNonEmpty(handover.Detail3, asset.Detail3),
		firstNonEmpty(handover.Detail4, asset.Detail4),
		firstNonEmpty(handover.Detail5, asset.Detail5),
		firstNonEmpty(handover.Detail6, asset.Detail6),
		firstNonEmpty(handover.Detail7, asset.Detail7),
		firstNonEmpty(handover.Detail8, asset.Detail8),
		firstNonEmpty(handover.Detail9, asset.Detail9),
	})

	accessories, _ := s.handoverRepo.GetAccessoriesByUnitCode(ctx, targetUnitCode)
	softwares, _ := s.handoverRepo.GetSoftwaresByUnitCode(ctx, targetUnitCode)

	pdfData := &ports.HandoverPDFData{
		Handover:     *handover,
		Asset:        asset,
		UserAFull:    userAFull,
		UserBFull:    userBFull,
		DeptAName:    resolveDept(handover.DepartmentA),
		DeptBName:    resolveDept(handover.DepartmentB),
		OfficeAName:  resolveDept(handover.OfficeA),
		OfficeBName:  resolveDept(handover.OfficeB),
		CompanyAName: resolveDept(handover.CompanyA),
		CompanyBName: resolveDept(handover.CompanyB),
		SpecLines:    specLines,
		Accessories:  accessories,
		Softwares:    softwares,
	}

	return s.pdfGen.GenerateHandoverPDF(pdfData)
}

func (s *HandoverService) buildSpecLines(ctx context.Context, deviceType string, details [9]string) []string {
	configs, _ := s.assetRepo.GetAllDeviceConfigs(ctx)
	var devCfg *domains.DEVICECONFIG
	for i := range configs {
		if strings.EqualFold(strings.TrimSpace(configs[i].Device), strings.TrimSpace(deviceType)) {
			devCfg = &configs[i]
			break
		}
	}

	items, _ := s.assetRepo.GetAllItemMasterData(ctx)
	itemMap := make(map[string]string, len(items))
	for _, it := range items {
		itemMap[it.ID] = it.TechSpecs
	}

	var labels [9]string
	if devCfg != nil {
		labels = [9]string{
			devCfg.Detail1, devCfg.Detail2, devCfg.Detail3,
			devCfg.Detail4, devCfg.Detail5, devCfg.Detail6,
			devCfg.Detail7, devCfg.Detail8, devCfg.Detail9,
		}
	}

	var lines []string
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
		lines = append(lines, fmt.Sprintf("%s: %s", label, val))
	}
	return lines
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
