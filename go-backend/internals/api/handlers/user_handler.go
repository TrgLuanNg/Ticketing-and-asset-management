package handlers

import (
	"net/http"

	"go-backend/internals/api/middlewares"
	"go-backend/internals/core/dtos"
	"go-backend/internals/core/ports"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userUseCase ports.UserInfoUseCase
	authUseCase ports.AuthUseCase
}

func NewUserHandler(userUseCase ports.UserInfoUseCase, authUseCase ports.AuthUseCase) *UserHandler {
	return &UserHandler{
		userUseCase: userUseCase,
		authUseCase: authUseCase,
	}
}

// Login handles POST /api/users/login
func (h *UserHandler) Login(c *gin.Context) {
	var dto dtos.LoginDto
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusOK, dtos.ErrorResponse("Username and password are required"))
		return
	}

	res, err := h.authUseCase.Login(c.Request.Context(), dto.Username, dto.Password)
	if err != nil {
		c.JSON(http.StatusOK, dtos.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, dtos.SuccessResponse(res))
}

// ChangePassword handles POST /api/users/change-password
func (h *UserHandler) ChangePassword(c *gin.Context) {
	requesterId := middlewares.GetUserSession(c)
	if requesterId == "" {
		c.JSON(http.StatusOK, dtos.ErrorResponse("Unauthorized: missing user session"))
		return
	}

	var dto dtos.ChangePasswordDto
	if err := c.ShouldBindJSON(&dto); err != nil || dto.NewPassword == "" {
		c.JSON(http.StatusOK, dtos.ErrorResponse("New password cannot be empty"))
		return
	}

	if err := h.userUseCase.ChangePassword(c.Request.Context(), requesterId, &dto); err != nil {
		c.JSON(http.StatusOK, dtos.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, dtos.SuccessResponse(nil, "Password changed successfully"))
}

// GetHistory handles GET /api/users/:id/history
func (h *UserHandler) GetHistory(c *gin.Context) {
	userId := c.Param("id")
	if userId == "" {
		userId = c.Param("userId")
	}
	history, err := h.userUseCase.GetHistory(c.Request.Context(), userId)
	if err != nil {
		c.JSON(http.StatusOK, dtos.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, dtos.SuccessResponse(history))
}

// GetById handles GET /api/users/:id
func (h *UserHandler) GetById(c *gin.Context) {
	id := c.Param("id")
	user, err := h.userUseCase.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusOK, dtos.ErrorResponse("User not found or Inactive"))
		return
	}

	c.JSON(http.StatusOK, dtos.SuccessResponse(user))
}

// Insert handles POST /api/users
func (h *UserHandler) Insert(c *gin.Context) {
	requesterId := middlewares.GetUserSession(c)
	var dto dtos.UserInfoDto
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusOK, dtos.ErrorResponse("Invalid user payload"))
		return
	}

	created, err := h.userUseCase.Insert(c.Request.Context(), requesterId, &dto)
	if err != nil {
		c.JSON(http.StatusOK, dtos.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, dtos.SuccessResponse(created, "User created successfully with default permissions"))
}

// Update handles PUT /api/users/:id
func (h *UserHandler) Update(c *gin.Context) {
	id := c.Param("id")
	requesterId := middlewares.GetUserSession(c)

	var dto dtos.UserInfoDto
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusOK, dtos.ErrorResponse("Invalid payload"))
		return
	}

	if err := h.userUseCase.Update(c.Request.Context(), requesterId, id, &dto); err != nil {
		c.JSON(http.StatusOK, dtos.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, dtos.SuccessResponse(dto, "Updated successfully"))
}

// Delete handles DELETE /api/users/:id
func (h *UserHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	requesterId := middlewares.GetUserSession(c)

	if err := h.userUseCase.Delete(c.Request.Context(), requesterId, id); err != nil {
		c.JSON(http.StatusOK, dtos.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, dtos.SuccessResponse(nil, "Deleted successfully"))
}

// Search handles POST /api/users/search
func (h *UserHandler) Search(c *gin.Context) {
	var filter dtos.Filter
	_ = c.ShouldBindJSON(&filter)

	entities, total, err := h.userUseCase.Search(c.Request.Context(), &filter)
	if err != nil {
		c.JSON(http.StatusOK, dtos.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, dtos.PagedResponse(entities, total))
}

// GetAll handles GET /api/users/get-all
func (h *UserHandler) GetAll(c *gin.Context) {
	entities, err := h.userUseCase.GetAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusOK, dtos.ErrorResponse(err.Error()))
		return
	}
	c.JSON(http.StatusOK, dtos.SuccessResponse(entities))
}

// Logout handles POST /api/users/logout
func (h *UserHandler) Logout(c *gin.Context) {
	userSession := middlewares.GetUserSession(c)
	h.userUseCase.Logout(c.Request.Context(), userSession)
	c.JSON(http.StatusOK, dtos.SuccessResponse(nil, "Logged out successfully"))
}
