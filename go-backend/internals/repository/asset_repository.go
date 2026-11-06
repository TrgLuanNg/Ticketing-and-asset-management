package repository

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

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
			"Vendor", "Notes", "HRCode", "Email", "IDTag",
			"Detail1", "Detail2", "Detail3", "Detail4", "Detail5",
			"Detail6", "Detail7", "Detail8", "Detail9",
		},
	}
}

func (r *AssetRepository) applySecurityFilter(query *gorm.DB, allowedAreas []string, seeAll bool, userSession string) (*gorm.DB, bool) {
	if len(allowedAreas) == 0 {
		return query, false
	}

	hasVNOrAll := false
	var lowerAreas []string
	for _, a := range allowedAreas {
		trimmed := strings.TrimSpace(a)
		if strings.EqualFold(trimmed, "VN") || strings.EqualFold(trimmed, "ALL") {
			hasVNOrAll = true
		}
		if trimmed != "" {
			lowerAreas = append(lowerAreas, strings.ToLower(trimmed))
		}
	}

	if !hasVNOrAll {
		query = query.Where("LOWER(\"Area\") IN (?)", lowerAreas)
	}

	if !seeAll {
		if strings.TrimSpace(userSession) == "" {
			return query, false
		}
		query = query.Where("LOWER(\"HRCode\") = LOWER(?)", strings.TrimSpace(userSession))
	}

	return query, true
}

