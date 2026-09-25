package dtos

import "go-backend/internals/core/domains"

type HandoverDto struct {
	ID                 string           `json:"id"`
	UnitCode           string           `json:"unitCode"`
	HandoverDate       domains.DateOnly `json:"handoverDate"`
	UserCreated        string           `json:"userCreated"`
	Form               string           `json:"form"`
	UserTransfer       string           `json:"userTransfer"`
	DepartmentTransfer string           `json:"departmentTransfer"`
	OfficeTransfer     string           `json:"officeTransfer"`
	CompanyTransfer    string           `json:"companyTransfer"`
	AreaTransfer       string           `json:"areaTransfer"`
	Type               string           `json:"type"`
	JiraCode           string           `json:"jiraCode"`
	Detail1            string           `json:"detail1"`
	Detail2            string           `json:"detail2"`
	Detail3            string           `json:"detail3"`
	Detail4            string           `json:"detail4"`
	Detail5            string           `json:"detail5"`
	Detail6            string           `json:"detail6"`
	Detail7            string           `json:"detail7"`
	Detail8            string           `json:"detail8"`
	Detail9            string           `json:"detail9"`
	DivisionTransfer   string           `json:"divisionTransfer,omitempty"`
	GroupDepTransfer   string           `json:"groupDepTransfer,omitempty"`
	EmailTransfer      string           `json:"emailTransfer,omitempty"`
}

type DepartmentDto struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	Department    bool   `json:"department"`
	Office        bool   `json:"office"`
	Company       bool   `json:"company"`
	Vendor        bool   `json:"vendor"`
	Division      *bool  `json:"division"`
	GroupDivision *bool  `json:"groupDivision"`
	ParentId      string `json:"parentId"`
}

type AssetTypeDto struct {
	Type     string `json:"type"`
	TypeName string `json:"typeName"`
	VN       string `json:"vn"`
}

type DeviceConfigDto struct {
	Device  string `json:"device"`
	Detail1 string `json:"detail1"`
	Detail2 string `json:"detail2"`
	Detail3 string `json:"detail3"`
	Detail4 string `json:"detail4"`
	Detail5 string `json:"detail5"`
	Detail6 string `json:"detail6"`
	Detail7 string `json:"detail7"`
	Detail8 string `json:"detail8"`
	Detail9 string `json:"detail9"`
}

type DisposalDto struct {
	UnitCode   string           `json:"unitCode"`
	IDDisposal string           `json:"idDisposal"`
	StartDate  domains.DateOnly `json:"startDate"`
	EndDate    domains.DateOnly `json:"endDate"`
	Status     bool             `json:"status"`
	Note       string           `json:"note"`
	UserID     string           `json:"userId"`
}

type SoftwareDto struct {
	ID       int64  `json:"id"`
	UnitCode string `json:"unitCode"`
	Software string `json:"software"`
	Version  string `json:"version"`
}

type KeyTechSpecDto struct {
	Name                 string `json:"name"`
	DescriptionTechSpecs string `json:"descriptionTechSpecs"`
}

type ItemMasterDataDto struct {
	ID        string `json:"id"`
	Item      string `json:"item"`
	TechSpecs string `json:"techSpecs"`
}

type StateMasterDataDto struct {
	ID          string `json:"id"`
	State       string `json:"state"`
	Description string `json:"description"`
}

type TitleEmployeeDto struct {
	EmployeeCode string `json:"employeeCode"`
	Title        string `json:"title"`
}

type ServiceDto struct {
	MainAssets    string            `json:"mainAssets"`
	SubAssets     string            `json:"subAssets"`
	ServiceLevel  string            `json:"serviceLevel"`
	TimeCreateTag *domains.DateOnly `json:"timeCreateTag"`
}

type AccessoryDto struct {
	ID          int64  `json:"id"`
	UnitCode    string `json:"unitCode"`
	Description string `json:"description"`
	Brand       string `json:"brand"`
	Serial      string `json:"serial"`
}

type AssetsLogDto struct {
	ID           int64  `json:"id"`
	UnitCode     string `json:"unitCode"`
	UserID       string `json:"userId"`
	Action       string `json:"action"`
	ActionLogs   string `json:"actionLogs"`
	DateModified string `json:"dateModified"`
}

type ProblemReportDto struct {
	ID               string           `json:"id"`
	UnitCode         string           `json:"unitCode"`
	ReportDate       domains.DateOnly `json:"reportDate"`
	UserCreated      string           `json:"userCreated"`
	ProblemSummary   string           `json:"problemSummary"`
	TestProcedures   string           `json:"testProcedures"`
	Solutions        string           `json:"solutions"`
	Type             string           `json:"type"`
	Detail1          string           `json:"detail1"`
	Detail2          string           `json:"detail2"`
	Detail3          string           `json:"detail3"`
	Detail4          string           `json:"detail4"`
	Detail5          string           `json:"detail5"`
	Detail6          string           `json:"detail6"`
	Detail7          string           `json:"detail7"`
	Detail8          string           `json:"detail8"`
	Detail9          string           `json:"detail9"`
	UserReport       string           `json:"userReport"`
	DepartmentReport string           `json:"departmentReport"`
	OfficeReport     string           `json:"officeReport"`
	CompanyReport    string           `json:"companyReport"`
	AreaReport       string           `json:"areaReport"`
	ProblemNo        string           `json:"problemNo"`
	JiraCode         string           `json:"jiraCode"`
}
