package repository

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"go-backend/internals/core/dtos"

	"gorm.io/gorm"
)

type txKey struct{}

// GormTransactionManager implements ports.TransactionManager using context-bound *gorm.DB.
type GormTransactionManager struct {
	db *gorm.DB
}

func NewTransactionManager(db *gorm.DB) *GormTransactionManager {
	return &GormTransactionManager{db: db}
}

func (tm *GormTransactionManager) RunInTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return getDB(ctx, tm.db).Transaction(func(tx *gorm.DB) error {
		txCtx := context.WithValue(ctx, txKey{}, tx)
		return fn(txCtx)
	})
}

func getDB(ctx context.Context, defaultDB *gorm.DB) *gorm.DB {
	if ctx != nil {
		if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok && tx != nil {
			return tx.WithContext(ctx)
		}
		return defaultDB.WithContext(ctx)
	}
	return defaultDB
}

type GenericRepository[T any] struct {
	db           *gorm.DB
	primaryKey   string
	searchProps  []string
	defaultOrder string
	colMap       map[string]string
}

func NewGenericRepository[T any](db *gorm.DB, primaryKey string, searchProps []string, defaultOrder ...string) *GenericRepository[T] {
	defOrd := ""
	if len(defaultOrder) > 0 {
		defOrd = defaultOrder[0]
	}
	return &GenericRepository[T]{
		db:           db,
		primaryKey:   primaryKey,
		searchProps:  searchProps,
		defaultOrder: defOrd,
		colMap:       buildColumnMap[T](),
	}
}

func buildColumnMap[T any]() map[string]string {
	m := make(map[string]string)
	var zero T
	t := reflect.TypeOf(zero)
	if t == nil {
		return m
	}
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return m
	}
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		gormTag := f.Tag.Get("gorm")
		if gormTag == "-" {
			continue
		}
		colName := f.Name
		for _, part := range strings.Split(gormTag, ";") {
			if strings.HasPrefix(part, "column:") {
				colName = strings.TrimPrefix(part, "column:")
				break
			}
		}
		m[strings.ToLower(f.Name)] = colName
		m[strings.ToLower(colName)] = colName
		jsonTag := strings.Split(f.Tag.Get("json"), ",")[0]
		if jsonTag != "" && jsonTag != "-" {
			m[strings.ToLower(jsonTag)] = colName
		}
	}
	return m
}

func (r *GenericRepository[T]) resolveColumn(field string) string {
	key := strings.ToLower(strings.TrimSpace(field))
	if col, ok := r.colMap[key]; ok {
		return col
	}
	return strings.TrimSpace(field)
}

