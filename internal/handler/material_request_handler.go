package handler

import (
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/remiehneppo/material-management/internal/logger"
	"github.com/remiehneppo/material-management/internal/service"
	"github.com/remiehneppo/material-management/types"
)

type MaterialRequestHandler interface {
	CreateMaterialRequest(ctx *gin.Context)
	GetMaterialRequestByID(ctx *gin.Context)
	FilterMaterialRequests(ctx *gin.Context)
	ExportMaterialsRequest(ctx *gin.Context)
	UpdateNumberOfRequest(ctx *gin.Context)
}

type materialRequestHandler struct {
	materialRequestService service.MaterialsRequestService
	logger                 *logger.Logger
}

func NewMaterialRequestHandler(materialRequestService service.MaterialsRequestService, logger *logger.Logger) MaterialRequestHandler {
	return &materialRequestHandler{
		materialRequestService: materialRequestService,
		logger:                 logger,
	}
}

// CreateMaterialRequest godoc
// @Summary Create a new material request
// @Description Create a new material request with the provided details
// @Tags material-requests
// @Accept json
// @Produce json
// @Param request body types.CreateMaterialRequestReq true "Material request data"
// @Success 200 {object} types.Response{data=string} "Material request created successfully"
// @Failure 400 {object} types.Response "Invalid request data"
// @Failure 500 {object} types.Response "Internal server error"
// @Security BearerAuth
// @Router /material-requests [post]
func (h *materialRequestHandler) CreateMaterialRequest(ctx *gin.Context) {
	req := types.CreateMaterialRequestReq{}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind JSON")
		ctx.JSON(http.StatusBadRequest, types.Response{
			Status:  false,
			Message: "Invalid request data: " + err.Error(),
		})
		return
	}

	id, err := h.materialRequestService.CreateMaterialsRequest(
		ctx,
		&req,
	)
	if err != nil {
		h.logger.Error("Failed to create material request: " + err.Error())
		ctx.JSON(http.StatusInternalServerError, types.Response{
			Status:  false,
			Message: "Failed to create material request: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, types.Response{
		Status:  true,
		Message: "Material request created successfully",
		Data:    id,
	})
}

// GetMaterialRequestByID godoc
// @Summary Get material request by ID
// @Description Retrieve a specific material request using its ID
// @Tags material-requests
// @Accept json
// @Produce json
// @Param id path string true "Material Request ID"
// @Success 200 {object} types.Response{data=types.MaterialRequest} "Material request retrieved successfully"
// @Failure 400 {object} types.Response "Invalid request - ID is required"
// @Failure 500 {object} types.Response "Internal server error"
// @Security BearerAuth
// @Router /material-requests/{id} [get]
func (h *materialRequestHandler) GetMaterialRequestByID(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		h.logger.Warn("GetMaterialRequestByID: Missing ID parameter")
		ctx.JSON(http.StatusBadRequest, types.Response{
			Status:  false,
			Message: "ID is required",
		})
		return
	}

	materialRequest, err := h.materialRequestService.GetMaterialsRequest(ctx, id)
	if err != nil {
		h.logger.Error("Failed to get material request by ID: " + err.Error())
		ctx.JSON(http.StatusInternalServerError, types.Response{
			Status:  false,
			Message: "Failed to get material request by ID: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, types.Response{
		Status:  true,
		Message: "Material request retrieved successfully",
		Data:    materialRequest,
	})
}

// FilterMaterialRequests godoc
// @Summary Filter material requests
// @Description Retrieve material requests based on filter criteria
// @Tags material-requests
// @Accept json
// @Produce json
// @Param filter body types.MaterialRequestFilter true "Filter criteria for material requests"
// @Success 200 {object} types.Response{data=[]types.MaterialRequest} "Material requests filtered successfully"
// @Failure 400 {object} types.Response "Invalid request data"
// @Failure 500 {object} types.Response "Internal server error"
// @Security BearerAuth
// @Router /material-requests/filter [post]
func (h *materialRequestHandler) FilterMaterialRequests(ctx *gin.Context) {
	req := types.MaterialRequestFilter{}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind JSON")
		ctx.JSON(http.StatusBadRequest, types.Response{
			Status:  false,
			Message: "Invalid request data: " + err.Error(),
		})
		return
	}

	materialRequests, err := h.materialRequestService.FilterMaterialsRequests(ctx, &req)
	if err != nil {
		h.logger.Error("Failed to filter material requests: " + err.Error())
		ctx.JSON(http.StatusInternalServerError, types.Response{
			Status:  false,
			Message: "Failed to filter material requests: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, types.Response{
		Status:  true,
		Message: "Material requests filtered successfully",
		Data:    materialRequests,
	})
}

