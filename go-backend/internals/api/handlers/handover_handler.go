package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"go-backend/internals/api/middlewares"
	"go-backend/internals/core/domains"
	"go-backend/internals/core/dtos"
	"go-backend/internals/repository"
	"go-backend/internals/services"

	"github.com/gin-gonic/gin"
)

type HandoverHandler struct {
	handoverService *services.HandoverService
	handoverRepo    *repository.GenericRepository[domains.HANDOVER]
}

func NewHandoverHandler(handoverService *services.HandoverService, handoverRepo *repository.GenericRepository[domains.HANDOVER]) *HandoverHandler {
	return &HandoverHandler{
		handoverService: handoverService,
		handoverRepo:    handoverRepo,
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
	data, total, err := h.handoverService.GetPagedHandovers(c.Request.Context(), unitCode, userSession, &filter)
	if err != nil {
		c.JSON(http.StatusOK, dtos.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, dtos.PagedResponse(data, total))
}

func (h *HandoverHandler) Export(c *gin.Context) {
	id := c.Param("id")
	unitCode := c.Query("unitCode")

	pdfBytes, err := h.handoverService.ExportHandoverForm(c.Request.Context(), id, unitCode)
	if err != nil {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse(err.Error()))
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=Handover_%s_%s.pdf", id, unitCode))
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}

func (h *HandoverHandler) Insert(c *gin.Context) {
	var entity domains.HANDOVER
	if err := c.ShouldBindJSON(&entity); err != nil {
		c.JSON(http.StatusOK, dtos.ErrorResponse("Invalid payload"))
		return
	}

	if err := h.handoverRepo.Create(c.Request.Context(), &entity); err != nil {
		c.JSON(http.StatusOK, dtos.ErrorResponse(err.Error()))
		return
	}
	c.JSON(http.StatusOK, dtos.SuccessResponse(entity, "Inserted successfully"))
}

func (h *HandoverHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var entity domains.HANDOVER
	if err := c.ShouldBindJSON(&entity); err != nil {
		c.JSON(http.StatusOK, dtos.ErrorResponse("Invalid payload"))
		return
	}

	if err := h.handoverRepo.Update(c.Request.Context(), id, &entity); err != nil {
		c.JSON(http.StatusOK, dtos.ErrorResponse(err.Error()))
		return
	}
	c.JSON(http.StatusOK, dtos.SuccessResponse(nil, "Updated successfully"))
}

func (h *HandoverHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.handoverRepo.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusOK, dtos.ErrorResponse(err.Error()))
		return
	}
	c.JSON(http.StatusOK, dtos.SuccessResponse(nil, "Deleted successfully"))
}
