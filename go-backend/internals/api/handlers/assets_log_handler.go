package handlers

import (
	"fmt"
	"net/http"

	"go-backend/internals/core/domains"
	"go-backend/internals/core/dtos"
	"go-backend/internals/core/ports"

	"github.com/gin-gonic/gin"
)

type AssetsLogHandler struct {
	*GenericHandler[domains.ASSETSLOG]
	logUseCase ports.AssetsLogUseCase
}

func NewAssetsLogHandler(logUseCase ports.AssetsLogUseCase) *AssetsLogHandler {
	return &AssetsLogHandler{
		GenericHandler: NewGenericHandler[domains.ASSETSLOG](logUseCase),
		logUseCase:     logUseCase,
	}
}

func (h *AssetsLogHandler) Export(c *gin.Context) {
	unitCode := c.Query("unitCode")
	fromDate := c.Query("fromDate")
	toDate := c.Query("toDate")

	fileBytes, err := h.logUseCase.ExportExcel(c.Request.Context(), unitCode, fromDate, toDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse(err.Error()))
		return
	}

	filename := "AssetsLog.xlsx"
	if unitCode != "" {
		filename = fmt.Sprintf("AssetsLog_%s.xlsx", unitCode)
	}
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", fileBytes)
}
