package services

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"go-backend/internals/core/domains"
	"go-backend/internals/core/dtos"
	"go-backend/internals/repository"
	"go-backend/pkg/cache"
	"go-backend/pkg/utils"

	"gorm.io/gorm"
)

type UserInfoService struct {
	userRepo *repository.UserRepository
	db       *gorm.DB
	cache    *cache.MemoryCache
	crypt    *utils.CryptLib
}

func NewUserInfoService(userRepo *repository.UserRepository, db *gorm.DB) *UserInfoService {
	return &UserInfoService{
		userRepo: userRepo,
		db:       db,
		cache:    cache.GetCache(),
		crypt:    utils.NewCryptLib(),
	}
}

func (s *UserInfoService) GetByID(ctx context.Context, userID string) (*dtos.UserInfoDto, error) {
	user, err := s.userRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	if user.Role != "" && strings.EqualFold(user.Role, "Inactive") {
		return nil, errors.New("user is inactive")
	}

	dto := &dtos.UserInfoDto{
		UserID:      user.UserID,
		UserName:    user.UserName,
		Description: user.Description,
		Area:        user.Area,
		Role:        user.Role,
		IsActive:    user.IsActive,
	}

	if user.Role != "" {
		dto.Roles = []string{user.Role}
	}

	isAdmin := strings.EqualFold(user.Role, "admin")
	isViewer := strings.EqualFold(user.Role, "viewer")

	if !isViewer {
		perm, _ := s.userRepo.FindPermission(ctx, user.UserID)
		if perm == nil {
			if isAdmin {
				dto.Permissions = []string{
					"AssetsModify", "AssetsDel", "DepartmentModify", "DepartmentDel",
					"TypeModify", "TypeDel", "UserManager", "Disposal", "PermissionApproval", "AllowSeeAllFromAllowedAreas",
				}
			}
		} else {
			addPerm := func(name string, val *bool) {
				if (val != nil && *val) || (val == nil && isAdmin) {
					dto.Permissions = append(dto.Permissions, name)
				}
			}
			addPerm("AssetsModify", perm.AssetsModify)
			addPerm("AssetsDel", perm.AssetsDel)
			addPerm("DepartmentModify", perm.DepartmentModify)
			addPerm("DepartmentDel", perm.DepartmentDel)
			addPerm("TypeModify", perm.TypeModify)
			addPerm("TypeDel", perm.TypeDel)
			addPerm("UserManager", perm.UserManager)
			addPerm("Disposal", perm.Disposal)
			addPerm("PermissionApproval", perm.PermissionApproval)
			addPerm("AllowSeeAllFromAllowedAreas", perm.AllowSeeAllFromAllowedAreas)
		}
	}

	return dto, nil
}

func (s *UserInfoService) GetHistory(ctx context.Context, userID string) ([]dtos.HistoryDto, error) {
	historyList, err := s.userRepo.GetUserHistory(ctx, userID)
	if err != nil {
		return nil, err
	}

	dtosList := make([]dtos.HistoryDto, len(historyList))
	for i, h := range historyList {
		dtosList[i] = dtos.HistoryDto{
			ID:         h.ID,
			UserID:     h.UserID,
			Action:     h.Action,
			DateTimeAC: h.DateTimeAC,
		}
	}
	return dtosList, nil
}

func (s *UserInfoService) ChangePassword(ctx context.Context, requesterId string, dto *dtos.ChangePasswordDto) error {
	if dto == nil || strings.TrimSpace(dto.NewPassword) == "" {
		return errors.New("new password cannot be empty")
	}

	user, err := s.userRepo.FindByUserID(ctx, requesterId)
	if err != nil || user == nil {
		return errors.New("user not found")
	}

	if dto.OldPassword != "" {
		encryptedOld, _ := s.crypt.Encrypt(dto.OldPassword, "Inventory", "DB")
		if user.Pass != dto.OldPassword && user.Pass != encryptedOld {
			return errors.New("current password is incorrect")
		}
	}

	encryptedNew, err := s.crypt.Encrypt(dto.NewPassword, "Inventory", "DB")
	if err != nil {
		return fmt.Errorf("failed to encrypt password: %w", err)
	}

	user.Pass = encryptedNew
	err = s.userRepo.UpdateUser(ctx, user)
	if err == nil {
		s.userRepo.AddHistory(ctx, requesterId, "Changed password")
	}
	return err
}

