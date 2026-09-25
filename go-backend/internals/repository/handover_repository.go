package repository

import (
	"context"

	"go-backend/internals/core/domains"
	"go-backend/internals/core/dtos"

	"gorm.io/gorm"
)

type HandoverRepository struct {
	db *gorm.DB
}

func NewHandoverRepository(db *gorm.DB) *HandoverRepository {
	return &HandoverRepository{db: db}
}

func (r *HandoverRepository) GetPagedHandovers(ctx context.Context, unitCode, userSession, userName string, isViewer bool, filter *dtos.Filter) ([]domains.HANDOVEROTHER, int64, error) {
	var list []domains.HANDOVEROTHER
	var total int64

	query := r.db.WithContext(ctx).Model(&domains.HANDOVEROTHER{}).
		Where("\"UnitCode\" = ?", unitCode)

	if isViewer {
		// Viewer can see handovers where UserB matches their UserID or UserName
		query = query.Where("(\"UserB\" IS NOT NULL AND (\"UserB\" LIKE ? OR \"UserB\" LIKE ?))",
			"%"+userSession+"%", "%"+userName+"%")
	}

	query = query.Order("\"HandoverDate\" DESC, \"ID\" DESC")

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	pageIndex := 1
	pageSize := 10
	if filter != nil {
		if filter.PageIndex > 0 {
			pageIndex = filter.PageIndex
		}
		if filter.PageSize > 0 {
			pageSize = filter.PageSize
		}
	}

	offset := (pageIndex - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Find(&list).Error

	return list, total, err
}

func (r *HandoverRepository) FindByID(ctx context.Context, id string) (*domains.HANDOVER, error) {
	var handover domains.HANDOVER
	err := r.db.WithContext(ctx).Where("\"ID\" = ?", id).First(&handover).Error
	if err != nil {
		return nil, err
	}
	return &handover, nil
}

func (r *HandoverRepository) FindOtherByID(ctx context.Context, id string) (*domains.HANDOVEROTHER, error) {
	var handover domains.HANDOVEROTHER
	err := r.db.WithContext(ctx).Where("\"ID\" = ?", id).First(&handover).Error
	if err != nil {
		return nil, err
	}
	return &handover, nil
}
