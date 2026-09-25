package handlers

import (
	"net/http"

	"go-backend/internals/core/dtos"
	"go-backend/internals/repository"

	"github.com/gin-gonic/gin"
)

type GenericHandler[T any] struct {
	repo *repository.GenericRepository[T]
}

func NewGenericHandler[T any](repo *repository.GenericRepository[T]) *GenericHandler[T] {
	return &GenericHandler[T]{repo: repo}
}

func (h *GenericHandler[T]) Search(c *gin.Context) {
	var filter dtos.Filter
	_ = c.ShouldBindJSON(&filter)

	data, total, err := h.repo.FindPaged(c.Request.Context(), &filter)
	if err != nil {
		c.JSON(http.StatusOK, dtos.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, dtos.PagedResponse(data, total))
}

func (h *GenericHandler[T]) GetAll(c *gin.Context) {
	data, err := h.repo.FindAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusOK, dtos.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, dtos.SuccessResponse(data))
}

func (h *GenericHandler[T]) GetById(c *gin.Context) {
	id := c.Param("id")
	entity, err := h.repo.FindByID(c.Request.Context(), id)
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

	if err := h.repo.Create(c.Request.Context(), &entity); err != nil {
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

	if err := h.repo.Update(c.Request.Context(), id, &entity); err != nil {
		c.JSON(http.StatusOK, dtos.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, dtos.SuccessResponse(entity, "Updated successfully"))
}

func (h *GenericHandler[T]) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.repo.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusOK, dtos.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, dtos.SuccessResponse(nil, "Deleted successfully"))
}