func (r *AssetRepository) applyFilter(query *gorm.DB, filter *dtos.Filter) *gorm.DB {
	if filter == nil {
		return query.Order("\"PurchasedDate\" DESC NULLS LAST, \"UnitCode\" DESC")
	}

	if strings.TrimSpace(filter.SearchText) != "" {
		placeholder := "___LITERAL_SLASH___"
		processedText := strings.ReplaceAll(filter.SearchText, "'/'", placeholder)
		rawParts := strings.Split(processedText, "/")
		for _, rawPart := range rawParts {
			part := strings.TrimSpace(strings.ReplaceAll(rawPart, placeholder, "/"))
			if part == "" {
				continue
			}
			likePattern := "%" + part + "%"
			var orClauses []string
			var orArgs []any
			for _, prop := range r.searchProps {
				orClauses = append(orClauses, fmt.Sprintf("LOWER(CAST(\"%s\" AS TEXT)) LIKE LOWER(?)", prop))
				orArgs = append(orArgs, likePattern)
			}
			if len(orClauses) > 0 {
				query = query.Where(strings.Join(orClauses, " OR "), orArgs...)
			}
		}
	}

	colMap := map[string]string{
		"unitcode":           "UnitCode",
		"type":               "Type",
		"username":           "UserName",
		"model":              "Model",
		"serial":             "Serial",
		"brand":              "Brand",
		"status":             "Status",
		"state":              "State",
		"department":         "Department",
		"office":             "Office",
		"company":            "Company",
		"area":               "Area",
		"hrcode":             "HRCode",
		"email":              "Email",
		"mainasset":          "MainAsset",
		"vendor":             "Vendor",
		"typeexp":            "TypeEXP",
		"notes":              "Notes",
		"watt":               "Watt",
		"groupdep":           "GroupDep",
		"division":           "Division",
		"idtag":              "IDTag",
		"purchaseddate":      "PurchasedDate",
		"usagedate":          "UsageDate",
		"recorddate":         "RecordDate",
		"warrantyexpiration": "WarrantyExpiration",
		"purchasedvalue":     "PurchasedValue",
		"depreciation":       "Depreciation",
		"value":              "Value",
		"accessories":        "Accessories",
		"detail1":            "Detail1",
		"detail2":            "Detail2",
		"detail3":            "Detail3",
		"detail4":            "Detail4",
		"detail5":            "Detail5",
		"detail6":            "Detail6",
		"detail7":            "Detail7",
		"detail8":            "Detail8",
		"detail9":            "Detail9",
	}

	for _, adv := range filter.AdvancedFilters {
		fieldKey := strings.ToLower(strings.TrimSpace(adv.Field))
		colName, exists := colMap[fieldKey]
		if !exists || strings.TrimSpace(adv.Value) == "" {
			continue
		}

		op := strings.TrimSpace(adv.Operator)
		rawVal := adv.Value

		if fieldKey == "mainasset" {
			if op == "==" || op == "=" || op == "eq" {
				query = query.Where(fmt.Sprintf("LOWER(\"%s\") = LOWER(?) AND LOWER(\"UnitCode\") != LOWER(?)", colName), rawVal, rawVal)
				continue
			}
		}

		placeholder := "___LITERAL_SLASH___"
		processedVal := strings.ReplaceAll(rawVal, "'/'", placeholder)

		if op == "==" || op == "=" || op == "eq" {
			parts := strings.Split(processedVal, "/")
			var trimmed []string
			for _, p := range parts {
				t := strings.TrimSpace(strings.ReplaceAll(p, placeholder, "/"))
				if t != "" {
					trimmed = append(trimmed, strings.ToLower(t))
				}
			}
			if len(trimmed) == 1 {
				query = query.Where(fmt.Sprintf("LOWER(CAST(\"%s\" AS TEXT)) = LOWER(?)", colName), trimmed[0])
			} else if len(trimmed) > 1 {
				query = query.Where(fmt.Sprintf("LOWER(CAST(\"%s\" AS TEXT)) IN (?)", colName), trimmed)
			}
		} else if op == "!=" || op == "neq" {
			parts := strings.Split(processedVal, "/")
			var trimmed []string
			for _, p := range parts {
				t := strings.TrimSpace(strings.ReplaceAll(p, placeholder, "/"))
				if t != "" {
					trimmed = append(trimmed, strings.ToLower(t))
				}
			}
			if len(trimmed) == 1 {
				query = query.Where(fmt.Sprintf("LOWER(CAST(\"%s\" AS TEXT)) != LOWER(?)", colName), trimmed[0])
			} else if len(trimmed) > 1 {
				query = query.Where(fmt.Sprintf("LOWER(CAST(\"%s\" AS TEXT)) NOT IN (?)", colName), trimmed)
			}
		} else if op == ">=" || op == "gte" {
			query = query.Where(fmt.Sprintf("\"%s\" >= ?", colName), rawVal)
		} else if op == "<=" || op == "lte" {
			query = query.Where(fmt.Sprintf("\"%s\" <= ?", colName), rawVal)
		} else if op == ">" || op == "gt" {
			query = query.Where(fmt.Sprintf("\"%s\" > ?", colName), rawVal)
		} else if op == "<" || op == "lt" {
			query = query.Where(fmt.Sprintf("\"%s\" < ?", colName), rawVal)
		} else if op == "startsWith" {
			query = query.Where(fmt.Sprintf("LOWER(CAST(\"%s\" AS TEXT)) LIKE LOWER(?)", colName), rawVal+"%")
		} else if op == "endsWith" {
			query = query.Where(fmt.Sprintf("LOWER(CAST(\"%s\" AS TEXT)) LIKE LOWER(?)", colName), "%"+rawVal)
		} else {
			parts := strings.Split(processedVal, "/")
			var orClauses []string
			var orArgs []any
			for _, p := range parts {
				t := strings.TrimSpace(strings.ReplaceAll(p, placeholder, "/"))
				if t != "" {
					orClauses = append(orClauses, fmt.Sprintf("LOWER(CAST(\"%s\" AS TEXT)) LIKE LOWER(?)", colName))
					orArgs = append(orArgs, "%"+t+"%")
				}
			}
			if len(orClauses) > 0 {
				query = query.Where(strings.Join(orClauses, " OR "), orArgs...)
			}
		}
	}

	if strings.TrimSpace(filter.OrderBy) != "" {
		parts := strings.Fields(filter.OrderBy)
		if len(parts) > 0 {
			fieldKey := strings.ToLower(parts[0])
			if colName, exists := colMap[fieldKey]; exists {
				direction := "ASC"
				if len(parts) > 1 && strings.EqualFold(parts[1], "desc") {
					direction = "DESC"
				}
				query = query.Order(fmt.Sprintf("\"%s\" %s", colName, direction))
			}
		}
	} else {
		query = query.Order("\"PurchasedDate\" DESC NULLS LAST, \"UnitCode\" DESC")
	}

	return query
}

