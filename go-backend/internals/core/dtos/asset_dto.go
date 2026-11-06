package dtos

import (
	"go-backend/internals/core/domains"
)

type AssetDto struct {
	UnitCode           string            `json:"unitCode"`
	Type               string            `json:"type"`
	UserName           string            `json:"userName"`
	Detail1            string            `json:"detail1"`
	Detail2            string            `json:"detail2"`
	Detail3            string            `json:"detail3"`
	Detail4            string            `json:"detail4"`
	Detail5            string            `json:"detail5"`
	Detail6            string            `json:"detail6"`
	Detail7            string            `json:"detail7"`
	Detail8            string            `json:"detail8"`
	Detail9            string            `json:"detail9"`
	Model              string            `json:"model"`
	IDTag              string            `json:"idTag"`
	PurchasedDate      *domains.DateOnly `json:"purchasedDate"`
	UsageDate          *domains.DateOnly `json:"usageDate"`
	Brand              string            `json:"brand"`
	WarrantyExpiration *int              `json:"warrantyExpiration"`
	PurchasedValue     *float64          `json:"purchasedValue"`
	Vendor             string            `json:"vendor"`
	TypeEXP            string            `json:"typeEXP"`
	Notes              string            `json:"notes"`
	Department         string            `json:"department"`
	Office             string            `json:"office"`
	Area               string            `json:"area"`
	Company            string            `json:"company"`
	Status             string            `json:"status"`
	Serial             string            `json:"serial"`
	RecordDate         *domains.DateOnly `json:"recordDate"`
	HRCode             string            `json:"hrCode"`
	Depreciation       *float64          `json:"depreciation"`
	Value              *float64          `json:"value"`
	Accessories        *bool             `json:"accessories"`
	MainAsset          string            `json:"mainAsset"`
	Email              string            `json:"email"`
	State              string            `json:"state"`
	Watt               string            `json:"watt"`
	GroupDep           string            `json:"groupDep"`
	Division           string            `json:"division"`

	DepartmentName string     `json:"departmentName,omitempty"`
	OfficeName     string     `json:"officeName,omitempty"`
	CompanyName    string     `json:"companyName,omitempty"`
	VendorName     string     `json:"vendorName,omitempty"`
	StateName      string     `json:"stateName,omitempty"`
	DivisionName   string     `json:"divisionName,omitempty"`
	GroupDepName   string     `json:"groupDepName,omitempty"`
	AD             bool       `json:"ad"`
	ME             bool       `json:"me"`
	TM             bool       `json:"tm"`
	AdOS           string     `json:"adOS,omitempty"`
	MeTime         string     `json:"meTime,omitempty"`
	TmTime         string     `json:"tmTime,omitempty"`
	SubAssets      []AssetDto `json:"subAssets,omitempty"`
}

type UnassignedSubAssetDto struct {
	UnitCode  string            `json:"unitCode"`
	Type      string            `json:"type"`
	Model     string            `json:"model"`
	Serial    string            `json:"serial"`
	Brand     string            `json:"brand"`
	Status    string            `json:"status"`
	State     string            `json:"state"`
	Notes     string            `json:"notes"`
	UsageDate *domains.DateOnly `json:"usageDate"`
}

type UpdateMainAssetDto struct {
	MainAsset *string `json:"mainAsset"`
}

type ChartItemDto struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

type PurchasedYearDto struct {
	Year  int `json:"year"`
	Count int `json:"count"`
}

type DashboardMetricsDto struct {
	TotalAssets     int64              `json:"totalAssets"`
	TransferPC      int                `json:"transferPC"`
	NewPC           int                `json:"newPC"`
	GroupData       []ChartItemDto     `json:"groupData"`
	DivisionData    []ChartItemDto     `json:"divisionData"`
	RegionData      []ChartItemDto     `json:"regionData"`
	StateData       []ChartItemDto     `json:"stateData"`
	EolData         []ChartItemDto     `json:"eolData"`
	PurchasedData   []PurchasedYearDto `json:"purchasedData"`
	TotalInUse      int64              `json:"totalInUse,omitempty"`
	TotalSpare      int64              `json:"totalSpare,omitempty"`
	TotalDisposed   int64              `json:"totalDisposed,omitempty"`
	TotalBroken     int64              `json:"totalBroken,omitempty"`
	AssetsByType    map[string]int     `json:"assetsByType,omitempty"`
	AssetsByDept    map[string]int     `json:"assetsByDept,omitempty"`
	AssetsByCompany map[string]int     `json:"assetsByCompany,omitempty"`
	AssetsByArea    map[string]int     `json:"assetsByArea,omitempty"`
}

type AssetSuggestionDto struct {
	Value    string `json:"value"`
	Category string `json:"category"`
	State    string `json:"state"`
}

type ImportResultDto struct {
	ImportedCount  int    `json:"importedCount"`
	FailedCount    int    `json:"failedCount"`
	Message        string `json:"message"`
	ErrorFileBytes []byte `json:"-"`
}
