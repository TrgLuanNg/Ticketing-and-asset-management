package domains

import "time"

type DEPARTMENT struct {
	ID            string `gorm:"column:ID;primaryKey;type:varchar(50)" json:"id"`
	Name          string `gorm:"column:Name;type:varchar(100)" json:"name"`
	Description   string `gorm:"column:Description;type:varchar(255)" json:"description"`
	Department    bool   `gorm:"column:Department" json:"department"`
	Office        bool   `gorm:"column:Office" json:"office"`
	Company       bool   `gorm:"column:Company" json:"company"`
	Vendor        bool   `gorm:"column:Vendor" json:"vendor"`
	Division      *bool  `gorm:"column:Division" json:"division"`
	GroupDivision *bool  `gorm:"column:GroupDivision" json:"groupDivision"`
	ParentId      string `gorm:"column:ParentId;type:varchar(50)" json:"parentId"`
}

func (DEPARTMENT) TableName() string {
	return "DEPARTMENT"
}

type TYPE struct {
	Type     string `gorm:"column:Type;primaryKey;type:varchar(50)" json:"type"`
	TypeName string `gorm:"column:TypeName;type:varchar(100)" json:"typeName"`
	VN       string `gorm:"column:VN;type:varchar(100)" json:"vn"`
}

func (TYPE) TableName() string {
	return "TYPE"
}

type DEVICECONFIG struct {
	Device  string `gorm:"column:Device;primaryKey;type:varchar(50)" json:"device"`
	Detail1 string `gorm:"column:Detail1;type:varchar(150)" json:"detail1"`
	Detail2 string `gorm:"column:Detail2;type:varchar(150)" json:"detail2"`
	Detail3 string `gorm:"column:Detail3;type:varchar(150)" json:"detail3"`
	Detail4 string `gorm:"column:Detail4;type:varchar(150)" json:"detail4"`
	Detail5 string `gorm:"column:Detail5;type:varchar(150)" json:"detail5"`
	Detail6 string `gorm:"column:Detail6;type:varchar(150)" json:"detail6"`
	Detail7 string `gorm:"column:Detail7;type:varchar(150)" json:"detail7"`
	Detail8 string `gorm:"column:Detail8;type:varchar(150)" json:"detail8"`
	Detail9 string `gorm:"column:Detail9;type:varchar(150)" json:"detail9"`
}

func (DEVICECONFIG) TableName() string {
	return "DEVICECONFIG"
}

type DISPOSAL struct {
	UnitCode   string   `gorm:"column:UnitCode;primaryKey;type:varchar(21)" json:"unitCode"`
	IDDisposal string   `gorm:"column:IDDisposal;primaryKey;type:varchar(50)" json:"idDisposal"`
	StartDate  DateOnly `gorm:"column:StartDate" json:"startDate"`
	EndDate    DateOnly `gorm:"column:EndDate" json:"endDate"`
	Status     bool     `gorm:"column:Status" json:"status"`
	Note       string   `gorm:"column:Note;type:text" json:"note"`
	UserID     string   `gorm:"column:UserID;type:varchar(50)" json:"userId"`
}

func (DISPOSAL) TableName() string {
	return "DISPOSAL"
}

type SOFTWARE struct {
	ID       int64  `gorm:"column:ID;primaryKey;autoIncrement" json:"id"`
	UnitCode string `gorm:"column:UnitCode;type:varchar(21)" json:"unitCode"`
	Software string `gorm:"column:Software;type:varchar(100)" json:"software"`
	Version  string `gorm:"column:Version;type:varchar(50)" json:"version"`
}

func (SOFTWARE) TableName() string {
	return "SOFTWARE"
}

type KEYTECHSPEC struct {
	Name                 string `gorm:"column:Name;primaryKey;type:varchar(100)" json:"name"`
	DescriptionTechSpecs string `gorm:"column:DescriptionTechSpecs;type:varchar(255)" json:"descriptionTechSpecs"`
}

func (KEYTECHSPEC) TableName() string {
	return "KEYTECHSPECS"
}

type ITEMMASTERDATA struct {
	ID        string `gorm:"column:ID;primaryKey;type:varchar(50)" json:"id"`
	Item      string `gorm:"column:Item;type:varchar(100)" json:"item"`
	TechSpecs string `gorm:"column:TechSpecs;type:text" json:"techSpecs"`
}