func (r *AssetRepository) enrichAssets(ctx context.Context, assets []domains.ASSET) {
	if len(assets) == 0 {
		return
	}

	db := getDB(ctx, r.db)

	// 1. Lookup Departments & States
	var depts []domains.DEPARTMENT
	_ = db.Find(&depts).Error
	deptMap := make(map[string]string, len(depts))
	for _, d := range depts {
		deptMap[strings.ToLower(strings.TrimSpace(d.ID))] = d.Name
	}

	var states []domains.STATEMASTERDATA
	_ = db.Find(&states).Error
	stateMap := make(map[string]string, len(states))
	for _, st := range states {
		stateMap[strings.ToLower(strings.TrimSpace(st.ID))] = st.State
	}

	// 2. Collect UnitCodes/IDTags for AD / ME / TM lookup
	var lookupNames []string
	for _, a := range assets {
		if a.UnitCode != "" {
			lookupNames = append(lookupNames, strings.ToLower(strings.TrimSpace(a.UnitCode)))
		}
		if a.IDTag != "" {
			lookupNames = append(lookupNames, strings.ToLower(strings.TrimSpace(a.IDTag)))
		}
	}

	adMap := make(map[string]domains.ASSETSAD)
	meMap := make(map[string]domains.ASSETSME)
	tmMap := make(map[string]domains.ASSETSTM)

	if len(lookupNames) > 0 {
		var adList []domains.ASSETSAD
		_ = db.Where("LOWER(\"CN\") IN (?)", lookupNames).Find(&adList).Error
		for _, ad := range adList {
			adMap[strings.ToLower(strings.TrimSpace(ad.CN))] = ad
		}

		var meList []domains.ASSETSME
		_ = db.Where("LOWER(\"resource_name\") IN (?)", lookupNames).Find(&meList).Error
		for _, me := range meList {
			meMap[strings.ToLower(strings.TrimSpace(me.ResourceName))] = me
		}

		var tmList []domains.ASSETSTM
		_ = db.Where("LOWER(\"endpointHost\") IN (?)", lookupNames).Find(&tmList).Error
		for _, tm := range tmList {
			tmMap[strings.ToLower(strings.TrimSpace(tm.EndpointHost))] = tm
		}
	}

	resolveDept := func(code string) string {
		if n, ok := deptMap[strings.ToLower(strings.TrimSpace(code))]; ok && n != "" {
			return n
		}
		return "Unknown"
	}

	for i := range assets {
		a := &assets[i]
		a.DepartmentName = resolveDept(a.Department)
		a.OfficeName = resolveDept(a.Office)
		a.CompanyName = resolveDept(a.Company)
		a.VendorName = resolveDept(a.Vendor)
		a.DivisionName = resolveDept(a.Division)
		a.GroupDepName = resolveDept(a.GroupDep)

		if stName, ok := stateMap[strings.ToLower(strings.TrimSpace(a.State))]; ok && stName != "" {
			a.StateName = stName
		} else {
			a.StateName = "Unknown"
		}

		keyUC := strings.ToLower(strings.TrimSpace(a.UnitCode))
		keyTag := strings.ToLower(strings.TrimSpace(a.IDTag))

		if ad, ok := adMap[keyUC]; ok {
			a.AD = true
			a.AdOS = ad.OperatingSystem
		} else if keyTag != "" {
			if ad, ok := adMap[keyTag]; ok {
				a.AD = true
				a.AdOS = ad.OperatingSystem
			}
		}

		if me, ok := meMap[keyUC]; ok {
			a.ME = true
			a.MeTime = me.AgentLastContactTime
		} else if keyTag != "" {
			if me, ok := meMap[keyTag]; ok {
				a.ME = true
				a.MeTime = me.AgentLastContactTime
			}
		}

		if tm, ok := tmMap[keyUC]; ok {
			a.TM = true
			if tm.LastConnected != nil {
				a.TmTime = tm.LastConnected.Format("2006-01-02 15:04:05")
			}
		} else if keyTag != "" {
			if tm, ok := tmMap[keyTag]; ok {
				a.TM = true
				if tm.LastConnected != nil {
					a.TmTime = tm.LastConnected.Format("2006-01-02 15:04:05")
				}
			}
		}
	}
}

