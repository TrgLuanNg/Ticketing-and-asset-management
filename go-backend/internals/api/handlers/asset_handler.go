package handlers

import (
	"encoding/base64"
	"net/http"
	"strconv"
	"time"

	"go-backend/internals/api/middlewares"
	"go-backend/internals/core/dtos"
	"go-backend/internals/core/ports"

	"github.com/gin-gonic/gin"
)

type AssetHandler struct {
	assetUseCase ports.AssetUseCase
}

func NewAssetHandler(assetUseCase ports.AssetUseCase) *AssetHandler {
	return &AssetHandler{assetUseCase: assetUseCase}
}

func (h *AssetHandler) SearchGrid(c *gin.Context) {
	var filter dtos.Filter
	_ = c.ShouldBindJSON(&filter)

	userSession := middlewares.GetUserSession(c)
	filter.UserSession = userSession

	data, total, err := h.assetUseCase.SearchGrid(c.Request.Context(), &filter, userSession)
	if err != nil {
		c.JSON(http.StatusOK, dtos.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, dtos.PagedResponse(data, total))
}

func (h *AssetHandler) GetAll(c *gin.Context) {
	userSession := middlewares.GetUserSession(c)
	data, err := h.assetUseCase.GetAll(c.Request.Context(), userSession)
	if err != nil {
		c.JSON(http.StatusOK, dtos.ErrorResponse(err.Error()))
		return
	}
	c.JSON(http.StatusOK, dtos.SuccessResponse(data))
}

func (h *AssetHandler) GetById(c *gin.Context) {
	id := c.Param("id")
	userSession := middlewares.GetUserSession(c)
	dto, err := h.assetUseCase.GetByID(c.Request.Context(), userSession, id)
	if err != nil {
		c.JSON(http.StatusNotFound, dtos.ErrorResponse(err.Error()))
		return
	}
	c.JSON(http.StatusOK, dtos.SuccessResponse(dto))
}

func (h *AssetHandler) Insert(c *gin.Context) {
	userSession := middlewares.GetUserSession(c)
	var dto dtos.AssetDto
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusOK, dtos.ErrorResponse(err.Error()))
		return
	}

	created, err := h.assetUseCase.Insert(c.Request.Context(), userSession, &dto)
	if err != nil {
		c.JSON(http.StatusOK, dtos.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, dtos.SuccessResponse(created, "Inserted successfully"))
}

func (h *AssetHandler) Update(c *gin.Context) {
	id := c.Param("id")
	userSession := middlewares.GetUserSession(c)

	var dto dtos.AssetDto
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusOK, dtos.ErrorResponse("Invalid payload"))
		return
	}

	if err := h.assetUseCase.Update(c.Request.Context(), userSession, id, &dto); err != nil {
		c.JSON(http.StatusOK, dtos.ErrorResponse(err.Error()))
		return
	}

	if dto.UnitCode == "" {
		dto.UnitCode = id
	}
	c.JSON(http.StatusOK, dtos.SuccessResponse(dto, "Updated successfully"))
}

func (h *AssetHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	userSession := middlewares.GetUserSession(c)

	if err := h.assetUseCase.Delete(c.Request.Context(), userSession, id); err != nil {
		c.JSON(http.StatusOK, dtos.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, dtos.SuccessResponse(nil, "Deleted successfully"))
}

func (h *AssetHandler) Export(c *gin.Context) {
	var filter dtos.Filter
	_ = c.ShouldBindJSON(&filter)

	userSession := middlewares.GetUserSession(c)
	fileBytes, err := h.assetUseCase.ExportExcel(c.Request.Context(), &filter, userSession)
	if err != nil {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse("Không tìm thấy dữ liệu để xuất Excel."))
		return
	}

	c.Header("Content-Disposition", "attachment; filename=AssetsExport.xlsx")
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", fileBytes)
}

