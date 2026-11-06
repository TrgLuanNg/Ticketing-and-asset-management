package pdf

import (
	"bytes"
	"fmt"
	"strings"

	"go-backend/internals/core/ports"

	"github.com/jung-kurt/gofpdf"
)

type PDFGenerator struct{}

func NewPDFGenerator() *PDFGenerator {
	return &PDFGenerator{}
}

// GenerateHandoverPDF renders the 2-page A4 bilingual Handover Form (FRM/IT - 04.01.0213)
func (g *PDFGenerator) GenerateHandoverPDF(data *ports.HandoverPDFData) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(12, 12, 12)
	pdf.SetAutoPageBreak(true, 12)

	// PAGE 1: HANDOVER FORM
	pdf.AddPage()

	// Header Box (Form Code & Title)
	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(50, 14, "YUSEN LOGISTICS", "1", 0, "C", false, 0, "")
	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(90, 14, "ASSET HANDOVER FORM\n(BIEN BAN BAN GIAO THIET BI)", "1", 0, "C", false, 0, "")
	pdf.SetFont("Arial", "", 8)
	dateStr := ""
	if !data.Handover.HandoverDate.IsZero() {
		dateStr = data.Handover.HandoverDate.Format("02/01/2006")
	}
	headerRight := fmt.Sprintf("Form: FRM/IT - 04.01.0213\nNo: %s\nDate: %s", toASCII(data.Handover.HandoverNo), dateStr)
	pdf.MultiCell(46, 4.66, headerRight, "1", "L", false)
	pdf.Ln(3)

	// Section 1: Party A & Party B
	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(93, 6, "PARTY A (BEN GIAO / HANDED OVER BY)", "1", 0, "L", false, 0, "")
	pdf.CellFormat(93, 6, "PARTY B (BEN NHAN / RECEIVED BY)", "1", 1, "L", false, 0, "")

	pdf.SetFont("Arial", "", 9)
	userA := toASCII(data.UserAFull)
	if userA == "" {
		userA = toASCII(data.Handover.UserA)
	}
	userB := toASCII(data.UserBFull)
	if userB == "" {
		userB = toASCII(data.Handover.UserB)
	}

	pdf.CellFormat(93, 6, fmt.Sprintf("Full Name: %s", userA), "LR", 0, "L", false, 0, "")
	pdf.CellFormat(93, 6, fmt.Sprintf("Full Name: %s", userB), "LR", 1, "L", false, 0, "")

	pdf.CellFormat(93, 6, fmt.Sprintf("Department: %s", toASCII(data.DeptAName)), "LR", 0, "L", false, 0, "")
	pdf.CellFormat(93, 6, fmt.Sprintf("Department: %s", toASCII(data.DeptBName)), "LR", 1, "L", false, 0, "")

	pdf.CellFormat(93, 6, fmt.Sprintf("Office / Area: %s / %s", toASCII(data.OfficeAName), toASCII(data.Handover.AreaA)), "LR", 0, "L", false, 0, "")
	pdf.CellFormat(93, 6, fmt.Sprintf("Office / Area: %s / %s", toASCII(data.OfficeBName), toASCII(data.Handover.AreaB)), "LR", 1, "L", false, 0, "")

	pdf.CellFormat(93, 6, fmt.Sprintf("Company: %s", toASCII(data.CompanyAName)), "LRB", 0, "L", false, 0, "")
	pdf.CellFormat(93, 6, fmt.Sprintf("Company: %s", toASCII(data.CompanyBName)), "LRB", 1, "L", false, 0, "")
	pdf.Ln(3)

	// Section 2: Equipment Information
	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(186, 6, "1. EQUIPMENT INFORMATION (THONG TIN THIET BI)", "1", 1, "L", false, 0, "")
	pdf.SetFont("Arial", "", 9)
	pdf.CellFormat(93, 6, fmt.Sprintf("Unit Code: %s", toASCII(data.Asset.UnitCode)), "LR", 0, "L", false, 0, "")
	pdf.CellFormat(93, 6, fmt.Sprintf("Form Type: %s | Jira: %s", toASCII(data.Handover.Form), toASCII(data.Handover.JiraCode)), "LR", 1, "L", false, 0, "")

	pdf.CellFormat(93, 6, fmt.Sprintf("Device Type: %s", toASCII(data.Asset.Type)), "LR", 0, "L", false, 0, "")
	pdf.CellFormat(93, 6, fmt.Sprintf("Brand / Model: %s / %s", toASCII(data.Asset.Brand), toASCII(data.Asset.Model)), "LR", 1, "L", false, 0, "")

	pdf.CellFormat(93, 6, fmt.Sprintf("Serial Number: %s", toASCII(data.Asset.Serial)), "LRB", 0, "L", false, 0, "")
	pdf.CellFormat(93, 6, fmt.Sprintf("Status: %s -> %s", toASCII(data.Handover.StatusA), toASCII(data.Handover.StatusB)), "LRB", 1, "L", false, 0, "")

	if len(data.SpecLines) > 0 {
		pdf.SetFont("Arial", "B", 9)
		pdf.CellFormat(186, 6, "Technical Specifications (Cau hinh chi tiet):", "LR", 1, "L", false, 0, "")
		pdf.SetFont("Arial", "", 8.5)
		for i := 0; i < len(data.SpecLines); i += 2 {
			left := toASCII(data.SpecLines[i])
			right := ""
			if i+1 < len(data.SpecLines) {
				right = toASCII(data.SpecLines[i+1])
			}
			border := "LR"
			if i+2 >= len(data.SpecLines) {
				border = "LRB"
			}
			pdf.CellFormat(93, 5.5, " - "+left, border, 0, "L", false, 0, "")
			pdf.CellFormat(93, 5.5, " "+right, border, 1, "L", false, 0, "")
		}
	}
	pdf.Ln(3)

	// Section 3: Accessories Table
	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(186, 6, "2. ACCESSORIES CHECKLIST (DANH MUC PHU KIEN KEM THEO)", "1", 1, "L", false, 0, "")
	pdf.SetFont("Arial", "B", 8.5)
	pdf.CellFormat(15, 6, "No.", "1", 0, "C", false, 0, "")
	pdf.CellFormat(55, 6, "Brand", "1", 0, "L", false, 0, "")
	pdf.CellFormat(76, 6, "Description", "1", 0, "L", false, 0, "")
	pdf.CellFormat(40, 6, "Serial", "1", 1, "L", false, 0, "")
	pdf.SetFont("Arial", "", 8.5)
	if len(data.Accessories) == 0 {
		pdf.CellFormat(186, 6, "Standard accessories included (Adapter / Mouse / Bag if applicable)", "1", 1, "C", false, 0, "")
	} else {
		for idx, acc := range data.Accessories {
			pdf.CellFormat(15, 5.5, fmt.Sprintf("%d", idx+1), "1", 0, "C", false, 0, "")
			pdf.CellFormat(55, 5.5, toASCII(acc.Brand), "1", 0, "L", false, 0, "")
			pdf.CellFormat(76, 5.5, toASCII(acc.Description), "1", 0, "L", false, 0, "")
			pdf.CellFormat(40, 5.5, toASCII(acc.Serial), "1", 1, "L", false, 0, "")
		}
	}
	pdf.Ln(3)

	// Section 4: Software Table
	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(186, 6, "3. INSTALLED SOFTWARE CHECKLIST (PHAN MEM CAI DAT)", "1", 1, "L", false, 0, "")
	pdf.SetFont("Arial", "B", 8.5)
	pdf.CellFormat(15, 6, "No.", "1", 0, "C", false, 0, "")
	pdf.CellFormat(121, 6, "Software Name", "1", 0, "L", false, 0, "")
	pdf.CellFormat(50, 6, "Version", "1", 1, "L", false, 0, "")
	pdf.SetFont("Arial", "", 8.5)
	if len(data.Softwares) == 0 {
		pdf.CellFormat(186, 6, "Standard corporate OS & office software package", "1", 1, "C", false, 0, "")
	} else {
		for idx, sw := range data.Softwares {
			pdf.CellFormat(15, 5.5, fmt.Sprintf("%d", idx+1), "1", 0, "C", false, 0, "")
			pdf.CellFormat(121, 5.5, toASCII(sw.Software), "1", 0, "L", false, 0, "")
			pdf.CellFormat(50, 5.5, toASCII(sw.Version), "1", 1, "L", false, 0, "")
		}
	}
	pdf.Ln(6)

	// Section 5: 5-Column Signatures
	pdf.SetFont("Arial", "B", 8.5)
	colW := 37.2
	pdf.CellFormat(colW, 6, "PARTY A (Ben giao)", "1", 0, "C", false, 0, "")
	pdf.CellFormat(colW, 6, "DEPT MANAGER A", "1", 0, "C", false, 0, "")
	pdf.CellFormat(colW, 6, "PARTY B (Ben nhan)", "1", 0, "C", false, 0, "")
	pdf.CellFormat(colW, 6, "DEPT MANAGER B", "1", 0, "C", false, 0, "")
	pdf.CellFormat(colW, 6, "IT DEPARTMENT", "1", 1, "C", false, 0, "")

	for i := 0; i < 5; i++ {
		ln := 0
		if i == 4 {
			ln = 1
		}
		pdf.CellFormat(colW, 22, "", "1", ln, "C", false, 0, "")
	}

	// PAGE 2: ACCEPTABLE USE POLICY
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(0, 8, "ACCEPTABLE USE POLICY / THOA THUAN CHAP NHAN SU DUNG", "", 1, "C", false, 0, "")
	pdf.Ln(4)
	pdf.SetFont("Arial", "", 9.5)
	policyText := strings.Join([]string{
		"1. Equipment Ownership & Responsibility:",
		"   - All IT equipment handed over remains the property of Yusen Logistics (Vietnam).",
		"   - The recipient (Party B) is responsible for safeguarding the equipment against loss, theft, or physical damage.",
		"",
		"2. Software & Security Compliance:",
		"   - Only authorized software installed by the IT Department is permitted on company devices.",
		"   - Users must not disable endpoint security, antivirus (Trend Micro), or patch management agents (ManageEngine).",
		"",
		"3. Return & Transfer Procedure:",
		"   - Employees must return all IT equipment to the IT Department upon resignation, transfer, or request.",
		"   - Direct transfer of equipment between employees without IT Department registration and inspection is prohibited.",
		"",
		"4. Incident Reporting:",
		"   - Any hardware malfunction, security incident, or loss must be reported immediately to the IT Helpdesk.",
	}, "\n")
	pdf.MultiCell(0, 5.5, policyText, "", "L", false)
	pdf.Ln(12)
	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(93, 6, "CONFIRMED BY RECIPIENT (PARTY B)", "", 0, "C", false, 0, "")
	pdf.CellFormat(93, 6, "CONFIRMED BY IT DEPARTMENT", "", 1, "C", false, 0, "")

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// GenerateProblemReportPDF renders the 2-page A4 bilingual Problem Report Form (FRM/IT - 07.01.0213)
func (g *PDFGenerator) GenerateProblemReportPDF(data *ports.ProblemReportPDFData) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(12, 12, 12)
	pdf.SetAutoPageBreak(true, 12)

	// PAGE 1: PROBLEM REPORT FORM
	pdf.AddPage()

	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(50, 14, "YUSEN LOGISTICS", "1", 0, "C", false, 0, "")
	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(90, 14, "IT PROBLEM REPORT FORM\n(BIEN BAN KIEM TRA THIET BI)", "1", 0, "C", false, 0, "")
	pdf.SetFont("Arial", "", 8)
	dateStr := ""
	if !data.Report.ReportDate.IsZero() {
		dateStr = data.Report.ReportDate.Format("02/01/2006")
	}
	headerRight := fmt.Sprintf("Form: FRM/IT - 07.01.0213\nNo: %s\nDate: %s", toASCII(data.Report.ProblemNo), dateStr)
	pdf.MultiCell(46, 4.66, headerRight, "1", "L", false)
	pdf.Ln(4)

	// Reporter & Asset Info
	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(186, 6, "1. USER & EQUIPMENT DETAILS (THONG TIN NGUOI DUNG & THIET BI)", "1", 1, "L", false, 0, "")
	pdf.SetFont("Arial", "", 9)
	pdf.CellFormat(93, 6, fmt.Sprintf("Report ID: %s | Jira: %s", toASCII(data.Report.ID), toASCII(data.Report.JiraCode)), "LR", 0, "L", false, 0, "")
	pdf.CellFormat(93, 6, fmt.Sprintf("Unit Code: %s (%s)", toASCII(data.Report.UnitCode), toASCII(data.Report.Type)), "LR", 1, "L", false, 0, "")

	pdf.CellFormat(93, 6, fmt.Sprintf("User / Reporter: %s", toASCII(data.UserReportFull)), "LR", 0, "L", false, 0, "")
	pdf.CellFormat(93, 6, fmt.Sprintf("Inspected By (IT): %s", toASCII(data.UserCreatedFull)), "LR", 1, "L", false, 0, "")

	pdf.CellFormat(93, 6, fmt.Sprintf("Department / Office: %s / %s", toASCII(data.DeptName), toASCII(data.OfficeName)), "LR", 0, "L", false, 0, "")
	pdf.CellFormat(93, 6, fmt.Sprintf("Company / Area: %s / %s", toASCII(data.CompanyName), toASCII(data.Report.AreaReport)), "LR", 1, "L", false, 0, "")

	pdf.CellFormat(93, 6, fmt.Sprintf("Brand / Model: %s / %s", toASCII(data.Asset.Brand), toASCII(data.Asset.Model)), "LRB", 0, "L", false, 0, "")
	pdf.CellFormat(93, 6, fmt.Sprintf("Serial Number: %s", toASCII(data.Asset.Serial)), "LRB", 1, "L", false, 0, "")

	if len(data.SpecLines) > 0 {
		pdf.SetFont("Arial", "B", 9)
		pdf.CellFormat(186, 6, "Technical Specifications:", "LR", 1, "L", false, 0, "")
		pdf.SetFont("Arial", "", 8.5)
		for i := 0; i < len(data.SpecLines); i += 2 {
			left := toASCII(data.SpecLines[i])
			right := ""
			if i+1 < len(data.SpecLines) {
				right = toASCII(data.SpecLines[i+1])
			}
			border := "LR"
			if i+2 >= len(data.SpecLines) {
				border = "LRB"
			}
			pdf.CellFormat(93, 5.5, " - "+left, border, 0, "L", false, 0, "")
			pdf.CellFormat(93, 5.5, " "+right, border, 1, "L", false, 0, "")
		}
	}
	pdf.Ln(4)

	// Problem Summary
	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(186, 6, "2. PROBLEM SUMMARY (MO TA SU CO)", "1", 1, "L", false, 0, "")
	pdf.SetFont("Arial", "", 9)
	pdf.MultiCell(186, 6, toASCII(defaultText(data.Report.ProblemSummary, "N/A")), "1", "L", false)
	pdf.Ln(3)

	// Test Procedures
	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(186, 6, "3. DIAGNOSTIC & TEST PROCEDURES (QUY TRINH KIEM TRA)", "1", 1, "L", false, 0, "")
	pdf.SetFont("Arial", "", 9)
	pdf.MultiCell(186, 6, toASCII(defaultText(data.Report.TestProcedures, "N/A")), "1", "L", false)
	pdf.Ln(3)

	// Solutions
	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(186, 6, "4. PROPOSED SOLUTION / CONCLUSION (HUONG XU LY & KET LUAN)", "1", 1, "L", false, 0, "")
	pdf.SetFont("Arial", "", 9)
	pdf.MultiCell(186, 6, toASCII(defaultText(data.Report.Solutions, "N/A")), "1", "L", false)
	pdf.Ln(8)

	// Signatures
	pdf.SetFont("Arial", "B", 9)
	pdf.CellFormat(62, 6, "REPORTED BY USER", "1", 0, "C", false, 0, "")
	pdf.CellFormat(62, 6, "INSPECTED BY IT", "1", 0, "C", false, 0, "")
	pdf.CellFormat(62, 6, "APPROVED BY MANAGER", "1", 1, "C", false, 0, "")
	pdf.CellFormat(62, 24, "", "1", 0, "C", false, 0, "")
	pdf.CellFormat(62, 24, "", "1", 0, "C", false, 0, "")
	pdf.CellFormat(62, 24, "", "1", 1, "C", false, 0, "")

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func defaultText(s, fallback string) string {
	if strings.TrimSpace(s) == "" {
		return fallback
	}
	return s
}

