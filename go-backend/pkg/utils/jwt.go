package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type TokenClaims struct {
	UserID   string   `json:"nameid"`
	UserName string   `json:"unique_name"`
	Subject  string   `json:"sub"`
	Roles    []string `json:"roles,omitempty"`
	jwt.RegisteredClaims
}

// GenerateJWT creates a signed JWT token containing standard claims expected by ITAM
func GenerateJWT(userId, userName, secretKey string, expMinutes int) (string, error) {
	if secretKey == "" {
		secretKey = "YusenLogisticsVietnamITAMSecretKey2026"
	}
	if expMinutes <= 0 {
		expMinutes = 120
	}

	claims := TokenClaims{
		UserID:   userId,
		UserName: userName,
		Subject:  userId,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expMinutes) * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   userId,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secretKey))
}

// ValidateJWT verifies a JWT token and returns its claims
func ValidateJWT(tokenString, secretKey string) (*TokenClaims, error) {
	if secretKey == "" {
		secretKey = "YusenLogisticsVietnamITAMSecretKey2026"
	}

	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*TokenClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token claims")
}

// DecodeUnverifiedClaims decodes token claims without verifying signature (used for external SSO tokens)
func DecodeUnverifiedClaims(tokenString string) (map[string]interface{}, error) {
	parser := jwt.NewParser()
	claims := jwt.MapClaims{}
	_, _, err := parser.ParseUnverified(tokenString, claims)
	if err != nil {
		return nil, err
	}
	return claims, nil
}
