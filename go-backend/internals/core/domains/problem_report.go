package domains

type PROBLEMREPORT struct {
	ID               string   `gorm:"column:ID;primaryKey;type:varchar(50)" json:"id"`
	UnitCode         string   `gorm:"column:UnitCode;primaryKey;type:varchar(21)" json:"unitCode"`
	ReportDate       DateOnly `gorm:"column:ReportDate" json:"reportDate"`
	UserCreated      string   `gorm:"column:UserCreated;type:varchar(50)" json:"userCreated"`
	ProblemSummary   string   `gorm:"column:ProblemSummary;type:text" json:"problemSummary"`
	TestProcedures   string   `gorm:"column:TestProcedures;type:text" json:"testProcedures"`
	Solutions        string   `gorm:"column:Solutions;type:text" json:"solutions"`
	Type             string   `gorm:"column:Type;type:varchar(50)" json:"type"`
	Detail1          string   `gorm:"column:Detail1;type:varchar(150)" json:"detail1"`
	Detail2          string   `gorm:"column:Detail2;type:varchar(150)" json:"detail2"`
	Detail3          string   `gorm:"column:Detail3;type:varchar(150)" json:"detail3"`
	Detail4          string   `gorm:"column:Detail4;type:varchar(150)" json:"detail4"`
	Detail5          string   `gorm:"column:Detail5;type:varchar(150)" json:"detail5"`
	Detail6          string   `gorm:"column:Detail6;type:varchar(150)" json:"detail6"`
	Detail7          string   `gorm:"column:Detail7;type:varchar(150)" json:"detail7"`
	Detail8          string   `gorm:"column:Detail8;type:varchar(150)" json:"detail8"`
	Detail9          string   `gorm:"column:Detail9;type:varchar(150)" json:"detail9"`
	UserReport       string   `gorm:"column:UserReport;type:varchar(100)" json:"userReport"`
	DepartmentReport string   `gorm:"column:DepartmentReport;type:varchar(50)" json:"departmentReport"`
	OfficeReport     string   `gorm:"column:OfficeReport;type:varchar(50)" json:"officeReport"`
	CompanyReport    string   `gorm:"column:CompanyReport;type:varchar(50)" json:"companyReport"`
	AreaReport       string   `gorm:"column:AreaReport;type:varchar(50)" json:"areaReport"`
	ProblemNo        string   `gorm:"column:ProblemNo;type:varchar(50)" json:"problemNo"`
	JiraCode         string   `gorm:"column:JiraCode;type:varchar(50)" json:"jiraCode"`
}

func (PROBLEMREPORT) TableName() string {
	return "PROBLEMREPORT"
}

type IMAGEREPORT struct {
	ID         string `gorm:"column:ID;primaryKey;type:varchar(9)" json:"id"`
	TypeReport string `gorm:"column:TypeReport;primaryKey;type:varchar(20)" json:"typeReport"`
	Image1     []byte `gorm:"column:Image1" json:"image1,omitempty"`
	Image2     []byte `gorm:"column:Image2" json:"image2,omitempty"`
}

func (IMAGEREPORT) TableName() string {
	return "IMAGEREPORT"
}
