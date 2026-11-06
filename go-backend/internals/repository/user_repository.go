package repository

import (
	"context"
	"strings"
	"time"

	"go-backend/internals/core/domains"
	"go-backend/internals/core/dtos"

	"gorm.io/gorm"
)

type UserRepository struct {
	*GenericRepository[domains.USERINFO]
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{
		GenericRepository: NewGenericRepository[domains.USERINFO](db, "UserID", []string{"UserID", "UserName", "Description", "Area", "Role"}),
		db:                db,
	}
}

func (r *UserRepository) FindByUserID(ctx context.Context, userID string) (*domains.USERINFO, error) {
	var user domains.USERINFO
	err := getDB(ctx, r.db).
		Where("LOWER(\"UserID\") = LOWER(?)", strings.TrimSpace(userID)).
		First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) CreateUser(ctx context.Context, user *domains.USERINFO) error {
	return getDB(ctx, r.db).Create(user).Error
}

func (r *UserRepository) UpdateUser(ctx context.Context, user *domains.USERINFO) error {
	return getDB(ctx, r.db).Save(user).Error
}

func (r *UserRepository) UpdateUserFields(ctx context.Context, id string, fields map[string]any) error {
	return getDB(ctx, r.db).Model(&domains.USERINFO{}).
		Where("\"UserID\" = ?", id).
		Updates(fields).Error
}

func (r *UserRepository) DeleteUser(ctx context.Context, userID string) error {
	return getDB(ctx, r.db).Where("\"UserID\" = ?", userID).Delete(&domains.USERINFO{}).Error
}

func (r *UserRepository) FindPermission(ctx context.Context, userID string) (*domains.PERMISSION, error) {
	var perm domains.PERMISSION
	err := getDB(ctx, r.db).
		Where("LOWER(\"UserID\") = LOWER(?)", strings.TrimSpace(userID)).
		First(&perm).Error
	if err != nil {
		return nil, err
	}
	return &perm, nil
}

func (r *UserRepository) CreatePermission(ctx context.Context, perm *domains.PERMISSION) error {
	return getDB(ctx, r.db).Create(perm).Error
}

func (r *UserRepository) UpsertPermission(ctx context.Context, perm *domains.PERMISSION) error {
	var existing domains.PERMISSION
	db := getDB(ctx, r.db)
	err := db.Where("LOWER(\"UserID\") = LOWER(?)", strings.TrimSpace(perm.UserID)).First(&existing).Error
	if err != nil {
		return db.Create(perm).Error
	}
	return db.Model(&domains.PERMISSION{}).
		Where("\"UserID\" = ?", existing.UserID).
		Select("*").
		Omit("UserID").
		Updates(perm).Error
}

func (r *UserRepository) CreateRefreshToken(ctx context.Context, token *domains.REFRESH_TOKEN) error {
	return getDB(ctx, r.db).Create(token).Error
}

func (r *UserRepository) FindRefreshToken(ctx context.Context, tokenStr string) (*domains.REFRESH_TOKEN, error) {
	var token domains.REFRESH_TOKEN
	err := getDB(ctx, r.db).Where("token = ?", tokenStr).First(&token).Error
	if err != nil {
		return nil, err
	}
	return &token, nil
}

func (r *UserRepository) UpdateRefreshToken(ctx context.Context, token *domains.REFRESH_TOKEN) error {
	return getDB(ctx, r.db).Save(token).Error
}

func (r *UserRepository) CleanupExpiredTokens(ctx context.Context) {
	cutoff := time.Now().Add(-72 * time.Hour)
	now := time.Now()
	getDB(ctx, r.db).
		Where("expires_at < ? OR (is_revoked = true AND created_at < ?)", now, cutoff).
		Delete(&domains.REFRESH_TOKEN{})
}

func (r *UserRepository) GetUserHistory(ctx context.Context, userID string) ([]domains.HISTORY, error) {
	var history []domains.HISTORY
	err := getDB(ctx, r.db).
		Where("user_id = ?", userID).
		Order("date_time_ac DESC").
		Find(&history).Error
	return history, err
}

func (r *UserRepository) AddHistory(ctx context.Context, userID, action string) {
	if strings.TrimSpace(userID) == "" {
		userID = "System"
	}
	now := time.Now()
	h := domains.HISTORY{
		UserID:     userID,
		Action:     action,
		DateTimeAC: &now,
	}
	_ = getDB(ctx, r.db).Create(&h).Error
}

func (r *UserRepository) FindPaged(ctx context.Context, filter *dtos.Filter) ([]domains.USERINFO, int64, error) {
	return r.GenericRepository.FindPaged(ctx, filter)
}
