package utils_test

import (
	"testing"

	"go-backend/pkg/utils"
)

func TestJWTGenerationAndValidation(t *testing.T) {
	userId := "user01"
	userName := "john.doe"
	secret := "test-secret-key-12345"
	expMinutes := 60

	token, err := utils.GenerateJWT(userId, userName, secret, expMinutes)
	if err != nil {
		t.Fatalf("Failed to generate JWT: %v", err)
	}

	claims, err := utils.ValidateJWT(token, secret)
	if err != nil {
		t.Fatalf("Failed to validate JWT: %v", err)
	}

	if claims.UserID != userId {
		t.Errorf("UserID mismatch. Expected %s, got %s", userId, claims.UserID)
	}

	if claims.UserName != userName {
		t.Errorf("UserName mismatch. Expected %s, got %s", userName, claims.UserName)
	}

	// Validation with wrong secret should fail
	_, err = utils.ValidateJWT(token, "wrong-secret")
	if err == nil {
		t.Errorf("Expected validation failure with wrong secret, but succeeded")
	}
}
