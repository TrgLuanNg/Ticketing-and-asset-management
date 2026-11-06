package handlers

import (
	"fmt"
	"net/http"

	"go-backend/internals/api/middlewares"
	"go-backend/internals/core/domains"
	"go-backend/internals/core/dtos"
	"go-backend/internals/core/ports"

	"github.com/gin-gonic/gin"
)

type DisposalHandler struct {
	*GenericHandler[domains.DISPOSAL]
	disposalUseCase ports.DisposalUseCase
}

func NewDisposalHandler(disposalUseCase ports.DisposalUseCase) *DisposalHandler {
	return &DisposalHandler{
		GenericHandler:  NewGenericHandler[domains.DISPOSAL](disposalUseCase),
		disposalUseCase: disposalUseCase,
	}
}

func (h *DisposalHandler) TransferToWFD(c *gin.Context) {
	var req dtos.TransferWfdRequestDto
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse("Invalid payload"))
		return
	}

	userSession := middlewares.GetUserSession(c)
	count, idDisposal, err := h.disposalUseCase.TransferToWFD(c.Request.Context(), userSession, &req)
	if err != nil {
		c.JSON(http.StatusOK, dtos.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, dtos.SuccessResponse(map[string]any{
		"count":      count,
		"idDisposal": idDisposal,
	}, fmt.Sprintf("Transferred %d assets to Waiting for Disposal (%s)", count, idDisposal)))
}

func (h *DisposalHandler) ConfirmDispose(c *gin.Context) {
	var req dtos.ConfirmDisposeRequestDto
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse("Invalid payload"))
		return
	}

	userSession := middlewares.GetUserSession(c)
	count, err := h.disposalUseCase.ConfirmDispose(c.Request.Context(), userSession, &req)
	if err != nil {
		c.JSON(http.StatusOK, dtos.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, dtos.SuccessResponse(count, fmt.Sprintf("Confirmed disposal for %d assets", count)))
}

func (h *DisposalHandler) GetBatches(c *gin.Context) {
	userSession := middlewares.GetUserSession(c)
	batches, err := h.disposalUseCase.GetBatches(c.Request.Context(), userSession)
	if err != nil {
		c.JSON(http.StatusOK, dtos.ErrorResponse(err.Error()))
		return
	}
	c.JSON(http.StatusOK, dtos.SuccessResponse(batches))
}

func (h *DisposalHandler) GetBatchDetails(c *gin.Context) {
	idDisposal := c.Param("idDisposal")
	if idDisposal == "" {
		idDisposal = c.Param("id")
	}
	userSession := middlewares.GetUserSession(c)
	details, err := h.disposalUseCase.GetBatchDetails(c.Request.Context(), userSession, idDisposal)
	if err != nil {
		c.JSON(http.StatusOK, dtos.ErrorResponse(err.Error()))
		return
	}
	c.JSON(http.StatusOK, dtos.SuccessResponse(details))
}

func (h *DisposalHandler) CheckUnitCodes(c *gin.Context) {
	var req dtos.CheckUnitCodesRequestDto
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse("Invalid payload"))
		return
	}
	userSession := middlewares.GetUserSession(c)
	results, err := h.disposalUseCase.CheckUnitCodes(c.Request.Context(), userSession, &req)
	if err != nil {
		c.JSON(http.StatusOK, dtos.ErrorResponse(err.Error()))
		return
	}
	c.JSON(http.StatusOK, dtos.SuccessResponse(results))
}

func (h *DisposalHandler) ImportUnitCodes(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse("File is empty or not provided"))
		return
	}
	f, err := file.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse("Failed to read uploaded file"))
		return
	}
	defer f.Close()

	userSession := middlewares.GetUserSession(c)
	codes, err := h.disposalUseCase.ImportUnitCodes(c.Request.Context(), userSession, f)
	if err != nil {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse(err.Error()))
		return
	}
	c.JSON(http.StatusOK, dtos.SuccessResponse(codes))
}