func (r *GenericRepository[T]) ApplyCommonFilters(query *gorm.DB, filter *dtos.Filter) *gorm.DB {
	if filter == nil {
		if r.defaultOrder != "" {
			query = query.Order(r.defaultOrder)
		}
		return query
	}

	// 1. Global SearchText (supports '/' AND-splitting and "'/'" literal slash escape matching .NET ServiceBase)
	if strings.TrimSpace(filter.SearchText) != "" && len(r.searchProps) > 0 {
		placeholder := "___LITERAL_SLASH___"
		processed := strings.ReplaceAll(filter.SearchText, "'/'", placeholder)
		rawParts := strings.Split(processed, "/")
		for _, rawPart := range rawParts {
			part := strings.TrimSpace(strings.ReplaceAll(rawPart, placeholder, "/"))
			if part == "" {
				continue
			}
			likePattern := "%" + part + "%"
			var orClauses []string
			var orArgs []any
			for _, prop := range r.searchProps {
				col := r.resolveColumn(prop)
				orClauses = append(orClauses, fmt.Sprintf("LOWER(CAST(\"%s\" AS VARCHAR(500))) LIKE LOWER(?)", col))
				orArgs = append(orArgs, likePattern)
			}
			if len(orClauses) > 0 {
				query = query.Where(strings.Join(orClauses, " OR "), orArgs...)
			}
		}
	}

	// 2. Advanced Filters
	for _, adv := range filter.AdvancedFilters {
		if strings.TrimSpace(adv.Field) == "" || strings.TrimSpace(adv.Value) == "" {
			continue
		}
		col := r.resolveColumn(adv.Field)
		op := strings.TrimSpace(adv.Operator)
		rawVal := adv.Value

		placeholder := "___LITERAL_SLASH___"
		processedVal := strings.ReplaceAll(rawVal, "'/'", placeholder)

		switch op {
		case "==", "=", "eq":
			parts := strings.Split(processedVal, "/")
			var trimmed []string
			for _, p := range parts {
				t := strings.TrimSpace(strings.ReplaceAll(p, placeholder, "/"))
				if t != "" {
					trimmed = append(trimmed, strings.ToLower(t))
				}
			}
			if len(trimmed) == 1 {
				query = query.Where(fmt.Sprintf("LOWER(CAST(\"%s\" AS VARCHAR(500))) = LOWER(?)", col), trimmed[0])
			} else if len(trimmed) > 1 {
				query = query.Where(fmt.Sprintf("LOWER(CAST(\"%s\" AS VARCHAR(500))) IN (?)", col), trimmed)
			}
		case "!=", "neq":
			parts := strings.Split(processedVal, "/")
			var trimmed []string
			for _, p := range parts {
				t := strings.TrimSpace(strings.ReplaceAll(p, placeholder, "/"))
				if t != "" {
					trimmed = append(trimmed, strings.ToLower(t))
				}
			}
			if len(trimmed) == 1 {
				query = query.Where(fmt.Sprintf("LOWER(CAST(\"%s\" AS VARCHAR(500))) != LOWER(?)", col), trimmed[0])
			} else if len(trimmed) > 1 {
				query = query.Where(fmt.Sprintf("LOWER(CAST(\"%s\" AS VARCHAR(500))) NOT IN (?)", col), trimmed)
			}
		case ">=", "gte":
			query = query.Where(fmt.Sprintf("\"%s\" >= ?", col), rawVal)
		case "<=", "lte":
			query = query.Where(fmt.Sprintf("\"%s\" <= ?", col), rawVal)
		case ">", "gt":
			query = query.Where(fmt.Sprintf("\"%s\" > ?", col), rawVal)
		case "<", "lt":
			query = query.Where(fmt.Sprintf("\"%s\" < ?", col), rawVal)
		case "startsWith":
			query = query.Where(fmt.Sprintf("LOWER(CAST(\"%s\" AS VARCHAR(500))) LIKE LOWER(?)", col), rawVal+"%")
		case "endsWith":
			query = query.Where(fmt.Sprintf("LOWER(CAST(\"%s\" AS VARCHAR(500))) LIKE LOWER(?)", col), "%"+rawVal)
		default:
			parts := strings.Split(processedVal, "/")
			var orClauses []string
			var orArgs []any
			for _, p := range parts {
				t := strings.TrimSpace(strings.ReplaceAll(p, placeholder, "/"))
				if t != "" {
					orClauses = append(orClauses, fmt.Sprintf("LOWER(CAST(\"%s\" AS VARCHAR(500))) LIKE LOWER(?)", col))
					orArgs = append(orArgs, "%"+t+"%")
				}
			}
			if len(orClauses) > 0 {
				query = query.Where(strings.Join(orClauses, " OR "), orArgs...)
			}
		}
	}

	// 3. OrderBy
	if strings.TrimSpace(filter.OrderBy) != "" {
		parts := strings.Fields(filter.OrderBy)
		if len(parts) > 0 {
			col := r.resolveColumn(parts[0])
			direction := "ASC"
			if len(parts) > 1 && strings.EqualFold(parts[1], "desc") {
				direction = "DESC"
			}
			query = query.Order(fmt.Sprintf("\"%s\" %s", col, direction))
		}
	} else if r.defaultOrder != "" {
		query = query.Order(r.defaultOrder)
	}

	return query
}

func (r *GenericRepository[T]) FindPaged(ctx context.Context, filter *dtos.Filter) ([]T, int64, error) {
	entities := make([]T, 0)
	var total int64

	db := getDB(ctx, r.db)
	query := db.Model(new(T))
	query = r.ApplyCommonFilters(query, filter)

	if filter == nil || !filter.SkipCount {
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
	entities := make([]T, 0)
	db := getDB(ctx, r.db)
	if r.defaultOrder != "" {
		db = db.Order(r.defaultOrder)
	}
	if err := db.Find(&entities).Error; err != nil {
		return nil, err
	}
	return entities, nil
}

func (r *GenericRepository[T]) applyPrimaryKeyWhere(q *gorm.DB, id any) *gorm.DB {
	if strID, ok := id.(string); ok {
		trimmed := strings.TrimSpace(strID)
		if intVal, err := strconv.ParseInt(trimmed, 10, 64); err == nil {
			return q.Where(fmt.Sprintf("\"%s\" = ? OR CAST(\"%s\" AS VARCHAR(500)) = ?", r.primaryKey, r.primaryKey), intVal, trimmed)
		}
		return q.Where(fmt.Sprintf("CAST(\"%s\" AS VARCHAR(500)) = ?", r.primaryKey), trimmed)
	}
	return q.Where(fmt.Sprintf("\"%s\" = ?", r.primaryKey), id)
}

func (r *GenericRepository[T]) FindByID(ctx context.Context, id any) (*T, error) {
	var entity T
	err := r.applyPrimaryKeyWhere(getDB(ctx, r.db), id).First(&entity).Error
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

func (r *GenericRepository[T]) Create(ctx context.Context, entity *T) error {
	return getDB(ctx, r.db).Create(entity).Error
}

func (r *GenericRepository[T]) Update(ctx context.Context, id any, entity *T) error {
	return r.applyPrimaryKeyWhere(getDB(ctx, r.db).Model(new(T)), id).
		Select("*").
		Omit(r.primaryKey).
		Updates(entity).Error
}

func (r *GenericRepository[T]) Delete(ctx context.Context, id any) error {
	return r.applyPrimaryKeyWhere(getDB(ctx, r.db), id).Delete(new(T)).Error
}
