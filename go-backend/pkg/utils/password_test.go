package utils_test

import (
	"testing"

	"go-backend/pkg/utils"
	"golang.org/x/crypto/bcrypt"
)

func TestPasswordHasher_HashAndVerify(t *testing.T) {
	hasher := utils.NewPasswordHasher(bcrypt.DefaultCost)
	password := "P@ssw0rd123!"

	hash1, err := hasher.HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}
	if !utils.IsBcryptHash(hash1) {
		t.Fatalf("Expected valid 60-char bcrypt hash, got: %s (len=%d)", hash1, len(hash1))
	}

	// Ensure random salt produces distinct hashes for the same password
	hash2, err := hasher.HashPassword(password)
	if err != nil {
		t.Fatalf("Second HashPassword failed: %v", err)
	}
	if hash1 == hash2 {
		t.Errorf("Expected different hashes due to random salt, but got identical hashes")
	}

	// Verify cost parameter embedded in hash
	cost, err := bcrypt.Cost([]byte(hash1))
	if err != nil {
		t.Fatalf("bcrypt.Cost failed: %v", err)
	}
	if cost != bcrypt.DefaultCost {
		t.Errorf("Expected bcrypt cost %d, got %d", bcrypt.DefaultCost, cost)
	}

	// Verify correct password succeeds on both hashes
	if !hasher.VerifyPassword(hash1, password) {
		t.Errorf("VerifyPassword returned false for valid password against hash1")
	}
	if !hasher.VerifyPassword(hash2, password) {
		t.Errorf("VerifyPassword returned false for valid password against hash2")
	}

	// Verify wrong password fails
	if hasher.VerifyPassword(hash1, "wrong-password") {
		t.Errorf("VerifyPassword returned true for wrong password")
	}

	// Verify plaintext stored password is rejected by VerifyPassword
	if hasher.VerifyPassword(password, password) {
		t.Errorf("VerifyPassword must reject unhashed plaintext passwords")
	}
}

func TestPasswordHasher_EmptyPassword(t *testing.T) {
	hasher := utils.NewPasswordHasher()
	if _, err := hasher.HashPassword(""); err == nil {
		t.Errorf("Expected error when hashing empty password")
	}
	if _, err := hasher.HashPassword("   "); err == nil {
		t.Errorf("Expected error when hashing whitespace-only password")
	}
	if hasher.VerifyPassword("", "secret") {
		t.Errorf("Expected false when hashedPassword is empty")
	}
	if hasher.VerifyPassword("$2a$10$abcdefghijklmnopqrstuv", "") {
		t.Errorf("Expected false when candidate password is empty")
	}
}

func TestSeedPasswordHash(t *testing.T) {
	hasher := utils.NewPasswordHasher()
	seedHash := "$2a$10$2K8gASm.rDv.S383QJj4iOqy96DTQozGGkejEjtjJSACw98S.hGwO"
	if !hasher.VerifyPassword(seedHash, "password123") {
		t.Fatal("Failed to verify seeded admin bcrypt hash for 'password123'")
	}
}

