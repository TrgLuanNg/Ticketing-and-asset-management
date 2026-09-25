package utils_test

import (
	"testing"

	"go-backend/pkg/utils"
)

func TestCryptLib(t *testing.T) {
	crypt := utils.NewCryptLib()
	plaintext := "123456"
	key := "Inventory"
	iv := "DB"
	expectedCipher := "BEgCqBzhpEYnFc6JxVrBJQ=="

	encrypted, err := crypt.Encrypt(plaintext, key, iv)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	if encrypted != expectedCipher {
		t.Errorf("Encrypt mismatch. Expected: %s, got: %s", expectedCipher, encrypted)
	}

	decrypted, err := crypt.Decrypt(encrypted, key, iv)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}

	if decrypted != plaintext {
		t.Errorf("Decrypt mismatch. Expected: %s, got: %s", plaintext, decrypted)
	}
}
