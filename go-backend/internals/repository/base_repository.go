package repository

import (
	"context"
	"fmt"

	"go-backend/internals/core/dtos"

	"gorm.io/gorm"
)

type GenericRepository[T any] struct {
	db          *gorm.DB
	primaryKey  string
	searchProps []string
}

func NewGenericRepository[T any](db *gorm.DB, primaryKey string, searchProps []string) *GenericRepository[T] {
	return &GenericRepository[T]{
		db:          db,
		primaryKey:  primaryKey,
		searchProps: searchProps,
	}
}

func (r *GenericRepository[T]) DB() *gorm.DB {
	return r.db
}

func (r *GenericRepository[T]) FindPaged(ctx context.Context, filter *dtos.Filter) ([]T, int64, error) {
	var entities []T
	var total int64

	query := r.db.WithContext(ctx).Model(new(T))

	if filter != nil && !filter.SkipCount {
		if err := query.Count(&total).Error; err != nil {
			return nil, 0, err
		}
	}

	if filter != nil && filter.PageSize > 0 {
		pageIndex := filter.PageIndex
		if pageIndex <= 0 {
			pageIndex = 1
		}
		offset := (pageIndex - 1) * filter.PageSize
		query = query.Offset(offset).Limit(filter.PageSize)
	}

	if err := query.Find(&entities).Error; err != nil {
		return nil, 0, err
	}

	return entities, total, nil
}

func (r *GenericRepository[T]) FindAll(ctx context.Context) ([]T, error) {
	var entities []T
	if err := r.db.WithContext(ctx).Find(&entities).Error; err != nil {
		return nil, err
	}
	return entities, nil
}

func (r *GenericRepository[T]) FindByID(ctx context.Context, id any) (*T, error) {
	var entity T
	err := r.db.WithContext(ctx).Where(fmt.Sprintf("\"%s\" = ?", r.primaryKey), id).First(&entity).Error
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

func (r *GenericRepository[T]) Create(ctx context.Context, entity *T) error {
	return r.db.WithContext(ctx).Create(entity).Error
}

func (r *GenericRepository[T]) Update(ctx context.Context, id any, entity *T) error {
	return r.db.WithContext(ctx).Model(new(T)).Where(fmt.Sprintf("\"%s\" = ?", r.primaryKey), id).Updates(entity).Error
}

func (r *GenericRepository[T]) Delete(ctx context.Context, id any) error {
	return r.db.WithContext(ctx).Where(fmt.Sprintf("\"%s\" = ?", r.primaryKey), id).Delete(new(T)).Error
}
