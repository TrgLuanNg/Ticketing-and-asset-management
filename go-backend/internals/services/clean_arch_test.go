package services

import (
	"context"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"go-backend/configs"
	"go-backend/internals/core/domains"
	"go-backend/internals/core/dtos"
	"go-backend/pkg/cache"
	"go-backend/pkg/utils"
)

// TestCleanArchitectureImportBoundaries verifies that core, services, and handlers strictly obey Clean Architecture dependency rules.
func TestCleanArchitectureImportBoundaries(t *testing.T) {
	forbiddenInCore := []string{
		"gorm.io/gorm",
		"github.com/gin-gonic/gin",
		"github.com/jung-kurt/gofpdf",
		"github.com/xuri/excelize",
		"github.com/robfig/cron",
		"go-backend/internals/repository",
		"go-backend/internals/services",
		"go-backend/internals/api",
		"go-backend/internals/infrastructure",
	}

	forbiddenInServices := []string{
		"gorm.io/gorm",
		"github.com/gin-gonic/gin",
		"github.com/jung-kurt/gofpdf",
		"github.com/xuri/excelize",
		"github.com/robfig/cron",
		"go-backend/internals/repository",
		"go-backend/internals/api",
		"go-backend/internals/infrastructure",
	}

	forbiddenInHandlers := []string{
		"gorm.io/gorm",
		"go-backend/internals/repository",
		"go-backend/internals/services",
		"go-backend/internals/infrastructure",
	}

	checkDirImports := func(relDir string, forbidden []string) {
		fset := token.NewFileSet()
		err := filepath.Walk(relDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return nil
			}
			f, parseErr := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
			if parseErr != nil {
				t.Fatalf("failed to parse %s: %v", path, parseErr)
			}
			for _, imp := range f.Imports {
				impPath := strings.Trim(imp.Path.Value, "\"")
				for _, bad := range forbidden {
					if impPath == bad || strings.HasPrefix(impPath, bad+"/") {
						t.Errorf("Clean Architecture violation in %s: forbidden import %q", path, impPath)
					}
				}
			}
			return nil
		})
		if err != nil {
			t.Fatalf("failed walking %s: %v", relDir, err)
		}
	}

	checkDirImports("../core", forbiddenInCore)
	checkDirImports(".", forbiddenInServices)
	checkDirImports("../api/handlers", forbiddenInHandlers)
}

// mockUserRepo is an in-memory mock implementing ports.UserRepository for testing AuthService and UserInfoService.
type mockUserRepo struct {
	users       map[string]*domains.USERINFO
	permissions map[string]*domains.PERMISSION
	tokens      map[string]*domains.REFRESH_TOKEN
	history     []domains.HISTORY
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{
		users:       make(map[string]*domains.USERINFO),
		permissions: make(map[string]*domains.PERMISSION),
		tokens:      make(map[string]*domains.REFRESH_TOKEN),
	}
}

func (m *mockUserRepo) FindPaged(ctx context.Context, filter *dtos.Filter) ([]domains.USERINFO, int64, error) {
	var list []domains.USERINFO
	for _, u := range m.users {
		list = append(list, *u)
	}
	return list, int64(len(list)), nil
}

func (m *mockUserRepo) FindAll(ctx context.Context) ([]domains.USERINFO, error) {
	var list []domains.USERINFO
	for _, u := range m.users {
		list = append(list, *u)
	}
	return list, nil
}

func (m *mockUserRepo) FindByUserID(ctx context.Context, userID string) (*domains.USERINFO, error) {
	for k, u := range m.users {
		if strings.EqualFold(k, userID) {
			copyU := *u
			return &copyU, nil
		}
	}
	return nil, os.ErrNotExist
}

func (m *mockUserRepo) CreateUser(ctx context.Context, user *domains.USERINFO) error {
	copyU := *user
	m.users[user.UserID] = &copyU
	return nil
}

func (m *mockUserRepo) UpdateUser(ctx context.Context, user *domains.USERINFO) error {
	copyU := *user
	m.users[user.UserID] = &copyU
	return nil
}

func (m *mockUserRepo) UpdateUserFields(ctx context.Context, id string, fields map[string]any) error {
	if u, ok := m.users[id]; ok {
		if v, ok := fields["UserName"].(string); ok {
			u.UserName = v
		}
		if v, ok := fields["Role"].(string); ok {
			u.Role = v
		}
		if v, ok := fields["Area"].(string); ok {
			u.Area = v
		}
	}
	return nil
}

func (m *mockUserRepo) DeleteUser(ctx context.Context, userID string) error {
	delete(m.users, userID)
	return nil
}

func (m *mockUserRepo) FindPermission(ctx context.Context, userID string) (*domains.PERMISSION, error) {
	if p, ok := m.permissions[userID]; ok {
		copyP := *p
		return &copyP, nil
	}
	return nil, os.ErrNotExist
}

func (m *mockUserRepo) CreatePermission(ctx context.Context, perm *domains.PERMISSION) error {
	copyP := *perm
	m.permissions[perm.UserID] = &copyP
	return nil
}

