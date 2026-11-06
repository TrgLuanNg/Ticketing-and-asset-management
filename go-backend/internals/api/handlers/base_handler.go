package handlers

import (
	"net/http"

	"go-backend/internals/api/middlewares"
	"go-backend/internals/core/dtos"
	"go-backend/internals/core/ports"

	"github.com/gin-gonic/gin"
)

type GenericHandler[T any] struct {
	useCase ports.GenericUseCase[T]
}

func NewGenericHandler[T any](useCase ports.GenericUseCase[T]) *GenericHandler[T] {
	return &GenericHandler[T]{useCase: useCase}
}

func (h *GenericHandler[T]) Search(c *gin.Context) {
	var filter dtos.Filter
	_ = c.ShouldBindJSON(&filter)

	data, total, err := h.useCase.Search(c.Request.Context(), &filter)
	if err != nil {
		c.JSON(http.StatusOK, dtos.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, dtos.PagedResponse(data, total))
}

func (h *GenericHandler[T]) GetAll(c *gin.Context) {
	data, err := h.useCase.GetAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusOK, dtos.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, dtos.SuccessResponse(data))
}

func (h *GenericHandler[T]) GetById(c *gin.Context) {
	id := c.Param("id")
	entity, err := h.useCase.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, dtos.ErrorResponse("Not found"))
		return
	}

	c.JSON(http.StatusOK, dtos.SuccessResponse(entity))
}

func (h *GenericHandler[T]) Insert(c *gin.Context) {
	var entity T
	if err := c.ShouldBindJSON(&entity); err != nil {
		c.JSON(http.StatusOK, dtos.ErrorResponse("Invalid payload"))
		return
	}

	userSession := middlewares.GetUserSession(c)
	if err := h.useCase.Create(c.Request.Context(), userSession, &entity); err != nil {
		c.JSON(http.StatusOK, dtos.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, dtos.SuccessResponse(entity, "Inserted successfully"))
}

func (h *GenericHandler[T]) Update(c *gin.Context) {
	id := c.Param("id")
	var entity T
	if err := c.ShouldBindJSON(&entity); err != nil {
		c.JSON(http.StatusOK, dtos.ErrorResponse("Invalid payload"))
		return
	}

	userSession := middlewares.GetUserSession(c)
	if err := h.useCase.Update(c.Request.Context(), userSession, id, &entity); err != nil {
		c.JSON(http.StatusOK, dtos.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, dtos.SuccessResponse(entity, "Updated successfully"))
}

func (h *GenericHandler[T]) Delete(c *gin.Context) {
	id := c.Param("id")
	userSession := middlewares.GetUserSession(c)
	if err := h.useCase.Delete(c.Request.Context(), userSession, id); err != nil {
		c.JSON(http.StatusOK, dtos.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, dtos.SuccessResponse(nil, "Deleted successfully"))
}