// ExportMaterialsRequest godoc
// @Summary Export material request to DOCX
// @Description Export a material request to a downloadable DOCX document
// @Tags material-requests
// @Accept json
// @Produce application/vnd.openxmlformats-officedocument.wordprocessingml.document
// @Param export body types.MaterialRequestExport true "Export request data"
// @Success 200 {file} file "DOCX file download"
// @Failure 400 {object} types.Response "Invalid request data"
// @Failure 500 {object} types.Response "Internal server error"
// @Security BearerAuth
// @Router /material-requests/export [post]
func (h *materialRequestHandler) ExportMaterialsRequest(ctx *gin.Context) {
	exportReq := types.MaterialRequestExport{}
	if err := ctx.ShouldBindJSON(&exportReq); err != nil {
		h.logger.Error("Failed to bind JSON")
		ctx.JSON(http.StatusBadRequest, types.Response{
			Status:  false,
			Message: "Invalid request data: " + err.Error(),
		})
		return
	}

	file, err := h.materialRequestService.ExportMaterialsRequest(ctx, &exportReq)
	if err != nil {
		h.logger.Error("Failed to export material request: " + err.Error())
		ctx.JSON(http.StatusInternalServerError, types.Response{
			Status:  false,
			Message: "Failed to export material request: " + err.Error(),
		})
		return
	}
	defer file.Close()

	// Get file info for content length
	fileInfo, err := file.Stat()
	if err != nil {
		h.logger.Error("Failed to get file info: " + err.Error())
		ctx.JSON(http.StatusInternalServerError, types.Response{
			Status:  false,
			Message: "Failed to get file info",
		})
		return
	}

	// Set headers for file download
	ctx.Header("Content-Type", "application/vnd.openxmlformats-officedocument.wordprocessingml.document")
	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filepath.Base(file.Name())))
	ctx.Header("Content-Transfer-Encoding", "binary")
	ctx.Header("Expires", "0")
	ctx.Header("Cache-Control", "must-revalidate")
	ctx.Header("Pragma", "public")
	ctx.Header("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))

	// Stream the file to the user
	ctx.DataFromReader(http.StatusOK, fileInfo.Size(), "application/vnd.openxmlformats-officedocument.wordprocessingml.document", file, nil)
}

// UpdateNumberOfRequest godoc
// @Summary Update number of material requests
// @Description Update the number of requests for materials
// @Tags material-requests
// @Accept json
// @Produce json
// @Param request body types.UpdateNumberOfRequestReq true "Update number of request data"
// @Success 200 {object} types.Response "Number of requests updated successfully"
// @Failure 400 {object} types.Response "Invalid request data"
// @Failure 500 {object} types.Response "Internal server error"
// @Security BearerAuth
// @Router /material-requests/update-number [put]
func (h *materialRequestHandler) UpdateNumberOfRequest(ctx *gin.Context) {
	req := types.UpdateNumberOfRequestReq{}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind JSON")
		ctx.JSON(http.StatusBadRequest, types.Response{
			Status:  false,
			Message: "Invalid request data: " + err.Error(),
		})
		return
	}

	err := h.materialRequestService.UpdateNumberOfRequest(ctx, req)
	if err != nil {
		h.logger.Error("Failed to update number of requests: " + err.Error())
		ctx.JSON(http.StatusInternalServerError, types.Response{
			Status:  false,
			Message: "Failed to update number of requests: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, types.Response{
		Status:  true,
		Message: "Number of requests updated successfully",
	})
}
