package utils

import (
	"errors"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// PasswordHasher defines salted, cost-parameterized one-way password hashing.
type PasswordHasher struct {
	cost int
}

// NewPasswordHasher creates a new bcrypt PasswordHasher with the given cost (defaults to bcrypt.DefaultCost if <= 0).
func NewPasswordHasher(cost ...int) *PasswordHasher {
	c := bcrypt.DefaultCost
	if len(cost) > 0 && cost[0] >= bcrypt.MinCost && cost[0] <= bcrypt.MaxCost {
		c = cost[0]
	}
	return &PasswordHasher{cost: c}
}

// HashPassword hashes a plaintext password using bcrypt with a random salt and configured cost.
func (h *PasswordHasher) HashPassword(password string) (string, error) {
	if strings.TrimSpace(password) == "" {
		return "", errors.New("password cannot be empty")
	}
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// VerifyPassword compares a bcrypt hashed password with a plaintext candidate.
func (h *PasswordHasher) VerifyPassword(hashedPassword, password string) bool {
	if hashedPassword == "" || password == "" {
		return false
	}
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

// IsBcryptHash checks whether a stored string is already a valid bcrypt hash prefix.
func IsBcryptHash(s string) bool {
	return len(s) == 60 && (strings.HasPrefix(s, "$2a$") || strings.HasPrefix(s, "$2b$") || strings.HasPrefix(s, "$2y$"))
}