func (ITEMMASTERDATA) TableName() string {
	return "ITEMMASTERDATA"
}

type STATEMASTERDATA struct {
	ID          string `gorm:"column:ID;primaryKey;type:varchar(50)" json:"id"`
	State       string `gorm:"column:State;type:varchar(50)" json:"state"`
	Description string `gorm:"column:Description;type:varchar(255)" json:"description"`
}

func (STATEMASTERDATA) TableName() string {
	return "STATEMASTERDATA"
}

type EmployeeTitle struct {
	EmployeeCode string `gorm:"column:EmployeeCode;primaryKey;type:varchar(50)" json:"employeeCode"`
	Title        string `gorm:"column:Title;type:varchar(100)" json:"title"`
}

func (EmployeeTitle) TableName() string {
	return "EmployeeTitle"
}

type SERVICE struct {
	MainAssets    string    `gorm:"column:MainAssets;primaryKey;type:varchar(21)" json:"mainAssets"`
	SubAssets     string    `gorm:"column:SubAssets;primaryKey;type:varchar(21)" json:"subAssets"`
	ServiceLevel  string    `gorm:"column:ServiceLevel;type:varchar(50)" json:"serviceLevel"`
	TimeCreateTag *DateOnly `gorm:"column:TimeCreateTag" json:"timeCreateTag"`
}

func (SERVICE) TableName() string {
	return "SERVICE"
}

type ASSETSAD struct {
	ID              int64  `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	OperatingSystem string `gorm:"column:OperatingSystem;type:varchar(150)" json:"operatingSystem"`
	ServicePack     string `gorm:"column:ServicePack;type:varchar(150)" json:"servicePack"`
	Version         string `gorm:"column:Version;type:varchar(50)" json:"version"`
	CreateYear      string `gorm:"column:CreateYear;type:varchar(50)" json:"createYear"`
	CN              string `gorm:"column:CN;type:varchar(50)" json:"cn"`
	LastLogon       string `gorm:"column:LastLogon;type:varchar(50)" json:"lastLogon"`
	Status          string `gorm:"column:Status;type:varchar(50)" json:"status"`
}

func (ASSETSAD) TableName() string {
	return "ASSETSAD"
}

type ASSETSME struct {
	ID                    int64  `gorm:"column:Id;primaryKey;autoIncrement" json:"id"`
	ResourceName          string `gorm:"column:resource_name;type:varchar(150)" json:"resource_name"`
	AgentLastContactTime  string `gorm:"column:agent_last_contact_time;type:varchar(150)" json:"agent_last_contact_time"`
	OSName                string `gorm:"column:os_name;type:varchar(150)" json:"os_name"`
	IPAddress             string `gorm:"column:ip_address;type:varchar(150)" json:"ip_address"`
	MACAddress            string `gorm:"column:mac_address;type:varchar(150)" json:"mac_address"`
	LastSyncTime          string `gorm:"column:last_sync_time;type:varchar(150)" json:"last_sync_time"`
}

func (ASSETSME) TableName() string {
	return "ASSETSME"
}

type ASSETSTM struct {
	ID              int64      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	EndpointHost    string     `gorm:"column:endpointHost;type:varchar(250)" json:"endpointHost"`
	EndpointIP      string     `gorm:"column:endpointIP;type:varchar(50)" json:"endpointIP"`
	EndpointMAC     string     `gorm:"column:endpointMAC;type:varchar(250)" json:"endpointMAC"`
	LastConnected   *time.Time `gorm:"column:lastConnected" json:"lastConnected"`
}

func (ASSETSTM) TableName() string {
	return "ASSETSTM"
}

type ASSETSEMPATCHDETAIL struct {
	ID           int64  `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	ResourceName string `gorm:"column:resource_name;type:varchar(300)" json:"resource_name"`
	PatchID      string `gorm:"column:patch_id;type:varchar(300)" json:"patch_id"`
	PatchName    string `gorm:"column:patch_name;type:varchar(300)" json:"patch_name"`
	PatchStatus  string `gorm:"column:patch_status;type:varchar(300)" json:"patch_status"`
	Severity     string `gorm:"column:severity;type:varchar(300)" json:"severity"`
}

func (ASSETSEMPATCHDETAIL) TableName() string {
	return "ASSETSEMPATCHDETAIL"
}