func (r *AssetRepository) FindPagedWithAreas(ctx context.Context, filter *dtos.Filter, allowedAreas []string, seeAll bool, userSession string) ([]domains.ASSET, int64, error) {
	var assets []domains.ASSET
	var total int64

	db := getDB(ctx, r.db)
	query := db.Model(&domains.ASSET{})

	query, allowed := r.applySecurityFilter(query, allowedAreas, seeAll, userSession)
	if !allowed {
		return []domains.ASSET{}, 0, nil
	}

	query = r.applyFilter(query, filter)

	if filter == nil || !filter.SkipCount {
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

	r.enrichAssets(ctx, assets)
	return assets, total, nil
}

func (r *AssetRepository) FindAllWithAreas(ctx context.Context, allowedAreas []string, seeAll bool, userSession string) ([]domains.ASSET, error) {
	var assets []domains.ASSET
	db := getDB(ctx, r.db)
	query := db.Model(&domains.ASSET{})

	query, allowed := r.applySecurityFilter(query, allowedAreas, seeAll, userSession)
	if !allowed {
		return []domains.ASSET{}, nil
	}

	if err := query.Order("\"PurchasedDate\" DESC NULLS LAST, \"UnitCode\" DESC").Find(&assets).Error; err != nil {
		return nil, err
	}
	r.enrichAssets(ctx, assets)
	return assets, nil
}

func (r *AssetRepository) FindByID(ctx context.Context, unitCode string) (*domains.ASSET, error) {
	var asset domains.ASSET
	err := getDB(ctx, r.db).Where("\"UnitCode\" = ?", unitCode).First(&asset).Error
	if err != nil {
		return nil, err
	}
	list := []domains.ASSET{asset}
	r.enrichAssets(ctx, list)
	return &list[0], nil
}

func (r *AssetRepository) FindSubAssets(ctx context.Context, mainUnitCode string) ([]domains.ASSET, error) {
	var subAssets []domains.ASSET
	err := getDB(ctx, r.db).Where("\"MainAsset\" = ? AND \"UnitCode\" != ?", mainUnitCode, mainUnitCode).Find(&subAssets).Error
	if err == nil && len(subAssets) > 0 {
		r.enrichAssets(ctx, subAssets)
	}
	return subAssets, err
}

func (r *AssetRepository) FindUnassignedSubAssets(ctx context.Context, filter *dtos.Filter) ([]dtos.UnassignedSubAssetDto, int64, error) {
	var dtosList []dtos.UnassignedSubAssetDto
	var total int64

	query := getDB(ctx, r.db).Model(&domains.ASSET{}).
		Where("(\"MainAsset\" IS NULL OR \"MainAsset\" = '')").
		Where("LOWER(\"Type\") NOT IN (?)", []string{"dpc", "ltp", "svr"})

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

func (r *AssetRepository) Create(ctx context.Context, asset *domains.ASSET) error {
	return getDB(ctx, r.db).Create(asset).Error
}

func (r *AssetRepository) Update(ctx context.Context, unitCode string, asset *domains.ASSET) error {
	return getDB(ctx, r.db).Model(&domains.ASSET{}).
		Where("\"UnitCode\" = ?", unitCode).
		Select("*").
		Omit("UnitCode").
		Updates(asset).Error
}

func (r *AssetRepository) Delete(ctx context.Context, unitCode string) error {
	return getDB(ctx, r.db).Where("\"UnitCode\" = ?", unitCode).Delete(&domains.ASSET{}).Error
}

func (r *AssetRepository) UpdateMainAsset(ctx context.Context, unitCode string, mainAsset *string) error {
	return getDB(ctx, r.db).Model(&domains.ASSET{}).
		Where("\"UnitCode\" = ?", unitCode).
		Update("MainAsset", mainAsset).Error
}

func (r *AssetRepository) UnitCodeExists(ctx context.Context, unitCode string) bool {
	var count int64
	getDB(ctx, r.db).Model(&domains.ASSET{}).Where("\"UnitCode\" = ?", unitCode).Count(&count)
	return count > 0
}

func (r *AssetRepository) SerialExists(ctx context.Context, serial string) bool {
	if strings.TrimSpace(serial) == "" {
		return false
	}
	var count int64
	getDB(ctx, r.db).Model(&domains.ASSET{}).Where("LOWER(\"Serial\") = LOWER(?)", strings.TrimSpace(serial)).Count(&count)
	return count > 0
}

func (r *AssetRepository) GetMatchingUnitCodes(ctx context.Context, prefix string) ([]string, error) {
	var matchingCodes []string
	err := getDB(ctx, r.db).Model(&domains.ASSET{}).
		Where("\"UnitCode\" LIKE ?", prefix+"%").
		Pluck("\"UnitCode\"", &matchingCodes).Error
	return matchingCodes, err
}

func (r *AssetRepository) AddAssetLog(ctx context.Context, log *domains.ASSETSLOG) error {
	return getDB(ctx, r.db).Create(log).Error
}

func (r *AssetRepository) GetAllDepartments(ctx context.Context) ([]domains.DEPARTMENT, error) {
	var list []domains.DEPARTMENT
	err := getDB(ctx, r.db).Find(&list).Error
	return list, err
}

func (r *AssetRepository) GetAllStates(ctx context.Context) ([]domains.STATEMASTERDATA, error) {
	var list []domains.STATEMASTERDATA
	err := getDB(ctx, r.db).Find(&list).Error
	return list, err
}

func (r *AssetRepository) GetAllDeviceConfigs(ctx context.Context) ([]domains.DEVICECONFIG, error) {
	var list []domains.DEVICECONFIG
	err := getDB(ctx, r.db).Find(&list).Error
	return list, err
}

func (r *AssetRepository) GetAllItemMasterData(ctx context.Context) ([]domains.ITEMMASTERDATA, error) {
	var list []domains.ITEMMASTERDATA
	err := getDB(ctx, r.db).Find(&list).Error
	return list, err
}

func (r *AssetRepository) GetAllTypes(ctx context.Context) ([]domains.TYPE, error) {
	var list []domains.TYPE
	err := getDB(ctx, r.db).Find(&list).Error
	return list, err
}

// GetDashboardMetrics computes dashboard charts matching .NET AssetService.GetDashboardMetrics and Angular statistics.ts
func (r *AssetRepository) GetDashboardMetrics(ctx context.Context, allowedAreas []string, seeAll bool, userSession, fromDate string) (*dtos.DashboardMetricsDto, error) {
	db := getDB(ctx, r.db)
	metrics := &dtos.DashboardMetricsDto{
		GroupData:       []dtos.ChartItemDto{},
		DivisionData:    []dtos.ChartItemDto{},
		RegionData:      []dtos.ChartItemDto{},
		StateData:       []dtos.ChartItemDto{},
		EolData:         []dtos.ChartItemDto{},
		PurchasedData:   []dtos.PurchasedYearDto{},
		AssetsByType:    make(map[string]int),
		AssetsByDept:    make(map[string]int),
		AssetsByCompany: make(map[string]int),
		AssetsByArea:    make(map[string]int),
	}

	startDate := time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC)
	if strings.TrimSpace(fromDate) != "" {
		if t, err := time.Parse("01/2006", strings.TrimSpace(fromDate)); err == nil {
			startDate = t
		} else if t, err := time.Parse("2006-01-02", strings.TrimSpace(fromDate)); err == nil {
			startDate = t
		}
	}

	baseQuery := db.Model(&domains.ASSET{})
	baseQuery, allowed := r.applySecurityFilter(baseQuery, allowedAreas, seeAll, userSession)
	if !allowed {
		return metrics, nil
	}

	query := baseQuery.
		Where("LOWER(\"Type\") IN (?)", []string{"dpc", "ltp"}).
		Where("\"State\" IN (?)", []string{"001", "002", "003"}).
		Where("LOWER(\"Area\") IN (?)", []string{"north", "south"}).
		Where("\"PurchasedDate\" >= ?", startDate.Format("2006-01-02"))

	var assets []domains.ASSET
	if err := query.Find(&assets).Error; err != nil {
		return nil, err
	}

	metrics.TotalAssets = int64(len(assets))

	// Lookup Departments and States
	depts, _ := r.GetAllDepartments(ctx)
	deptMap := make(map[string]string, len(depts))
	for _, d := range depts {
		deptMap[strings.ToLower(strings.TrimSpace(d.ID))] = d.Name
	}

	states, _ := r.GetAllStates(ctx)
	stateMap := make(map[string]string, len(states))
	for _, s := range states {
		stateMap[strings.ToLower(strings.TrimSpace(s.ID))] = s.State
	}

	// Count TransferPC ("RETURN") and NewPC ("ASS") from HANDOVEROTHERS
	var returnCount, assCount int64
	_ = db.Model(&domains.HANDOVEROTHER{}).
		Where("\"Form\" = ? AND \"HandoverDate\" >= ?", "RETURN", startDate.Format("2006-01-02")).
		Count(&returnCount).Error
	_ = db.Model(&domains.HANDOVEROTHER{}).
		Where("\"Form\" = ? AND \"HandoverDate\" >= ?", "ASS", startDate.Format("2006-01-02")).
		Count(&assCount).Error
	metrics.TransferPC = int(returnCount)
	metrics.NewPC = int(assCount)

	groupCounts := make(map[string]int)
	divCounts := make(map[string]int)
	regionCounts := make(map[string]int)
	stateCounts := make(map[string]int)
	yearCounts := make(map[int]int)
	over4Years := 0
	under4Years := 0
	fourYearsAgo := time.Now().AddDate(-4, 0, 0)

	for _, a := range assets {
		// GroupDep
		gName := deptMap[strings.ToLower(strings.TrimSpace(a.GroupDep))]
		if gName == "" {
			gName = "Unknown"
		}
		groupCounts[gName]++

		// Division
		dName := deptMap[strings.ToLower(strings.TrimSpace(a.Division))]
		if dName == "" {
			dName = "Unknown"
		}
		divCounts[dName]++

		// Region / Area
		reg := strings.TrimSpace(a.Area)
		if reg == "" {
			reg = "Unknown"
		}
		regionCounts[reg]++
		metrics.AssetsByArea[reg]++

		// State
		sName := stateMap[strings.ToLower(strings.TrimSpace(a.State))]
		if sName == "" {
			sName = a.State
		}
		if sName == "" {
			sName = "Unknown"
		}
		stateCounts[sName]++

		// Type & Dept legacy maps
		if a.Type != "" {
			metrics.AssetsByType[a.Type]++
		}
		if a.Department != "" {
			metrics.AssetsByDept[a.Department]++
		}

		// EOL & Purchased Year
		if a.PurchasedDate != nil && !a.PurchasedDate.IsZero() {
			yearCounts[a.PurchasedDate.Year()]++
			if a.PurchasedDate.Time.Before(fourYearsAgo) {
				over4Years++
			} else {
				under4Years++
			}
		} else {
			under4Years++
		}
	}

	toSortedChartItems := func(m map[string]int) []dtos.ChartItemDto {
		list := make([]dtos.ChartItemDto, 0, len(m))
		for k, v := range m {
			list = append(list, dtos.ChartItemDto{Name: k, Value: v})
		}
		sort.Slice(list, func(i, j int) bool {
			if list[i].Value != list[j].Value {
				return list[i].Value > list[j].Value
			}
			return list[i].Name < list[j].Name
		})
		return list
	}

	metrics.GroupData = toSortedChartItems(groupCounts)
	metrics.DivisionData = toSortedChartItems(divCounts)
	metrics.RegionData = toSortedChartItems(regionCounts)
	metrics.StateData = toSortedChartItems(stateCounts)
	metrics.EolData = []dtos.ChartItemDto{
		{Name: "Over 4 Years", Value: over4Years},
		{Name: "Under 4 Years", Value: under4Years},
	}

	for yr, cnt := range yearCounts {
		metrics.PurchasedData = append(metrics.PurchasedData, dtos.PurchasedYearDto{Year: yr, Count: cnt})
	}
	sort.Slice(metrics.PurchasedData, func(i, j int) bool {
		return metrics.PurchasedData[i].Year < metrics.PurchasedData[j].Year
	})

	return metrics, nil
}

// GetSuggestions returns up to 10 autocomplete suggestions matching .NET AssetService.GetSuggestions
func (r *AssetRepository) GetSuggestions(ctx context.Context, filter *dtos.Filter, allowedAreas []string, seeAll bool, userSession string) ([]dtos.AssetSuggestionDto, error) {
	if filter == nil || strings.TrimSpace(filter.SearchText) == "" {
		return []dtos.AssetSuggestionDto{}, nil
	}

	db := getDB(ctx, r.db)
	query := db.Model(&domains.ASSET{})
	query, allowed := r.applySecurityFilter(query, allowedAreas, seeAll, userSession)
	if !allowed {
		return []dtos.AssetSuggestionDto{}, nil
	}

	query = r.applyFilter(query, filter)

	var assets []domains.ASSET
	if err := query.Limit(50).Find(&assets).Error; err != nil {
		return nil, err
	}

	keyword := strings.ToLower(strings.TrimSpace(filter.SearchText))
	seen := make(map[string]bool)
	var results []dtos.AssetSuggestionDto

	addCandidate := func(val, category, state string) {
		trimmed := strings.TrimSpace(val)
		if trimmed == "" || len(results) >= 10 {
			return
		}
		if strings.Contains(strings.ToLower(trimmed), keyword) {
			key := category + ":" + strings.ToLower(trimmed)
			if !seen[key] {
				seen[key] = true
				results = append(results, dtos.AssetSuggestionDto{
					Value:    trimmed,
					Category: category,
					State:    state,
				})
			}
		}
	}

	for _, a := range assets {
		addCandidate(a.UnitCode, "UnitCode", a.State)
		addCandidate(a.UserName, "UserName", a.State)
		addCandidate(a.Serial, "Serial", a.State)
		addCandidate(a.Brand, "Brand", a.State)
		if len(results) >= 10 {
			break
		}
	}

	return results, nil
}
