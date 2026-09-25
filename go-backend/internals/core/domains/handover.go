package domains

type HANDOVER struct {
	ID                 string    `gorm:"column:ID;primaryKey;type:varchar(50)" json:"id"`
	UnitCode           string    `gorm:"column:UnitCode;type:varchar(21)" json:"unitCode"`
	HandoverDate       DateOnly  `gorm:"column:HandoverDate" json:"handoverDate"`
	UserCreated        string    `gorm:"column:UserCreated;type:varchar(50)" json:"userCreated"`
	Form               string    `gorm:"column:Form;type:varchar(50)" json:"form"`
	UserTransfer       string    `gorm:"column:UserTransfer;type:varchar(100)" json:"userTransfer"`
	DepartmentTransfer string    `gorm:"column:DepartmentTransfer;type:varchar(50)" json:"departmentTransfer"`
	OfficeTransfer     string    `gorm:"column:OfficeTransfer;type:varchar(50)" json:"officeTransfer"`
	CompanyTransfer    string    `gorm:"column:CompanyTransfer;type:varchar(50)" json:"companyTransfer"`
	AreaTransfer       string    `gorm:"column:AreaTransfer;type:varchar(50)" json:"areaTransfer"`
	Type               string    `gorm:"column:Type;type:varchar(50)" json:"type"`
	JiraCode           string    `gorm:"column:JiraCode;type:varchar(50)" json:"jiraCode"`
	Detail1            string    `gorm:"column:Detail1;type:varchar(150)" json:"detail1"`
	Detail2            string    `gorm:"column:Detail2;type:varchar(150)" json:"detail2"`
	Detail3            string    `gorm:"column:Detail3;type:varchar(150)" json:"detail3"`
	Detail4            string    `gorm:"column:Detail4;type:varchar(150)" json:"detail4"`
	Detail5            string    `gorm:"column:Detail5;type:varchar(150)" json:"detail5"`
	Detail6            string    `gorm:"column:Detail6;type:varchar(150)" json:"detail6"`
	Detail7            string    `gorm:"column:Detail7;type:varchar(150)" json:"detail7"`
	Detail8            string    `gorm:"column:Detail8;type:varchar(150)" json:"detail8"`
	Detail9            string    `gorm:"column:Detail9;type:varchar(150)" json:"detail9"`

	DivisionTransfer string `gorm:"-" json:"divisionTransfer,omitempty"`
	GroupDepTransfer string `gorm:"-" json:"groupDepTransfer,omitempty"`
	EmailTransfer    string `gorm:"-" json:"emailTransfer,omitempty"`
}

func (HANDOVER) TableName() string {
	return "HANDOVER"
}

type HANDOVEROTHER struct {
	ID           string   `gorm:"column:ID;primaryKey;type:varchar(50)" json:"id"`
	UnitCode     string   `gorm:"column:UnitCode;type:varchar(21)" json:"unitCode"`
	HandoverDate DateOnly `gorm:"column:HandoverDate" json:"handoverDate"`
	UserCreated  string   `gorm:"column:UserCreated;type:varchar(50)" json:"userCreated"`
	Form         string   `gorm:"column:Form;type:varchar(50)" json:"form"`
	UserA        string   `gorm:"column:UserA;type:varchar(100)" json:"userA"`
	DepartmentA  string   `gorm:"column:DepartmentA;type:varchar(50)" json:"departmentA"`
	OfficeA      string   `gorm:"column:OfficeA;type:varchar(50)" json:"officeA"`
	CompanyA     string   `gorm:"column:CompanyA;type:varchar(50)" json:"companyA"`
	AreaA        string   `gorm:"column:AreaA;type:varchar(50)" json:"areaA"`
	UserB        string   `gorm:"column:UserB;type:varchar(100)" json:"userB"`
	DepartmentB  string   `gorm:"column:DepartmentB;type:varchar(50)" json:"departmentB"`
	OfficeB      string   `gorm:"column:OfficeB;type:varchar(50)" json:"officeB"`
	CompanyB     string   `gorm:"column:CompanyB;type:varchar(50)" json:"companyB"`
	AreaB        string   `gorm:"column:AreaB;type:varchar(50)" json:"areaB"`
	Type         string   `gorm:"column:Type;type:varchar(50)" json:"type"`
	Detail1      string   `gorm:"column:Detail1;type:varchar(150)" json:"detail1"`
	Detail2      string   `gorm:"column:Detail2;type:varchar(150)" json:"detail2"`
	Detail3      string   `gorm:"column:Detail3;type:varchar(150)" json:"detail3"`
	Detail4      string   `gorm:"column:Detail4;type:varchar(150)" json:"detail4"`
	Detail5      string   `gorm:"column:Detail5;type:varchar(150)" json:"detail5"`
	Detail6      string   `gorm:"column:Detail6;type:varchar(150)" json:"detail6"`
	Detail7      string   `gorm:"column:Detail7;type:varchar(150)" json:"detail7"`
	Detail8      string   `gorm:"column:Detail8;type:varchar(150)" json:"detail8"`
	Detail9      string   `gorm:"column:Detail9;type:varchar(150)" json:"detail9"`
	StateA       string   `gorm:"column:StateA;type:varchar(50)" json:"stateA"`
	StateB       string   `gorm:"column:StateB;type:varchar(50)" json:"stateB"`
	StatusA      string   `gorm:"column:StatusA;type:varchar(50)" json:"statusA"`
	StatusB      string   `gorm:"column:StatusB;type:varchar(50)" json:"statusB"`
	HandoverNo   string   `gorm:"column:HandoverNo;type:varchar(50)" json:"handoverNo"`
	JiraCode     string   `gorm:"column:JiraCode;type:varchar(50)" json:"jiraCode"`
}

func (HANDOVEROTHER) TableName() string {
	return "HANDOVEROTHER"
}