func (m *mockUserRepo) UpsertPermission(ctx context.Context, perm *domains.PERMISSION) error {
	copyP := *perm
	m.permissions[perm.UserID] = &copyP
	return nil
}

func (m *mockUserRepo) CreateRefreshToken(ctx context.Context, token *domains.REFRESH_TOKEN) error {
	copyT := *token
	m.tokens[token.Token] = &copyT
	return nil
}

func (m *mockUserRepo) FindRefreshToken(ctx context.Context, tokenStr string) (*domains.REFRESH_TOKEN, error) {
	if t, ok := m.tokens[tokenStr]; ok {
		copyT := *t
		return &copyT, nil
	}
	return nil, os.ErrNotExist
}

func (m *mockUserRepo) UpdateRefreshToken(ctx context.Context, token *domains.REFRESH_TOKEN) error {
	copyT := *token
	m.tokens[token.Token] = &copyT
	return nil
}

func (m *mockUserRepo) CleanupExpiredTokens(ctx context.Context) {}

func (m *mockUserRepo) GetUserHistory(ctx context.Context, userID string) ([]domains.HISTORY, error) {
	return m.history, nil
}

func (m *mockUserRepo) AddHistory(ctx context.Context, userID, action string) {
	now := time.Now()
	m.history = append(m.history, domains.HISTORY{
		UserID:     userID,
		Action:     action,
		DateTimeAC: &now,
	})
}

type noopTxManager struct{}

func (noopTxManager) RunInTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

func TestAuthAndUserPasswordHashingWorkflow(t *testing.T) {
	ctx := context.Background()
	repo := newMockUserRepo()
	hasher := utils.NewPasswordHasher()
	memCache := cache.GetCache()
	cfg := &configs.Config{}
	cfg.JWT.SecretKey = "unit-test-secret-key"
	cfg.JWT.ExpirationMinutes = 60

	// Seed admin user with bcrypt hash
	adminHash, err := hasher.HashPassword("password123")
	if err != nil {
		t.Fatalf("failed to hash admin password: %v", err)
	}
	active := true
	repo.users["admin"] = &domains.USERINFO{
		UserID:   "admin",
		UserName: "System Admin",
		Pass:     adminHash,
		Role:     "Admin",
		Area:     "ALL",
		IsActive: &active,
	}

	authSvc := NewAuthService(repo, nil, hasher, memCache, cfg)
	userSvc := NewUserInfoService(repo, noopTxManager{}, hasher, memCache)

	// 1. Login with wrong password fails and increments FailedLoginAttempts
	if _, err := authSvc.Login(ctx, "admin", "wrong-pass"); err == nil {
		t.Fatal("expected login with wrong password to fail")
	}
	if repo.users["admin"].FailedLoginAttempts == nil || *repo.users["admin"].FailedLoginAttempts != 1 {
		t.Fatalf("expected FailedLoginAttempts == 1, got %v", repo.users["admin"].FailedLoginAttempts)
	}

	// 2. Login with valid password succeeds and resets FailedLoginAttempts
	loginRes, err := authSvc.Login(ctx, "admin", "password123")
	if err != nil {
		t.Fatalf("expected login with valid password to succeed, got: %v", err)
	}
	if loginRes.Token == "" || loginRes.RefreshToken == "" {
		t.Fatal("expected non-empty JWT and RefreshToken")
	}
	if *repo.users["admin"].FailedLoginAttempts != 0 {
		t.Fatalf("expected FailedLoginAttempts reset to 0, got %d", *repo.users["admin"].FailedLoginAttempts)
	}

	// 3. Insert new user hashes password with bcrypt
	_, err = userSvc.Insert(ctx, "admin", &dtos.UserInfoDto{
		UserID:   "emp01",
		UserName: "Employee 01",
		Pass:     "InitPass#2026",
		Area:     "South",
		Role:     "User",
	})
	if err != nil {
		t.Fatalf("Insert user failed: %v", err)
	}
	storedEmp := repo.users["emp01"]
	if storedEmp == nil {
		t.Fatal("expected emp01 to be stored")
	}
	if storedEmp.Pass == "InitPass#2026" || !utils.IsBcryptHash(storedEmp.Pass) {
		t.Fatalf("expected emp01 password to be stored as bcrypt hash, got %q", storedEmp.Pass)
	}

	// 4. ChangePassword verifies old password and stores new bcrypt hash
	if err := userSvc.ChangePassword(ctx, "emp01", &dtos.ChangePasswordDto{
		OldPassword: "wrong-old-password",
		NewPassword: "NewSecurePassword!99",
	}); err == nil {
		t.Fatal("expected ChangePassword with wrong old password to fail")
	}

	if err := userSvc.ChangePassword(ctx, "emp01", &dtos.ChangePasswordDto{
		OldPassword: "InitPass#2026",
		NewPassword: "NewSecurePassword!99",
	}); err != nil {
		t.Fatalf("expected ChangePassword with valid old password to succeed, got: %v", err)
	}

	if _, err := authSvc.Login(ctx, "emp01", "NewSecurePassword!99"); err != nil {
		t.Fatalf("expected login with new password to succeed, got: %v", err)
	}
}
