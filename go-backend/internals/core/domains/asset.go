package domains

import (
	"time"
)

type ASSET struct {
	UnitCode           string    `gorm:"column:UnitCode;primaryKey;type:varchar(21)" json:"unitCode"`
	Type               string    `gorm:"column:Type;type:varchar(50)" json:"type"`
	UserName           string    `gorm:"column:UserName;type:varchar(100)" json:"userName"`
	Detail1            string    `gorm:"column:Detail1;type:varchar(150)" json:"detail1"`
	Detail2            string    `gorm:"column:Detail2;type:varchar(150)" json:"detail2"`
	Detail3            string    `gorm:"column:Detail3;type:varchar(150)" json:"detail3"`
	Detail4            string    `gorm:"column:Detail4;type:varchar(150)" json:"detail4"`
	Detail5            string    `gorm:"column:Detail5;type:varchar(150)" json:"detail5"`
	Detail6            string    `gorm:"column:Detail6;type:varchar(150)" json:"detail6"`
	Detail7            string    `gorm:"column:Detail7;type:varchar(150)" json:"detail7"`
	Detail8            string    `gorm:"column:Detail8;type:varchar(150)" json:"detail8"`
	Detail9            string    `gorm:"column:Detail9;type:varchar(150)" json:"detail9"`
	Model              string    `gorm:"column:Model;type:varchar(100)" json:"model"`
	IDTag              string    `gorm:"column:IDTag;type:varchar(50)" json:"idTag"`
	PurchasedDate      *DateOnly `gorm:"column:PurchasedDate" json:"purchasedDate"`
	UsageDate          *DateOnly `gorm:"column:UsageDate" json:"usageDate"`
	Brand              string    `gorm:"column:Brand;type:varchar(50)" json:"brand"`
	WarrantyExpiration *int      `gorm:"column:WarrantyExpiration" json:"warrantyExpiration"`
	PurchasedValue     *float64  `gorm:"column:PurchasedValue" json:"purchasedValue"`
	Vendor             string    `gorm:"column:Vendor;type:varchar(100)" json:"vendor"`
	TypeEXP            string    `gorm:"column:TypeEXP;type:varchar(50)" json:"typeEXP"`
	Notes              string    `gorm:"column:Notes;type:text" json:"notes"`
	Department         string    `gorm:"column:Department;type:varchar(50)" json:"department"`
	Office             string    `gorm:"column:Office;type:varchar(50)" json:"office"`
	Area               string    `gorm:"column:Area;type:varchar(50)" json:"area"`
	Company            string    `gorm:"column:Company;type:varchar(50)" json:"company"`
	Status             string    `gorm:"column:Status;type:varchar(50)" json:"status"`
	Serial             string    `gorm:"column:Serial;type:varchar(100)" json:"serial"`
	RecordDate         *DateOnly `gorm:"column:RecordDate" json:"recordDate"`
	HRCode             string    `gorm:"column:HRCode;type:varchar(50)" json:"hrCode"`
	Depreciation       *float64  `gorm:"column:Depreciation" json:"depreciation"`
	Value              *float64  `gorm:"column:Value" json:"value"`
	Accessories        *bool     `gorm:"column:Accessories" json:"accessories"`
	MainAsset          string    `gorm:"column:MainAsset;type:varchar(21)" json:"mainAsset"`
	Email              string    `gorm:"column:Email;type:varchar(150)" json:"email"`
	State              string    `gorm:"column:State;type:varchar(50)" json:"state"`
	Watt               string    `gorm:"column:Watt;type:varchar(50)" json:"watt"`
	GroupDep           string    `gorm:"column:GroupDep;type:varchar(50)" json:"groupDep"`
	Division           string    `gorm:"column:Division;type:varchar(50)" json:"division"`

	DepartmentName string `gorm:"-" json:"departmentName,omitempty"`
	OfficeName     string `gorm:"-" json:"officeName,omitempty"`
	CompanyName    string `gorm:"-" json:"companyName,omitempty"`
	VendorName     string `gorm:"-" json:"vendorName,omitempty"`
	StateName      string `gorm:"-" json:"stateName,omitempty"`
	DivisionName   string `gorm:"-" json:"divisionName,omitempty"`
	GroupDepName   string `gorm:"-" json:"groupDepName,omitempty"`
	AD             bool   `gorm:"-" json:"ad"`
	ME             bool   `gorm:"-" json:"me"`
	TM             bool   `gorm:"-" json:"tm"`
	AdOS           string `gorm:"-" json:"adOS,omitempty"`
	MeTime         string `gorm:"-" json:"meTime,omitempty"`
	TmTime         string `gorm:"-" json:"tmTime,omitempty"`
}

func (ASSET) TableName() string {
	return "ASSETS"
}

type ACCESSORY struct {
	ID          int64  `gorm:"column:ID;primaryKey;autoIncrement" json:"id"`
	UnitCode    string `gorm:"column:UnitCode;type:varchar(21)" json:"unitCode"`
	Description string `gorm:"column:Description;type:varchar(150)" json:"description"`
	Brand       string `gorm:"column:Brand;type:varchar(20)" json:"brand"`
	Serial      string `gorm:"column:Serial;type:varchar(50)" json:"serial"`
}

func (ACCESSORY) TableName() string {
	return "ACCESSORIES"
}

type ASSETSLOG struct {
	ID           int64     `gorm:"column:ID;primaryKey;autoIncrement" json:"id"`
	UnitCode     string    `gorm:"column:UnitCode;type:varchar(21)" json:"unitCode"`
	UserID       string    `gorm:"column:UserID;type:varchar(50)" json:"userId"`
	Action       string    `gorm:"column:Action;type:varchar(50)" json:"action"`
	ActionLogs   string    `gorm:"column:ActionLogs;type:text" json:"actionLogs"`
	DateModified time.Time `gorm:"column:DateModified" json:"dateModified"`
}

func (ASSETSLOG) TableName() string {
	return "ASSETSLOGS"
}
