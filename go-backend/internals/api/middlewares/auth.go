package middlewares

import (
	"net/http"
	"strings"
	"time"

	"go-backend/internals/core/dtos"
	"go-backend/pkg/utils"

	"github.com/gin-gonic/gin"
)

// TokenValidationMiddleware validates incoming Bearer tokens and checks X-User-Id header
func TokenValidationMiddleware(secretKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := strings.ToLower(c.Request.URL.Path)

		// Whitelist public endpoints
		if path == "/" ||
			path == "/api/auths" ||
			path == "/api/auths/login" ||
			path == "/api/users/login" ||
			path == "/api/auths/refresh" ||
			strings.HasPrefix(path, "/api/auths") ||
			strings.HasPrefix(path, "/openapi") ||
			strings.Contains(path, "swagger") ||
			strings.HasPrefix(path, "/ws") {
			c.Next()
			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, dtos.ErrorResponse("Authorization token is missing or invalid."))
			return
		}

		tokenStr := strings.TrimSpace(authHeader[7:])
		claims, err := utils.DecodeUnverifiedClaims(tokenStr)
		if err != nil || len(claims) == 0 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, dtos.ErrorResponse("Failed to decode token."))
			return
		}

		// Verify expiration
		if expVal, ok := claims["exp"]; ok {
			var expSeconds int64
			switch v := expVal.(type) {
			case float64:
				expSeconds = int64(v)
			case int64:
				expSeconds = v
			}
			if expSeconds > 0 && time.Now().Unix() > expSeconds {
				c.AbortWithStatusJSON(http.StatusUnauthorized, dtos.ErrorResponse("Access token has expired."))
				return
			}
		}

		// Extract User ID from claims
		tokenUserId := ""
		if v, ok := claims["http://schemas.xmlsoap.org/ws/2005/05/identity/claims/nameidentifier"].(string); ok {
			tokenUserId = v
		} else if v, ok := claims["nameid"].(string); ok {
			tokenUserId = v
		} else if v, ok := claims["sub"].(string); ok {
			tokenUserId = v
		}

		headerUserId := c.GetHeader("X-User-Id")
		if headerUserId == "" || tokenUserId == "" || !strings.EqualFold(headerUserId, tokenUserId) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, dtos.ErrorResponse("User ID mismatch or missing header."))
			return
		}

		c.Set("userSession", tokenUserId)
		if userName, ok := claims["unique_name"].(string); ok {
			c.Set("userName", userName)
		}

		c.Next()
	}
}

// GetUserSession retrieves the authenticated user ID from Gin context
func GetUserSession(c *gin.Context) string {
	if val, exists := c.Get("userSession"); exists {
		if s, ok := val.(string); ok {
			return s
		}
	}
	return c.GetHeader("X-User-Id")
}
