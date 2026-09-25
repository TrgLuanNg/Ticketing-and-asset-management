package services

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"go-backend/configs"
	"go-backend/internals/core/domains"
	"go-backend/internals/core/dtos"
	"go-backend/internals/repository"
	"go-backend/pkg/cache"
	"go-backend/pkg/logger"
	"go-backend/pkg/utils"
)

type AuthService struct {
	userRepo *repository.UserRepository
	cache    *cache.MemoryCache
	crypt    *utils.CryptLib
	cfg      *configs.Config
}

func NewAuthService(userRepo *repository.UserRepository, cfg *configs.Config) *AuthService {
	return &AuthService{
		userRepo: userRepo,
		cache:    cache.GetCache(),
		crypt:    utils.NewCryptLib(),
		cfg:      cfg,
	}
}

func (s *AuthService) Login(ctx context.Context, username, password string) (*dtos.LoginResponseDto, error) {
	if strings.TrimSpace(username) == "" || strings.TrimSpace(password) == "" {
		return nil, errors.New("username and password are required")
	}

	user, err := s.userRepo.FindByUserID(ctx, username)
	if err != nil {
		return nil, errors.New("invalid username or password")
	}

	if user.Role != "" && strings.EqualFold(user.Role, "Inactive") {
		return nil, errors.New("user account is inactive")
	}

	if user.IsActive != nil && !*user.IsActive {
		return nil, errors.New("user account is disabled")
	}

	if user.LockoutEnd != nil && user.LockoutEnd.After(time.Now()) {
		return nil, errors.New("user account is temporarily locked due to failed login attempts. Please try again later")
	}

	// Verify password: check plaintext or AES-256-CBC encrypted
	encryptedInput, _ := s.crypt.Encrypt(password, "Inventory", "DB")
	isValid := (user.Pass == password || (encryptedInput != "" && user.Pass == encryptedInput))

	if !isValid {
		failedCount := 0
		if user.FailedLoginAttempts != nil {
			failedCount = *user.FailedLoginAttempts
		}
		failedCount++
		user.FailedLoginAttempts = &failedCount

		if failedCount >= 5 {
			lockout := time.Now().Add(15 * time.Minute)
			user.LockoutEnd = &lockout
		}
		_ = s.userRepo.UpdateUser(ctx, user)
		return nil, errors.New("invalid username or password")
	}

	// Reset failed attempts upon successful login
	zero := 0
	user.FailedLoginAttempts = &zero
	user.LockoutEnd = nil
	_ = s.userRepo.UpdateUser(ctx, user)

	s.userRepo.AddHistory(ctx, user.UserID, "User logged in via local credentials")
	return s.buildSessionResponse(ctx, user)
}

func (s *AuthService) GetResource(ctx context.Context, authCode string) (*dtos.LoginResponseDto, error) {
	claims, _, err := s.userRepo.GetSSOUserInfo(authCode, &s.cfg.SSO)
	if err != nil {
		return nil, fmt.Errorf("SSO verification failed: %w", err)
	}

	var userId, userName string
	if v, ok := claims["preferred_username"].(string); ok && v != "" {
		userId = v
		userName = v
	} else if v, ok := claims["unique_name"].(string); ok && v != "" {
		userId = v
		userName = v
	} else if v, ok := claims["http://schemas.xmlsoap.org/ws/2005/05/identity/claims/nameidentifier"].(string); ok && v != "" {
		userId = v
	} else if v, ok := claims["nameid"].(string); ok && v != "" {
		userId = v
	} else if v, ok := claims["sub"].(string); ok && v != "" {
		userId = v
	}

	if v, ok := claims["name"].(string); ok && v != "" {
		userName = v
	} else if v, ok := claims["http://schemas.xmlsoap.org/ws/2005/05/identity/claims/name"].(string); ok && v != "" {
		userName = v
	}
	if userName == "" {
		userName = userId
	}

	if userId == "" {
		return nil, errors.New("user ID claim missing in SSO token")
	}

	logger.Infof("SSO token claims resolved: UserID=%s, UserName=%s", userId, userName)

	user, err := s.userRepo.FindByUserID(ctx, userId)
	if err != nil {
		role := "User"
		if strings.EqualFold(userId, "admin") {
			role = "Admin"
		}
		newUser := &domains.USERINFO{
			UserID:   userId,
			UserName: userName,
			Role:     role,
		}
		_ = s.userRepo.UpdateUser(ctx, newUser)
		user = newUser
	}

	logger.Infof("SSO user authenticated: UserID=%s, Role=%s", user.UserID, user.Role)

	if user.Role != "" && strings.EqualFold(user.Role, "Inactive") {
		return nil, errors.New("user account is inactive")
	}

	s.userRepo.AddHistory(ctx, user.UserID, "User logged in via SSO")
	return s.buildSessionResponse(ctx, user)
}