var vnReplacer = strings.NewReplacer(
	"à", "a", "á", "a", "ả", "a", "ã", "a", "ạ", "a",
	"ă", "a", "ằ", "a", "ắ", "a", "ẳ", "a", "ẵ", "a", "ặ", "a",
	"â", "a", "ầ", "a", "ấ", "a", "ẩ", "a", "ẫ", "a", "ậ", "a",
	"è", "e", "é", "e", "ẻ", "e", "ẽ", "e", "ẹ", "e",
	"ê", "e", "ề", "e", "ế", "e", "ể", "e", "ễ", "e", "ệ", "e",
	"ì", "i", "í", "i", "ỉ", "i", "ĩ", "i", "ị", "i",
	"ò", "o", "ó", "o", "ỏ", "o", "õ", "o", "ọ", "o",
	"ô", "o", "ồ", "o", "ố", "o", "ổ", "o", "ỗ", "o", "ộ", "o",
	"ơ", "o", "ờ", "o", "ớ", "o", "ở", "o", "ỡ", "o", "ợ", "o",
	"ù", "u", "ú", "u", "ủ", "u", "ũ", "u", "ụ", "u",
	"ư", "u", "ừ", "u", "ứ", "u", "ử", "u", "ữ", "u", "ự", "u",
	"ỳ", "y", "ý", "y", "ỷ", "y", "ỹ", "y", "ỵ", "y",
	"đ", "d",
	"À", "A", "Á", "A", "Ả", "A", "Ã", "A", "Ạ", "A",
	"Ă", "A", "Ằ", "A", "Ắ", "A", "Ẳ", "A", "Ẵ", "A", "Ặ", "A",
	"Â", "A", "Ầ", "A", "Ấ", "A", "Ẩ", "A", "Ẫ", "A", "Ậ", "A",
	"È", "E", "É", "E", "Ẻ", "E", "Ẽ", "E", "Ẹ", "E",
	"Ê", "E", "Ề", "E", "Ế", "E", "Ể", "E", "Ễ", "E", "Ệ", "E",
	"Ì", "I", "Í", "I", "Ỉ", "I", "Ĩ", "I", "Ị", "I",
	"Ò", "O", "Ó", "O", "Ỏ", "O", "Õ", "O", "Ọ", "O",
	"Ô", "O", "Ồ", "O", "Ố", "O", "Ổ", "O", "Ỗ", "O", "Ộ", "O",
	"Ơ", "O", "Ờ", "O", "Ớ", "O", "Ở", "O", "Ỡ", "O", "Ợ", "O",
	"Ù", "U", "Ú", "U", "Ủ", "U", "Ũ", "U", "Ụ", "U",
	"Ư", "U", "Ừ", "U", "Ứ", "U", "Ử", "U", "Ữ", "U", "Ự", "U",
	"Ỳ", "Y", "Ý", "Y", "Ỷ", "Y", "Ỹ", "Y", "Ỵ", "Y",
	"Đ", "D",
)

func toASCII(s string) string {
	return vnReplacer.Replace(s)
}
