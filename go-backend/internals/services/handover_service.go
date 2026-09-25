package services

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"

	"go-backend/internals/core/domains"
	"go-backend/internals/core/dtos"
	"go-backend/internals/repository"

	"github.com/jung-kurt/gofpdf"
	"gorm.io/gorm"
)

type HandoverService struct {
	handoverRepo *repository.HandoverRepository
	userRepo     *repository.UserRepository
	db           *gorm.DB
}

func NewHandoverService(handoverRepo *repository.HandoverRepository, userRepo *repository.UserRepository, db *gorm.DB) *HandoverService {
	return &HandoverService{
		handoverRepo: handoverRepo,
		userRepo:     userRepo,
		db:           db,
	}
}

func (s *HandoverService) GetPagedHandovers(ctx context.Context, unitCode, userSession string, filter *dtos.Filter) ([]dtos.HandoverDto, int64, error) {
	userInfo, _ := s.userRepo.FindByUserID(ctx, userSession)
	isViewer := userInfo != nil && strings.EqualFold(userInfo.Role, "viewer")
	userName := ""
	if userInfo != nil {
		userName = userInfo.UserName
	}

	entities, total, err := s.handoverRepo.GetPagedHandovers(ctx, unitCode, userSession, userName, isViewer, filter)
	if err != nil {
		return nil, 0, err
	}

	dtosList := make([]dtos.HandoverDto, len(entities))
	for i, h := range entities {
		dtosList[i] = dtos.HandoverDto{
			ID:                 h.ID,
			UnitCode:           h.UnitCode,
			HandoverDate:       h.HandoverDate,
			UserCreated:        h.UserCreated,
			Form:               h.Form,
			UserTransfer:       h.UserB,
			DepartmentTransfer: h.DepartmentB,
			OfficeTransfer:     h.OfficeB,
			CompanyTransfer:    h.CompanyB,
			AreaTransfer:       h.AreaB,
			Type:               h.Type,
			JiraCode:           h.JiraCode,
			Detail1:            h.Detail1,
			Detail2:            h.Detail2,
			Detail3:            h.Detail3,
			Detail4:            h.Detail4,
			Detail5:            h.Detail5,
			Detail6:            h.Detail6,
			Detail7:            h.Detail7,
			Detail8:            h.Detail8,
			Detail9:            h.Detail9,
		}
	}

	return dtosList, total, nil
}

// ExportHandoverForm generates a PDF handover receipt matching C# ExportHandoverForm
func (s *HandoverService) ExportHandoverForm(ctx context.Context, handoverID, unitCode string) ([]byte, error) {
	handover, err := s.handoverRepo.FindOtherByID(ctx, handoverID)
	if err != nil {
		return nil, errors.New("handover record not found")
	}

	var asset domains.ASSET
	_ = s.db.WithContext(ctx).Where("\"UnitCode\" = ?", unitCode).First(&asset).Error

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.AddPage()

	// Title
	pdf.SetFont("Arial", "B", 14)
	pdf.CellFormat(0, 8, "BIEN BAN BAN GIAO THIET BI / ASSET HANDOVER RECEIPT", "", 1, "C", false, 0, "")
	pdf.SetFont("Arial", "I", 10)
	pdf.CellFormat(0, 6, fmt.Sprintf("Ma ban giao / Handover ID: %s | Ma thiet bi / Unit Code: %s", handover.ID, unitCode), "", 1, "C", false, 0, "")
	pdf.Ln(6)

	// Section 1: Thong tin ban giao
	pdf.SetFont("Arial", "B", 11)
	pdf.CellFormat(0, 7, "1. THONG TIN BAN GIAO / HANDOVER DETAILS", "B", 1, "L", false, 0, "")
	pdf.SetFont("Arial", "", 10)
	dateStr := handover.HandoverDate.Format("02/01/2006")
	pdf.CellFormat(95, 6, fmt.Sprintf("Ngay ban giao / Date: %s", dateStr), "", 0, "L", false, 0, "")
	pdf.CellFormat(95, 6, fmt.Sprintf("Nguoi tao / Created By: %s", handover.UserCreated), "", 1, "L", false, 0, "")

	pdf.CellFormat(95, 6, fmt.Sprintf("Ben giao / Handed Over By (A): %s", handover.UserA), "", 0, "L", false, 0, "")
	pdf.CellFormat(95, 6, fmt.Sprintf("Ben nhan / Received By (B): %s", handover.UserB), "", 1, "L", false, 0, "")

	pdf.CellFormat(95, 6, fmt.Sprintf("Phong ban / Department: %s", handover.DepartmentB), "", 0, "L", false, 0, "")
	pdf.CellFormat(95, 6, fmt.Sprintf("Van phong / Office: %s", handover.OfficeB), "", 1, "L", false, 0, "")

	pdf.CellFormat(95, 6, fmt.Sprintf("Cong ty / Company: %s", handover.CompanyB), "", 0, "L", false, 0, "")
	pdf.CellFormat(95, 6, fmt.Sprintf("Khu vuc / Area: %s", handover.AreaB), "", 1, "L", false, 0, "")
	pdf.Ln(4)

	// Section 2: Thong tin thiet bi
	pdf.SetFont("Arial", "B", 11)
	pdf.CellFormat(0, 7, "2. THONG TIN THIET BI / ASSET SPECIFICATIONS", "B", 1, "L", false, 0, "")
	pdf.SetFont("Arial", "", 10)
	pdf.CellFormat(95, 6, fmt.Sprintf("Loai / Type: %s", asset.Type), "", 0, "L", false, 0, "")
	pdf.CellFormat(95, 6, fmt.Sprintf("Hang san xuat / Brand: %s", asset.Brand), "", 1, "L", false, 0, "")

	pdf.CellFormat(95, 6, fmt.Sprintf("Mau / Model: %s", asset.Model), "", 0, "L", false, 0, "")
	pdf.CellFormat(95, 6, fmt.Sprintf("So serial / Serial Number: %s", asset.Serial), "", 1, "L", false, 0, "")

	if asset.Detail1 != "" {
		pdf.CellFormat(95, 6, fmt.Sprintf("CPU / Processor: %s", asset.Detail1), "", 0, "L", false, 0, "")
	}
	if asset.Detail2 != "" {
		pdf.CellFormat(95, 6, fmt.Sprintf("RAM: %s", asset.Detail2), "", 1, "L", false, 0, "")
	}
	if asset.Detail3 != "" {
		pdf.CellFormat(95, 6, fmt.Sprintf("Hard Drive / O cung: %s", asset.Detail3), "", 0, "L", false, 0, "")
	}
	if asset.Detail4 != "" {
		pdf.CellFormat(95, 6, fmt.Sprintf("Man hinh / Screen: %s", asset.Detail4), "", 1, "L", false, 0, "")
	}
	pdf.Ln(6)

	// Signatures
	pdf.Ln(10)
	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(95, 6, "NGUOI GIAO / HANDED OVER BY", "", 0, "C", false, 0, "")
	pdf.CellFormat(95, 6, "NGUOI NHAN / RECEIVED BY", "", 1, "C", false, 0, "")
	pdf.SetFont("Arial", "I", 9)
	pdf.CellFormat(95, 5, "(Ky va ghi ro ho ten / Sign & print name)", "", 0, "C", false, 0, "")
	pdf.CellFormat(95, 5, "(Ky va ghi ro ho ten / Sign & print name)", "", 1, "C", false, 0, "")

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