func (h *AssetHandler) ImportFile(c *gin.Context) {
	userSession := middlewares.GetUserSession(c)
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse("File is empty or not provided."))
		return
	}

	f, err := file.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse("Failed to read uploaded file."))
		return
	}
	defer f.Close()

	res, err := h.assetUseCase.ImportExcel(c.Request.Context(), f, userSession)
	if err != nil {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse(err.Error()))
		return
	}

	if res.FailedCount > 0 && len(res.ErrorFileBytes) > 0 {
		base64File := base64.StdEncoding.EncodeToString(res.ErrorFileBytes)
		c.JSON(http.StatusOK, dtos.ResponseAPI{
			Status:  true,
			Data:    base64File,
			Message: res.Message,
		})
		return
	}

	c.JSON(http.StatusOK, dtos.SuccessResponse(nil, res.Message))
}

func (h *AssetHandler) DownloadTemplate(c *gin.Context) {
	c.FileAttachment("Templates/AssetsCatalouges.xlsx", "AssetsCatalouges.xlsx")
}

func (h *AssetHandler) GetDashboardMetrics(c *gin.Context) {
	userSession := middlewares.GetUserSession(c)
	fromDate := c.Query("fromDate")

	metrics, err := h.assetUseCase.GetDashboardMetrics(c.Request.Context(), userSession, fromDate)
	if err != nil {
		c.JSON(http.StatusOK, dtos.ErrorResponse(err.Error()))
		return
	}
	c.JSON(http.StatusOK, dtos.SuccessResponse(metrics))
}

func (h *AssetHandler) GenerateUnitCode(c *gin.Context) {
	assetType := c.Query("type")
	company := c.DefaultQuery("company", "YLVN")
	area := c.DefaultQuery("area", "South")
	purchasedDateStr := c.Query("purchasedDate")

	purchasedDate := time.Now()
	if purchasedDateStr != "" {
		if t, err := time.Parse("2006-01-02", purchasedDateStr); err == nil {
			purchasedDate = t
		}
	}

	code := h.assetUseCase.GenerateUnitCode(c.Request.Context(), assetType, purchasedDate, company, area)
	c.JSON(http.StatusOK, dtos.SuccessResponse(code))
}

func (h *AssetHandler) GetSuggestions(c *gin.Context) {
	var filter dtos.Filter
	_ = c.ShouldBindJSON(&filter)
	userSession := middlewares.GetUserSession(c)

	suggestions, err := h.assetUseCase.GetSuggestions(c.Request.Context(), &filter, userSession)
	if err != nil {
		c.JSON(http.StatusOK, dtos.ErrorResponse(err.Error()))
		return
	}
	c.JSON(http.StatusOK, dtos.SuccessResponse(suggestions))
}

func (h *AssetHandler) GetUnassignedSubAssets(c *gin.Context) {
	var filter dtos.Filter
	_ = c.ShouldBindJSON(&filter)

	if searchText := c.Query("searchText"); searchText != "" {
		filter.SearchText = searchText
	}
	if pageIdxStr := c.Query("pageIndex"); pageIdxStr != "" {
		if idx, err := strconv.Atoi(pageIdxStr); err == nil {
			filter.PageIndex = idx
		}
	}
	if pageSzStr := c.Query("pageSize"); pageSzStr != "" {
		if sz, err := strconv.Atoi(pageSzStr); err == nil {
			filter.PageSize = sz
		}
	}

	data, total, err := h.assetUseCase.GetUnassignedSubAssets(c.Request.Context(), &filter)
	if err != nil {
		c.JSON(http.StatusOK, dtos.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, dtos.PagedResponse(data, total))
}

func (h *AssetHandler) UpdateMainAsset(c *gin.Context) {
	unitCode := c.Param("id")
	if unitCode == "" {
		unitCode = c.Param("unitCode")
	}
	var dto dtos.UpdateMainAssetDto
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse("Invalid payload"))
		return
	}

	ok := h.assetUseCase.UpdateMainAsset(c.Request.Context(), unitCode, dto.MainAsset)
	if !ok {
		c.JSON(http.StatusNotFound, dtos.ErrorResponse("Asset not found"))
		return
	}

	c.JSON(http.StatusOK, dtos.SuccessResponse(nil, "Main asset updated successfully"))
}
