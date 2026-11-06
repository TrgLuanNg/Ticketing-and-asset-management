package handlers

import (
	"net/http"

	"go-backend/internals/core/dtos"
	"go-backend/internals/core/ports"

	"github.com/gin-gonic/gin"
)

type AutoSyncHandler struct {
	syncUseCase ports.AutoSyncUseCase
}

func NewAutoSyncHandler(syncUseCase ports.AutoSyncUseCase) *AutoSyncHandler {
	return &AutoSyncHandler{syncUseCase: syncUseCase}
}

func (h *AutoSyncHandler) SyncAll(c *gin.Context) {
	results := h.syncUseCase.SyncAll(c.Request.Context())
	c.JSON(http.StatusOK, dtos.SuccessResponse(results))
}

func (h *AutoSyncHandler) SyncAD(c *gin.Context) {
	ok, err := h.syncUseCase.SyncAD(c.Request.Context())
	if err != nil || !ok {
		c.JSON(http.StatusOK, dtos.ErrorResponse("AD synchronization failed"))
		return
	}
	c.JSON(http.StatusOK, dtos.SuccessResponse(true, "AD sync completed"))
}

func (h *AutoSyncHandler) SyncPatchManager(c *gin.Context) {
	ok, err := h.syncUseCase.SyncPatchManager(c.Request.Context())
	if err != nil || !ok {
		c.JSON(http.StatusOK, dtos.ErrorResponse("PatchManager synchronization failed"))
		return
	}
	c.JSON(http.StatusOK, dtos.SuccessResponse(true, "PatchManager sync completed"))
}

func (h *AutoSyncHandler) SyncTrendMicro(c *gin.Context) {
	ok, err := h.syncUseCase.SyncTrendMicro(c.Request.Context())
	if err != nil || !ok {
		c.JSON(http.StatusOK, dtos.ErrorResponse("TrendMicro synchronization failed"))
		return
	}
	c.JSON(http.StatusOK, dtos.SuccessResponse(true, "TrendMicro sync completed"))
}

func (h *AutoSyncHandler) SyncMEDetail(c *gin.Context) {
	ok, err := h.syncUseCase.SyncMEDetail(c.Request.Context())
	if err != nil || !ok {
		c.JSON(http.StatusOK, dtos.ErrorResponse("MEDetail synchronization failed"))
		return
	}
	c.JSON(http.StatusOK, dtos.SuccessResponse(true, "MEDetail sync completed"))
}

func (h *AutoSyncHandler) GetADAssets(c *gin.Context) {
	data, err := h.syncUseCase.GetADAssets(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusOK, dtos.ErrorResponse(err.Error()))
		return
	}
	c.JSON(http.StatusOK, dtos.SuccessResponse(data))
}

func (h *AutoSyncHandler) GetTMAssets(c *gin.Context) {
	data, err := h.syncUseCase.GetTMAssets(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusOK, dtos.ErrorResponse(err.Error()))
		return
	}
	c.JSON(http.StatusOK, dtos.SuccessResponse(data))
}

func (h *AutoSyncHandler) GetMEAssets(c *gin.Context) {
	data, err := h.syncUseCase.GetMEAssets(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusOK, dtos.ErrorResponse(err.Error()))
		return
	}
	c.JSON(http.StatusOK, dtos.SuccessResponse(data))
}

func (h *AutoSyncHandler) GetMEDetails(c *gin.Context) {
	data, err := h.syncUseCase.GetMEDetails(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusOK, dtos.ErrorResponse(err.Error()))
		return
	}
	c.JSON(http.StatusOK, dtos.SuccessResponse(data))
}
