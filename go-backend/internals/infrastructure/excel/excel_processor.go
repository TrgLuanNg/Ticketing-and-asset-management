package excel

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"go-backend/internals/core/domains"
	"go-backend/internals/core/ports"

	"github.com/xuri/excelize/v2"
)

type ExcelProcessor struct {
	templateDir string
}

func NewExcelProcessor(templateDir ...string) *ExcelProcessor {
	dir := "Templates"
	if len(templateDir) > 0 && templateDir[0] != "" {
		dir = templateDir[0]
	}
	return &ExcelProcessor{templateDir: dir}
}

// ExportAssetsExcel exports assets using Templates/templateAssetsExport.xlsx (23 columns starting at row 3)
// matching .NET AssetService.Export.
func (p *ExcelProcessor) ExportAssetsExcel(
	assets []domains.ASSET,
	deptMap map[string]string,
	deviceConfigMap map[string]domains.DEVICECONFIG,
	itemMap map[string]string,
) ([]byte, error) {
	if len(assets) == 0 {
		return nil, errors.New("no data to export")
	}

	templatePath := p.templateDir + "/templateAssetsExport.xlsx"
	var f *excelize.File
	var err error
	sheet := "Sheet1"

	if _, statErr := os.Stat(templatePath); statErr == nil {
		f, err = excelize.OpenFile(templatePath)
		if err != nil {
			return nil, err
		}
		sheets := f.GetSheetList()
		if len(sheets) > 0 {
			sheet = sheets[0]
		}
	} else {
		f = excelize.NewFile()
		sheet = "Assets"
		f.SetSheetName("Sheet1", sheet)
		headers := []string{
			"No.", "Unit Code", "Asset Code", "Fin Code", "Term", "Type", "Brand", "Model",
			"Configuration", "Serial", "Purchased Date", "Vendor", "Status", "User Name",
			"HR Code", "Department", "Group Dep", "Division", "Office", "Company", "Area", "Host Name", "Notes",
		}
		for colIdx, h := range headers {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, 2)
			_ = f.SetCellValue(sheet, cell, h)
		}
	}
	defer f.Close()

	resolveDept := func(code string) string {
		if code == "" {
			return ""
		}
		if name, ok := deptMap[strings.ToLower(strings.TrimSpace(code))]; ok && name != "" {
			return name
		}
		return code
	}

	buildConfig := func(a domains.ASSET) string {
		cfg, ok := deviceConfigMap[strings.ToUpper(strings.TrimSpace(a.Type))]
		if !ok {
			return ""
		}
		details := []struct {
			label string
			val   string
		}{
			{cfg.Detail1, a.Detail1},
			{cfg.Detail2, a.Detail2},
			{cfg.Detail3, a.Detail3},
			{cfg.Detail4, a.Detail4},
			{cfg.Detail5, a.Detail5},
			{cfg.Detail6, a.Detail6},
			{cfg.Detail7, a.Detail7},
			{cfg.Detail8, a.Detail8},
			{cfg.Detail9, a.Detail9},
		}
		var parts []string
		for _, d := range details {
			if strings.TrimSpace(d.label) != "" && strings.TrimSpace(d.val) != "" {
				specVal := d.val
				if resolved, found := itemMap[strings.ToLower(strings.TrimSpace(d.val))]; found && resolved != "" {
					specVal = resolved
				}
				parts = append(parts, fmt.Sprintf("%s: %s", strings.TrimSpace(d.label), specVal))
			}
		}
		return strings.Join(parts, " / ")
	}

	for idx, a := range assets {
		row := idx + 3
		purchasedStr := ""
		if a.PurchasedDate != nil && !a.PurchasedDate.IsZero() {
			purchasedStr = a.PurchasedDate.Format("02/01/2006")
		}

		values := []any{
			idx + 1,
			a.UnitCode,
			a.IDTag,
			"",
			a.TypeEXP,
			a.Type,
			a.Brand,
			a.Model,
			buildConfig(a),
			a.Serial,
			purchasedStr,
			resolveDept(a.Vendor),
			a.Status,
			a.UserName,
			a.HRCode,
			resolveDept(a.Department),
			resolveDept(a.GroupDep),
			resolveDept(a.Division),
			resolveDept(a.Office),
			resolveDept(a.Company),
			a.Area,
			a.IDTag,
			a.Notes,
		}

		for colIdx, val := range values {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, row)
			_ = f.SetCellValue(sheet, cell, val)
		}
	}

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// ParseAssetsImportExcel parses uploaded Excel files matching the 23-column AssetsCatalouges.xlsx template.
func (p *ExcelProcessor) ParseAssetsImportExcel(reader io.Reader) ([]ports.AssetImportRow, error) {
	f, err := excelize.OpenReader(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to open excel reader: %w", err)
	}
	defer f.Close()

	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, errors.New("excel file contains no sheets")
	}

	rows, err := f.GetRows(sheets[0])
	if err != nil {
		return nil, err
	}

	if len(rows) < 2 {
		return nil, errors.New("excel file is empty")
	}

	startRow := 2 // 0-indexed row 2 = Excel row 3 (standard template)
	if len(rows) == 2 {
		startRow = 1
	} else if len(rows) > 1 {
		// Detect if row 2 (index 1) is header or data
		r1 := rows[1]
		if len(r1) > 0 && !strings.EqualFold(strings.TrimSpace(r1[0]), "No.") && !strings.EqualFold(strings.TrimSpace(r1[0]), "STT") {
			startRow = 1
		}
	}

	getCell := func(row []string, idx int) string {
		if idx >= 0 && idx < len(row) {
			return strings.TrimSpace(row[idx])
		}
		return ""
	}

	var parsed []ports.AssetImportRow
	for i := startRow; i < len(rows); i++ {
		row := rows[i]
		// Check if row is completely blank
		nonEmpty := false
		for _, c := range row {
			if strings.TrimSpace(c) != "" {
				nonEmpty = true
				break
			}
		}
		if !nonEmpty {
			continue
		}

		rawCells := make([]string, 23)
		for cIdx := 0; cIdx < 23; cIdx++ {
			rawCells[cIdx] = getCell(row, cIdx)
		}

		var item ports.AssetImportRow
		item.RowNumber = i + 1
		item.RawCells = rawCells

		// Support both 23-column template (col 6 = Type, col 7 = Brand, col 8 = Model, col 9 = Config, col 10 = Serial...)
		// and compact 4+ column sheets
		if len(row) >= 10 {
			item.IDTag = getCell(row, 2)
			item.Type = getCell(row, 5)
			item.Brand = getCell(row, 6)
			item.Model = getCell(row, 7)
			item.Configuration = getCell(row, 8)
			item.Serial = getCell(row, 9)
			dateStr := getCell(row, 10)
			if dateStr != "" {
				if d := parseFlexibleDate(dateStr); d != nil {
					item.PurchasedDate = d
				}
			}
			item.Vendor = getCell(row, 11)
			item.Status = getCell(row, 12)
			item.UserName = getCell(row, 13)
			item.HRCode = getCell(row, 14)
			item.Department = getCell(row, 15)
			item.GroupDep = getCell(row, 16)
			item.Division = getCell(row, 17)
			item.Office = getCell(row, 18)
			item.Company = getCell(row, 19)
			item.Area = getCell(row, 20)
			if host := getCell(row, 21); host != "" && item.IDTag == "" {
				item.IDTag = host
			}
			item.Notes = getCell(row, 22)
		} else {
			// Compact fallback: [UnitCode/IDTag, Type, UserName, Department]
			item.IDTag = getCell(row, 0)
			item.Type = getCell(row, 1)
			item.UserName = getCell(row, 2)
			item.Department = getCell(row, 3)
			item.Brand = "Generic"
			item.Model = "Standard"
			item.Status = "GOOD"
			item.Area = "South"
			item.Company = "YLVN"
		}

		parsed = append(parsed, item)
	}

	return parsed, nil
}

