package ports

import (
	"io"
	"time"

	"go-backend/internals/core/domains"
)

type SSOProvider interface {
	GetSSOUserInfo(authCode string) (map[string]any, string, error)
}

type PasswordHasherPort interface {
	HashPassword(password string) (string, error)
	VerifyPassword(hashedPassword, password string) bool
}

type CacheProvider interface {
	Get(key string) (any, bool)
	Set(key string, val any, ttl time.Duration)
	Remove(key string)
}

type HandoverPDFData struct {
	Handover     domains.HANDOVEROTHER
	Asset        domains.ASSET
	UserAFull    string
	UserBFull    string
	DeptAName    string
	DeptBName    string
	OfficeAName  string
	OfficeBName  string
	CompanyAName string
	CompanyBName string
	SpecLines    []string
	Accessories  []domains.ACCESSORY
	Softwares    []domains.SOFTWARE
}

type ProblemReportPDFData struct {
	Report          domains.PROBLEMREPORT
	Asset           domains.ASSET
	UserCreatedFull string
	UserReportFull  string
	DeptName        string
	OfficeName      string
	CompanyName     string
	SpecLines       []string
}

type PDFGenerator interface {
	GenerateHandoverPDF(data *HandoverPDFData) ([]byte, error)
	GenerateProblemReportPDF(data *ProblemReportPDFData) ([]byte, error)
}

type AssetImportRow struct {
	RowNumber     int
	RawCells      []string
	Type          string
	Brand         string
	Model         string
	Configuration string
	Serial        string
	PurchasedDate *domains.DateOnly
	Vendor        string
	Status        string
	UserName      string
	HRCode        string
	Department    string
	GroupDep      string
	Division      string
	Office        string
	Company       string
	Area          string
	IDTag         string
	Notes         string
	State         string
}

type AssetImportErrorRow struct {
	Row         AssetImportRow
	ErrorReason string
}

type ExcelProcessor interface {
	ExportAssetsExcel(
		assets []domains.ASSET,
		deptMap map[string]string,
		deviceConfigMap map[string]domains.DEVICECONFIG,
		itemMap map[string]string,
	) ([]byte, error)
	ParseAssetsImportExcel(reader io.Reader) ([]AssetImportRow, error)
	GenerateAssetsErrorExcel(failedRows []AssetImportErrorRow) ([]byte, error)
	ExportAssetsLogExcel(logs []domains.ASSETSLOG) ([]byte, error)
	ParseDisposalUnitCodesExcel(reader io.Reader) ([]string, error)
}