func (s *AuthService) RefreshSession(ctx context.Context, refreshTokenStr string) (*dtos.LoginResponseDto, error) {
	refreshToken, err := s.userRepo.FindRefreshToken(ctx, refreshTokenStr)
	if err != nil {
		return nil, errors.New("refresh token does not exist")
	}

	if refreshToken.IsRevoked != nil && *refreshToken.IsRevoked {
		return nil, errors.New("refresh token has been revoked")
	}

	if refreshToken.ExpiresAt != nil && refreshToken.ExpiresAt.Before(time.Now()) {
		return nil, errors.New("refresh token has expired")
	}

	user, err := s.userRepo.FindByUserID(ctx, refreshToken.UserID)
	if err != nil || (user.Role != "" && strings.EqualFold(user.Role, "Inactive")) {
		return nil, errors.New("user associated with this refresh token not found or is inactive")
	}

	// Rotate refresh token
	newRefreshToken := s.generateRandomToken()
	refreshToken.ReplacedByToken = refreshToken.Token
	refreshToken.Token = newRefreshToken
	_ = s.userRepo.UpdateRefreshToken(ctx, refreshToken)

	s.userRepo.CleanupExpiredTokens(ctx)
	return s.buildSessionResponse(ctx, user, newRefreshToken)
}

func (s *AuthService) buildSessionResponse(ctx context.Context, user *domains.USERINFO, existingRefreshToken ...string) (*dtos.LoginResponseDto, error) {
	var roles []string
	if user.Role != "" {
		roles = append(roles, user.Role)
	}

	var permissions []string
	isAdmin := strings.EqualFold(user.Role, "admin")
	isViewer := strings.EqualFold(user.Role, "viewer")

	if !isViewer {
		perm, _ := s.userRepo.FindPermission(ctx, user.UserID)
		if perm == nil {
			if isAdmin {
				permissions = []string{
					"AssetsModify", "AssetsDel", "DepartmentModify", "DepartmentDel",
					"TypeModify", "TypeDel", "UserManager", "Disposal", "PermissionApproval", "AllowSeeAllFromAllowedAreas",
				}
			}
		} else {
			addPerm := func(name string, val *bool) {
				if (val != nil && *val) || (val == nil && isAdmin) {
					permissions = append(permissions, name)
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

	// JWT Token
	jwtToken, err := utils.GenerateJWT(user.UserID, user.UserName, s.cfg.JWT.SecretKey, s.cfg.JWT.ExpirationMinutes)
	if err != nil {
		return nil, fmt.Errorf("failed to generate JWT: %w", err)
	}

	// Refresh Token
	refreshToken := ""
	if len(existingRefreshToken) > 0 && existingRefreshToken[0] != "" {
		refreshToken = existingRefreshToken[0]
	} else {
		refreshToken = s.generateRandomToken()
		now := time.Now()
		exp := now.Add(72 * time.Hour)
		revoked := false
		refEntity := &domains.REFRESH_TOKEN{
			Token:     refreshToken,
			UserID:    user.UserID,
			CreatedAt: &now,
			ExpiresAt: &exp,
			IsRevoked: &revoked,
		}
		_ = s.userRepo.CreateRefreshToken(ctx, refEntity)
	}

	s.userRepo.CleanupExpiredTokens(ctx)

	return &dtos.LoginResponseDto{
		UserName:     user.UserName,
		UserID:       user.UserID,
		Token:        jwtToken,
		RefreshToken: refreshToken,
		Roles:        roles,
		Area:         user.Area,
		Permissions:  permissions,
	}, nil
}

func (s *AuthService) generateRandomToken() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}

func (s *AuthService) GetUserAllowedAreas(ctx context.Context, userID string) ([]string, bool) {
	cacheKey := fmt.Sprintf("user_areas_%s", userID)
	if val, found := s.cache.Get(cacheKey); found {
		if res, ok := val.([]string); ok {
			for _, a := range res {
				if strings.EqualFold(a, "ALL") {
					return res, true
				}
			}
			return res, false
		}
	}

	user, err := s.userRepo.FindByUserID(ctx, userID)
	if err != nil || user == nil {
		return []string{}, false
	}

	if strings.EqualFold(user.Role, "admin") {
		s.cache.Set(cacheKey, []string{"ALL"}, 30*time.Minute)
		return []string{"ALL"}, true
	}

	var areas []string
	if user.Area != "" {
		for _, a := range strings.Split(user.Area, ",") {
			trimmed := strings.TrimSpace(a)
			if trimmed != "" {
				areas = append(areas, trimmed)
			}
		}
	}

	perm, _ := s.userRepo.FindPermission(ctx, userID)
	seeAll := (perm != nil && perm.AllowSeeAllFromAllowedAreas != nil && *perm.AllowSeeAllFromAllowedAreas)

	s.cache.Set(cacheKey, areas, 30*time.Minute)
	return areas, seeAll
}