func (s *UserInfoService) Insert(ctx context.Context, requesterId string, dto *dtos.UserInfoDto) (*dtos.UserInfoDto, error) {
	if dto == nil || strings.TrimSpace(dto.UserID) == "" {
		return nil, errors.New("invalid user ID")
	}

	requesterUser, _ := s.userRepo.FindByUserID(ctx, requesterId)
	isAdmin := requesterUser != nil && strings.EqualFold(requesterUser.Role, "admin")
	requesterPerm, _ := s.userRepo.FindPermission(ctx, requesterId)
	hasUserManager := (requesterPerm == nil && isAdmin) || (requesterPerm != nil && requesterPerm.UserManager != nil && *requesterPerm.UserManager) || (requesterPerm != nil && requesterPerm.UserManager == nil && isAdmin)

	if !hasUserManager {
		return nil, errors.New("you do not have permission to manage users")
	}

	existing, _ := s.userRepo.FindByUserID(ctx, dto.UserID)
	if existing != nil {
		return nil, errors.New("user ID already exists")
	}

	role := dto.Role
	if role == "" && len(dto.Roles) > 0 {
		role = dto.Roles[0]
	}

	user := domains.USERINFO{
		UserID:      dto.UserID,
		UserName:    dto.UserName,
		Description: dto.Description,
		Area:        dto.Area,
		Role:        role,
		IsActive:    dto.IsActive,
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&user).Error; err != nil {
			return err
		}

		f := false
		defaultPerm := domains.PERMISSION{
			UserID:                     user.UserID,
			AssetsModify:               &f,
			AssetsDel:                  &f,
			DepartmentModify:           &f,
			DepartmentDel:              &f,
			TypeModify:                 &f,
			TypeDel:                    &f,
			UserManager:                &f,
			Disposal:                   &f,
			PermissionApproval:         &f,
			AllowSeeAllFromAllowedAreas: &f,
		}
		return tx.Create(&defaultPerm).Error
	})

	if err != nil {
		return nil, err
	}

	s.userRepo.AddHistory(ctx, requesterId, fmt.Sprintf("Created user %s", user.UserID))
	return dto, nil
}

func (s *UserInfoService) Update(ctx context.Context, requesterId string, id string, dto *dtos.UserInfoDto) error {
	requesterUser, _ := s.userRepo.FindByUserID(ctx, requesterId)
	isAdmin := requesterUser != nil && strings.EqualFold(requesterUser.Role, "admin")
	requesterPerm, _ := s.userRepo.FindPermission(ctx, requesterId)

	hasPermissionApproval := (requesterPerm == nil && isAdmin) || (requesterPerm != nil && requesterPerm.PermissionApproval != nil && *requesterPerm.PermissionApproval)
	hasUserManager := (requesterPerm == nil && isAdmin) || (requesterPerm != nil && requesterPerm.UserManager != nil && *requesterPerm.UserManager)

	if !hasPermissionApproval && !hasUserManager {
		return errors.New("you do not have permission to modify user settings")
	}

	role := dto.Role
	if role == "" && len(dto.Roles) > 0 {
		role = dto.Roles[0]
	}

	err := s.db.WithContext(ctx).Model(&domains.USERINFO{}).
		Where("\"UserID\" = ?", id).
		Updates(map[string]interface{}{
			"UserName":    dto.UserName,
			"Description": dto.Description,
			"Area":        dto.Area,
			"Role":        role,
			"IsActive":    dto.IsActive,
		}).Error

	// Evict memory cache for user allowed areas
	s.cache.Remove(fmt.Sprintf("user_areas_%s", id))
	s.userRepo.AddHistory(ctx, requesterId, fmt.Sprintf("Updated user %s", id))
	return err
}
