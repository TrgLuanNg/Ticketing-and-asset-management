package handlers

import (
	"fmt"
	"net/http"

	"go-backend/internals/core/domains"
	"go-backend/internals/core/dtos"
	"go-backend/internals/core/ports"

	"github.com/gin-gonic/gin"
)

type ProblemReportHandler struct {
	*GenericHandler[domains.PROBLEMREPORT]
	problemUseCase ports.ProblemReportUseCase
}

func NewProblemReportHandler(problemUseCase ports.ProblemReportUseCase) *ProblemReportHandler {
	return &ProblemReportHandler{
		GenericHandler: NewGenericHandler[domains.PROBLEMREPORT](problemUseCase),
		problemUseCase: problemUseCase,
	}
}

func (h *ProblemReportHandler) Export(c *gin.Context) {
	id := c.Param("id")
	pdfBytes, err := h.problemUseCase.ExportProblemReport(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse(err.Error()))
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=ProblemReport_%s.pdf", id))
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}
