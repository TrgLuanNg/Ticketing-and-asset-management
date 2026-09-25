package handlers

import (
	"net/http"

	"go-backend/internals/api/middlewares"
	"go-backend/internals/core/domains"
	"go-backend/internals/core/dtos"
	"go-backend/internals/repository"
	"go-backend/internals/services"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService *services.UserInfoService
	authService *services.AuthService
	userRepo    *repository.GenericRepository[domains.USERINFO]
}

func NewUserHandler(userService *services.UserInfoService, authService *services.AuthService, userRepo *repository.GenericRepository[domains.USERINFO]) *UserHandler {
	return &UserHandler{
		userService: userService,
		authService: authService,
		userRepo:    userRepo,
	}
}

// Login handles POST /api/users/login
func (h *UserHandler) Login(c *gin.Context) {
	var dto dtos.LoginDto
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusOK, dtos.ErrorResponse("Username and password are required"))
		return
	}

	res, err := h.authService.Login(c.Request.Context(), dto.Username, dto.Password)
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

	if err := h.userService.ChangePassword(c.Request.Context(), requesterId, &dto); err != nil {
		c.JSON(http.StatusOK, dtos.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, dtos.SuccessResponse(nil, "Password changed successfully"))
}

// GetHistory handles GET /api/users/:userId/history
func (h *UserHandler) GetHistory(c *gin.Context) {
	userId := c.Param("id")
	if userId == "" {
		userId = c.Param("userId")
	}
	history, err := h.userService.GetHistory(c.Request.Context(), userId)
	if err != nil {
		c.JSON(http.StatusOK, dtos.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, dtos.SuccessResponse(history))
}

// GetById handles GET /api/users/:id
func (h *UserHandler) GetById(c *gin.Context) {
	id := c.Param("id")
	user, err := h.userService.GetByID(c.Request.Context(), id)
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

	created, err := h.userService.Insert(c.Request.Context(), requesterId, &dto)
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

	if err := h.userService.Update(c.Request.Context(), requesterId, id, &dto); err != nil {
		c.JSON(http.StatusOK, dtos.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, dtos.SuccessResponse(dto, "Updated successfully"))
}

// Search handles POST /api/users/search
func (h *UserHandler) Search(c *gin.Context) {
	var filter dtos.Filter
	_ = c.ShouldBindJSON(&filter)

	entities, total, err := h.userRepo.FindPaged(c.Request.Context(), &filter)
	if err != nil {
		c.JSON(http.StatusOK, dtos.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, dtos.PagedResponse(entities, total))
}

// GetAll handles GET /api/users/get-all
func (h *UserHandler) GetAll(c *gin.Context) {
	entities, err := h.userRepo.FindAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusOK, dtos.ErrorResponse(err.Error()))
		return
	}
	c.JSON(http.StatusOK, dtos.SuccessResponse(entities))
}

// Logout handles POST /api/users/logout
func (h *UserHandler) Logout(c *gin.Context) {
	c.JSON(http.StatusOK, dtos.SuccessResponse(nil, "Logged out successfully"))
}
