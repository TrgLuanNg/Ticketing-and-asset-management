package repository

import (
	"context"
	"strconv"
	"strings"

	"go-backend/internals/core/domains"
	"go-backend/internals/core/dtos"

	"gorm.io/gorm"
)

type HandoverRepository struct {
	*GenericRepository[domains.HANDOVER]
	db *gorm.DB
}

func NewHandoverRepository(db *gorm.DB) *HandoverRepository {
	return &HandoverRepository{
		GenericRepository: NewGenericRepository[domains.HANDOVER](db, "ID", []string{"ID", "UnitCode", "UserCreated", "UserTransfer"}, "\"HandoverDate\" DESC, \"ID\" DESC"),
		db:                db,
	}
}

func (r *HandoverRepository) GetPagedHandovers(ctx context.Context, unitCode, userSession, userName string, isViewer bool, filter *dtos.Filter) ([]domains.HANDOVEROTHER, int64, error) {
	var list []domains.HANDOVEROTHER
	var total int64

	query := getDB(ctx, r.db).Model(&domains.HANDOVEROTHER{}).
		Where("\"UnitCode\" = ?", unitCode)

	if isViewer {
		query = query.Where("(\"UserB\" IS NOT NULL AND (LOWER(\"UserB\") LIKE LOWER(?) OR LOWER(\"UserB\") LIKE LOWER(?)))",
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

func (r *HandoverRepository) FindAll(ctx context.Context) ([]domains.HANDOVER, error) {
	return r.GenericRepository.FindAll(ctx)
}

func (r *HandoverRepository) FindByID(ctx context.Context, id string) (*domains.HANDOVER, error) {
	return r.GenericRepository.FindByID(ctx, id)
}

func (r *HandoverRepository) FindOtherByID(ctx context.Context, id, unitCode string) (*domains.HANDOVEROTHER, error) {
	var handover domains.HANDOVEROTHER
	q := getDB(ctx, r.db).Where("\"ID\" = ?", id)
	if strings.TrimSpace(unitCode) != "" {
		q = q.Where("\"UnitCode\" = ?", unitCode)
	}
	err := q.First(&handover).Error
	if err != nil {
		return nil, err
	}
	return &handover, nil
}

func (r *HandoverRepository) CreateHandover(ctx context.Context, h *domains.HANDOVER) error {
	return getDB(ctx, r.db).Create(h).Error
}

func (r *HandoverRepository) UpdateHandover(ctx context.Context, id string, h *domains.HANDOVER) error {
	return r.GenericRepository.Update(ctx, id, h)
}

func (r *HandoverRepository) DeleteHandover(ctx context.Context, id string) error {
	return r.GenericRepository.Delete(ctx, id)
}

func (r *HandoverRepository) CreateHandoverOther(ctx context.Context, ho *domains.HANDOVEROTHER) error {
	return getDB(ctx, r.db).Create(ho).Error
}

func (r *HandoverRepository) GetMaxHandoverOtherID(ctx context.Context) (int64, error) {
	var ids []string
	err := getDB(ctx, r.db).Model(&domains.HANDOVEROTHER{}).Pluck("\"ID\"", &ids).Error
	if err != nil {
		return 0, err
	}
	var maxID int64
	for _, idStr := range ids {
		if v, err := strconv.ParseInt(strings.TrimSpace(idStr), 10, 64); err == nil && v > maxID {
			maxID = v
		}
	}
	return maxID, nil
}

func (r *HandoverRepository) CountHandoverOthersByPrefix(ctx context.Context, prefix string) (int64, error) {
	var count int64
	err := getDB(ctx, r.db).Model(&domains.HANDOVEROTHER{}).
		Where("\"HandoverNo\" LIKE ?", prefix+"%").
		Count(&count).Error
	return count, err
}

func (r *HandoverRepository) UpsertUpdateEmployeeCode(ctx context.Context, uec *domains.UpdateEmployeeCode) error {
	if uec == nil || strings.TrimSpace(uec.UnitCode) == "" {
		return nil
	}
	db := getDB(ctx, r.db)
	var existing domains.UpdateEmployeeCode
	if err := db.Where("\"UnitCode\" = ?", uec.UnitCode).First(&existing).Error; err != nil {
		return db.Create(uec).Error
	}
	return db.Model(&domains.UpdateEmployeeCode{}).
		Where("\"UnitCode\" = ?", uec.UnitCode).
		Update("HRCode", uec.HRCode).Error
}

func (r *HandoverRepository) GetAccessoriesByUnitCode(ctx context.Context, unitCode string) ([]domains.ACCESSORY, error) {
	var list []domains.ACCESSORY
	err := getDB(ctx, r.db).Where("\"UnitCode\" = ?", unitCode).Find(&list).Error
	return list, err
}

func (r *HandoverRepository) GetSoftwaresByUnitCode(ctx context.Context, unitCode string) ([]domains.SOFTWARE, error) {
	var list []domains.SOFTWARE
	err := getDB(ctx, r.db).Where("\"UnitCode\" = ?", unitCode).Find(&list).Error
	return list, err
}
