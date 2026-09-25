package handlers

import (
	"net/http"

	"go-backend/internals/core/dtos"
	"go-backend/internals/services"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService *services.AuthService
}

func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Login handles local POST /api/auths/login
func (h *AuthHandler) Login(c *gin.Context) {
	var dto dtos.LoginDto
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse("Username and password are required."))
		return
	}

	res, err := h.authService.Login(c.Request.Context(), dto.Username, dto.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, dtos.SuccessResponse(res))
}

// GetByAuthCode handles SSO GET /api/auths?authCode=...
func (h *AuthHandler) GetByAuthCode(c *gin.Context) {
	authCode := c.Query("authCode")
	if authCode == "" {
		authCode = c.Query("code")
	}

	if authCode == "" {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse("Invalid Code"))
		return
	}

	res, err := h.authService.GetResource(c.Request.Context(), authCode)
	if err != nil {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, dtos.SuccessResponse(res))
}

// Refresh handles POST /api/auths/refresh
func (h *AuthHandler) Refresh(c *gin.Context) {
	var dto dtos.RefreshRequestDto
	if err := c.ShouldBindJSON(&dto); err != nil || dto.RefreshToken == "" {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse("Invalid refresh token"))
		return
	}

	res, err := h.authService.RefreshSession(c.Request.Context(), dto.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dtos.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, dtos.SuccessResponse(res))
}