// GenerateAssetsErrorExcel generates a 24-column Excel workbook containing failed import rows and their error reasons.
func (p *ExcelProcessor) GenerateAssetsErrorExcel(failedRows []ports.AssetImportErrorRow) ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()
	sheet := "ImportErrors"
	f.SetSheetName("Sheet1", sheet)

	headers := []string{
		"No.", "Unit Code", "Asset Code", "Fin Code", "Term", "Type", "Brand", "Model",
		"Configuration", "Serial", "Purchased Date", "Vendor", "Status", "User Name",
		"HR Code", "Department", "Group Dep", "Division", "Office", "Company", "Area", "Host Name", "Notes",
		"Lý do lỗi (Error Reason)",
	}

	for colIdx, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(colIdx+1, 1)
		_ = f.SetCellValue(sheet, cell, h)
	}

	for idx, fail := range failedRows {
		rowNum := idx + 2
		for colIdx := 0; colIdx < 23; colIdx++ {
			val := ""
			if colIdx < len(fail.Row.RawCells) {
				val = fail.Row.RawCells[colIdx]
			}
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, rowNum)
			_ = f.SetCellValue(sheet, cell, val)
		}
		errCell, _ := excelize.CoordinatesToCellName(24, rowNum)
		_ = f.SetCellValue(sheet, errCell, fail.ErrorReason)
	}

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// ExportAssetsLogExcel exports asset logs matching .NET AssetsLogService.ExportExcel.
func (p *ExcelProcessor) ExportAssetsLogExcel(logs []domains.ASSETSLOG) ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()
	sheet := "Logs"
	f.SetSheetName("Sheet1", sheet)

	headers := []string{"UnitCode", "UserID", "Action", "ActionLogs", "DateModified"}
	for colIdx, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(colIdx+1, 1)
		_ = f.SetCellValue(sheet, cell, h)
	}

	for idx, item := range logs {
		row := idx + 2
		_ = f.SetCellValue(sheet, fmt.Sprintf("A%d", row), item.UnitCode)
		_ = f.SetCellValue(sheet, fmt.Sprintf("B%d", row), item.UserID)
		_ = f.SetCellValue(sheet, fmt.Sprintf("C%d", row), item.Action)
		_ = f.SetCellValue(sheet, fmt.Sprintf("D%d", row), item.ActionLogs)
		_ = f.SetCellValue(sheet, fmt.Sprintf("E%d", row), item.DateModified.Format("02/01/2006 15:04:05"))
	}

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// ParseDisposalUnitCodesExcel extracts unit codes from column 1 starting at row 2 matching .NET DisposalsController.ImportUnitCodes.
func (p *ExcelProcessor) ParseDisposalUnitCodesExcel(reader io.Reader) ([]string, error) {
	f, err := excelize.OpenReader(reader)
	if err != nil {
		return nil, fmt.Errorf("error parsing excel file: %w", err)
	}
	defer f.Close()

	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, errors.New("excel worksheet not found")
	}

	rows, err := f.GetRows(sheets[0])
	if err != nil {
		return nil, err
	}

	var unitCodes []string
	for i := 1; i < len(rows); i++ {
		if len(rows[i]) > 0 {
			val := strings.TrimSpace(rows[i][0])
			if val != "" {
				unitCodes = append(unitCodes, val)
			}
		}
	}
	return unitCodes, nil
}

func parseFlexibleDate(s string) *domains.DateOnly {
	formats := []string{
		"02/01/2006",
		"2006-01-02",
		"01-02-06",
		"1/2/2006",
		"2/1/2006",
		"02-01-2006",
		"2006/01/02",
	}
	for _, layout := range formats {
		if t, err := time.Parse(layout, strings.TrimSpace(s)); err == nil {
			d := domains.NewDateOnly(t)
			return &d
		}
	}
	return nil
}
