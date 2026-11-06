package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"go-backend/internals/api/middlewares"
	"go-backend/internals/core/domains"
	"go-backend/internals/core/dtos"
	"go-backend/internals/core/ports"

	"github.com/gin-gonic/gin"
)

type HandoverHandler struct {
	handoverUseCase ports.HandoverUseCase
}

func NewHandoverHandler(handoverUseCase ports.HandoverUseCase) *HandoverHandler {
	return &HandoverHandler{
		handoverUseCase: handoverUseCase,
	}
}

func (h *HandoverHandler) Search(c *gin.Context) {
	var filter dtos.Filter
	_ = c.ShouldBindJSON(&filter)

	unitCode := ""
	for _, adv := range filter.AdvancedFilters {
		if strings.EqualFold(adv.Field, "unitCode") {
			unitCode = adv.Value
			break
		}
	}

	if unitCode == "" {
		c.JSON(http.StatusOK, dtos.PagedResponse([]dtos.HandoverDto{}, 0))
		return
	}

	userSession := middlewares.GetUserSession(c)
	data, total, err := h.handoverUseCase.GetPagedHandovers(c.Request.Context(), unitCode, userSession, &filter)
	if err != nil {
		c.JSON(http.StatusOK, dtos.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, dtos.PagedResponse(data, total))
}

func (h *HandoverHandler) GetAll(c *gin.Context) {
	data, err := h.handoverUseCase.GetAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusOK, dtos.ErrorResponse(err.Error()))
		return
	}
	c.JSON(http.StatusOK, dtos.SuccessResponse(data))
}

func (h *HandoverHandler) GetById(c *gin.Context) {
	id := c.Param("id")
	entity, err := h.handoverUseCase.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, dtos.ErrorResponse("Not found"))
		return
	}
	c.JSON(http.StatusOK, dtos.SuccessResponse(entity))
}

func (h *HandoverHandler) Export(c *gin.Context) {
	id := c.Param("id")
	unitCode := c.Query("unitCode")

	pdfBytes, err := h.handoverUseCase.ExportHandoverForm(c.Request.Context(), id, unitCode)
	if err != nil {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse(err.Error()))
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=Handover_%s_%s.pdf", id, unitCode))
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}

func (h *HandoverHandler) Insert(c *gin.Context) {
	var dto dtos.HandoverDto
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusOK, dtos.ErrorResponse("Invalid payload"))
		return
	}

	userSession := middlewares.GetUserSession(c)
	if dto.UserCreated == "" {
		dto.UserCreated = userSession
	}

	created, err := h.handoverUseCase.Insert(c.Request.Context(), userSession, &dto)
	if err != nil {
		c.JSON(http.StatusOK, dtos.ErrorResponse(err.Error()))
		return
	}
	c.JSON(http.StatusOK, dtos.SuccessResponse(created, "Inserted successfully"))
}

func (h *HandoverHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var entity domains.HANDOVER
	if err := c.ShouldBindJSON(&entity); err != nil {
		c.JSON(http.StatusOK, dtos.ErrorResponse("Invalid payload"))
		return
	}

	userSession := middlewares.GetUserSession(c)
	if err := h.handoverUseCase.Update(c.Request.Context(), userSession, id, &entity); err != nil {
		c.JSON(http.StatusOK, dtos.ErrorResponse(err.Error()))
		return
	}
	c.JSON(http.StatusOK, dtos.SuccessResponse(entity, "Updated successfully"))
}

func (h *HandoverHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	userSession := middlewares.GetUserSession(c)
	if err := h.handoverUseCase.Delete(c.Request.Context(), userSession, id); err != nil {
		c.JSON(http.StatusOK, dtos.ErrorResponse(err.Error()))
		return
	}
	c.JSON(http.StatusOK, dtos.SuccessResponse(nil, "Deleted successfully"))
}
