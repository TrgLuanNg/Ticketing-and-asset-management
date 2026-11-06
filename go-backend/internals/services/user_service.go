package services

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"go-backend/internals/core/domains"
	"go-backend/internals/core/dtos"
	"go-backend/internals/core/ports"
)

type UserInfoService struct {
	userRepo  ports.UserRepository
	txManager ports.TransactionManager
	hasher    ports.PasswordHasherPort
	cache     ports.CacheProvider
}

func NewUserInfoService(
	userRepo ports.UserRepository,
	txManager ports.TransactionManager,
	hasher ports.PasswordHasherPort,
	cache ports.CacheProvider,
) *UserInfoService {
	return &UserInfoService{
		userRepo:  userRepo,
		txManager: txManager,
		hasher:    hasher,
		cache:     cache,
	}
}

func (s *UserInfoService) Search(ctx context.Context, filter *dtos.Filter) ([]dtos.UserInfoDto, int64, error) {
	users, total, err := s.userRepo.FindPaged(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	res := make([]dtos.UserInfoDto, len(users))
	for i := range users {
		res[i] = *s.mapUserToDto(ctx, &users[i], false)
	}
	return res, total, nil
}

func (s *UserInfoService) GetAll(ctx context.Context) ([]dtos.UserInfoDto, error) {
	users, err := s.userRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	res := make([]dtos.UserInfoDto, len(users))
	for i := range users {
		res[i] = *s.mapUserToDto(ctx, &users[i], false)
	}
	return res, nil
}

func (s *UserInfoService) GetByID(ctx context.Context, userID string) (*dtos.UserInfoDto, error) {
	user, err := s.userRepo.FindByUserID(ctx, userID)
	if err != nil || user == nil {
		return nil, errors.New("user not found")
	}

	if user.Role != "" && strings.EqualFold(user.Role, "Inactive") {
		return nil, errors.New("user is inactive")
	}

	return s.mapUserToDto(ctx, user, true), nil
}

func (s *UserInfoService) mapUserToDto(ctx context.Context, user *domains.USERINFO, loadPermissions bool) *dtos.UserInfoDto {
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

	if loadPermissions {
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
	}

	return dto
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
		if !s.hasher.VerifyPassword(user.Pass, dto.OldPassword) {
			return errors.New("current password is incorrect")
		}
	}

	hashedPass, err := s.hasher.HashPassword(dto.NewPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	user.Pass = hashedPass
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
	if role == "" {
		role = "User"
	}

	rawPass := strings.TrimSpace(dto.Pass)
	if rawPass == "" {
		rawPass = "whatever"
	}
	hashedPass, err := s.hasher.HashPassword(rawPass)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	isActive := true
	if dto.IsActive != nil {
		isActive = *dto.IsActive
	}
	zero := 0

	user := domains.USERINFO{
		UserID:              dto.UserID,
		UserName:            dto.UserName,
		Pass:                hashedPass,
		Description:         dto.Description,
		Area:                dto.Area,
		Role:                role,
		IsActive:            &isActive,
		FailedLoginAttempts: &zero,
	}

	err = s.txManager.RunInTransaction(ctx, func(txCtx context.Context) error {
		if err := s.userRepo.CreateUser(txCtx, &user); err != nil {
			return err
		}

		f := false
		defaultPerm := domains.PERMISSION{
			UserID:                      user.UserID,
			AssetsModify:                &f,
			AssetsDel:                   &f,
			DepartmentModify:            &f,
			DepartmentDel:               &f,
			TypeModify:                  &f,
			TypeDel:                     &f,
			UserManager:                 &f,
			Disposal:                    &f,
			PermissionApproval:          &f,
			AllowSeeAllFromAllowedAreas: &f,
		}
		return s.userRepo.CreatePermission(txCtx, &defaultPerm)
	})

	if err != nil {
		return nil, err
	}

	dto.Pass = ""
	dto.Role = role
	dto.IsActive = &isActive
	s.userRepo.AddHistory(ctx, requesterId, fmt.Sprintf("Created user %s", user.UserID))
	return dto, nil
}

func (s *UserInfoService) Update(ctx context.Context, requesterId string, id string, dto *dtos.UserInfoDto) error {
	requesterUser, _ := s.userRepo.FindByUserID(ctx, requesterId)
	isAdmin := requesterUser != nil && strings.EqualFold(requesterUser.Role, "admin")
	requesterPerm, _ := s.userRepo.FindPermission(ctx, requesterId)

	hasPermissionApproval := (requesterPerm == nil && isAdmin) || (requesterPerm != nil && requesterPerm.PermissionApproval != nil && *requesterPerm.PermissionApproval) || (requesterPerm != nil && requesterPerm.PermissionApproval == nil && isAdmin)
	hasUserManager := (requesterPerm == nil && isAdmin) || (requesterPerm != nil && requesterPerm.UserManager != nil && *requesterPerm.UserManager) || (requesterPerm != nil && requesterPerm.UserManager == nil && isAdmin)

	if !hasPermissionApproval && !hasUserManager {
		return errors.New("you do not have permission to modify user settings")
	}

	role := dto.Role
	if role == "" && len(dto.Roles) > 0 {
		role = dto.Roles[0]
	}

	fields := map[string]any{
		"UserName":    dto.UserName,
		"Description": dto.Description,
		"Area":        dto.Area,
		"Role":        role,
	}
	if dto.IsActive != nil {
		fields["IsActive"] = *dto.IsActive
	}

	err := s.userRepo.UpdateUserFields(ctx, id, fields)
	if s.cache != nil {
		s.cache.Remove(fmt.Sprintf("user_areas_%s", id))
	}
	if err == nil {
		s.userRepo.AddHistory(ctx, requesterId, fmt.Sprintf("Updated user %s", id))
	}
	return err
}

func (s *UserInfoService) Delete(ctx context.Context, requesterId string, id string) error {
	requesterUser, _ := s.userRepo.FindByUserID(ctx, requesterId)
	isAdmin := requesterUser != nil && strings.EqualFold(requesterUser.Role, "admin")
	requesterPerm, _ := s.userRepo.FindPermission(ctx, requesterId)
	hasUserManager := (requesterPerm == nil && isAdmin) || (requesterPerm != nil && requesterPerm.UserManager != nil && *requesterPerm.UserManager) || (requesterPerm != nil && requesterPerm.UserManager == nil && isAdmin)

	if !hasUserManager {
		return errors.New("you do not have permission to manage users")
	}

	err := s.userRepo.DeleteUser(ctx, id)
	if s.cache != nil {
		s.cache.Remove(fmt.Sprintf("user_areas_%s", id))
	}
	if err == nil {
		s.userRepo.AddHistory(ctx, requesterId, fmt.Sprintf("Deleted user %s", id))
	}
	return err
}

func (s *UserInfoService) Logout(ctx context.Context, userSession string) {
	if strings.TrimSpace(userSession) != "" {
		s.userRepo.AddHistory(ctx, userSession, "Logged out of the system")
	}
}
