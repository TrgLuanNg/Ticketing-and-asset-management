package handlers

import (
	"net/http"

	"go-backend/internals/core/dtos"
	"go-backend/internals/services"

	"github.com/gin-gonic/gin"
)

type AutoSyncHandler struct {
	syncService *services.AutoSyncService
}

func NewAutoSyncHandler(syncService *services.AutoSyncService) *AutoSyncHandler {
	return &AutoSyncHandler{syncService: syncService}
}

func (h *AutoSyncHandler) SyncAll(c *gin.Context) {
	results := h.syncService.SyncAll(c.Request.Context())
	c.JSON(http.StatusOK, dtos.SuccessResponse(results))
}

func (h *AutoSyncHandler) SyncAD(c *gin.Context) {
	ok, err := h.syncService.SyncAD(c.Request.Context())
	if err != nil || !ok {
		c.JSON(http.StatusOK, dtos.ErrorResponse("AD synchronization failed"))
		return
	}
	c.JSON(http.StatusOK, dtos.SuccessResponse(true, "AD sync completed"))
}

func (h *AutoSyncHandler) SyncPatchManager(c *gin.Context) {
	ok, err := h.syncService.SyncPatchManager(c.Request.Context())
	if err != nil || !ok {
		c.JSON(http.StatusOK, dtos.ErrorResponse("PatchManager synchronization failed"))
		return
	}
	c.JSON(http.StatusOK, dtos.SuccessResponse(true, "PatchManager sync completed"))
}

func (h *AutoSyncHandler) SyncTrendMicro(c *gin.Context) {
	ok, err := h.syncService.SyncTrendMicro(c.Request.Context())
	if err != nil || !ok {
		c.JSON(http.StatusOK, dtos.ErrorResponse("TrendMicro synchronization failed"))
		return
	}
	c.JSON(http.StatusOK, dtos.SuccessResponse(true, "TrendMicro sync completed"))
}
