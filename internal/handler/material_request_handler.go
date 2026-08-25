package handler

import (
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/remiehneppo/material-management/internal/domain/materialrequest"
	"github.com/remiehneppo/material-management/internal/logger"
	"github.com/remiehneppo/material-management/internal/service"
	"github.com/remiehneppo/material-management/types"
)

type MaterialRequestHandler struct {
	materialRequestService *service.MaterialsRequestService
	issuer                 *materialrequest.Issuer
	logger                 *logger.Logger
}

func NewMaterialRequestHandler(materialRequestService *service.MaterialsRequestService, issuer *materialrequest.Issuer, logger *logger.Logger) *MaterialRequestHandler {
	return &MaterialRequestHandler{
		materialRequestService: materialRequestService,
		issuer:                 issuer,
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
// @Router /materials-request [post]
func (h *MaterialRequestHandler) CreateMaterialRequest(ctx *gin.Context) {
	req := types.CreateMaterialRequestReq{}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind JSON")
		ctx.JSON(http.StatusBadRequest, types.Response{
			Status:  false,
			Message: "Thông tin yêu cầu vật tư không hợp lệ.",
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
			Message: "Không thể tạo yêu cầu vật tư. Vui lòng thử lại.",
		})
		return
	}

	ctx.JSON(http.StatusOK, types.Response{
		Status:  true,
		Message: "Đã tạo yêu cầu vật tư.",
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
// @Router /materials-request/{id} [get]
func (h *MaterialRequestHandler) GetMaterialRequestByID(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		h.logger.Warn("GetMaterialRequestByID: Missing ID parameter")
		ctx.JSON(http.StatusBadRequest, types.Response{
			Status:  false,
			Message: "Thiếu mã yêu cầu vật tư.",
		})
		return
	}

	materialRequest, err := h.materialRequestService.GetMaterialsRequest(ctx, id)
	if err != nil {
		h.logger.Error("Failed to get material request by ID: " + err.Error())
		ctx.JSON(http.StatusInternalServerError, types.Response{
			Status:  false,
			Message: "Không thể tải yêu cầu vật tư.",
		})
		return
	}

	ctx.JSON(http.StatusOK, types.Response{
		Status:  true,
		Message: "Đã tải yêu cầu vật tư.",
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
// @Param page query int false "Page number for pagination" default(1)
// @Param limit query int false "Number of items per page" default(10)
// @Success 200 {object} types.Response{data=[]types.MaterialRequest} "Material requests filtered successfully"
// @Failure 400 {object} types.Response "Invalid request data"
// @Failure 500 {object} types.Response "Internal server error"
// @Security BearerAuth
// @Router /materials-request/filter [post]
func (h *MaterialRequestHandler) FilterMaterialRequests(ctx *gin.Context) {
	req := types.MaterialRequestFilter{}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind JSON")
		ctx.JSON(http.StatusBadRequest, types.Response{
			Status:  false,
			Message: "Bộ lọc yêu cầu vật tư không hợp lệ.",
		})
		return
	}
	pageQuery := ctx.DefaultQuery("page", "1")
	limitQuery := ctx.DefaultQuery("limit", "10")

	page, err := strconv.ParseInt(pageQuery, 10, 64)
	if err != nil {
		h.logger.Error("Invalid page query parameter")
		ctx.JSON(http.StatusBadRequest, types.Response{
			Status:  false,
			Message: "Số trang không hợp lệ.",
		})
		return
	}

	limit, err := strconv.ParseInt(limitQuery, 10, 64)
	if err != nil {
		h.logger.Error("Invalid limit query parameter")
		ctx.JSON(http.StatusBadRequest, types.Response{
			Status:  false,
			Message: "Số bản ghi trên mỗi trang không hợp lệ.",
		})
		return
	}

	materialRequests, total, err := h.materialRequestService.FilterMaterialsRequests(ctx, &req, page, limit)
	if err != nil {
		h.logger.Error("Failed to filter material requests: " + err.Error())
		ctx.JSON(http.StatusInternalServerError, types.Response{
			Status:  false,
			Message: "Không thể tải danh sách yêu cầu vật tư.",
		})
		return
	}

	ctx.JSON(http.StatusOK, types.PaginatedResponse{
		Status:  true,
		Message: "Đã tải danh sách yêu cầu vật tư.",
		Data: types.PaginatedData{
			Total: total,
			Limit: limit,
			Page:  page,
			Items: materialRequests,
		},
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
// @Router /materials-request/export [post]
func (h *MaterialRequestHandler) ExportMaterialsRequest(ctx *gin.Context) {
	exportReq := types.MaterialRequestExport{}
	if err := ctx.ShouldBindJSON(&exportReq); err != nil {
		h.logger.Error("Failed to bind JSON")
		ctx.JSON(http.StatusBadRequest, types.Response{
			Status:  false,
			Message: "Thông tin xuất tệp không hợp lệ.",
		})
		return
	}

	file, err := h.materialRequestService.ExportMaterialsRequest(ctx, &exportReq)
	if err != nil {
		h.logger.Error("Failed to export material request: " + err.Error())
		ctx.JSON(http.StatusInternalServerError, types.Response{
			Status:  false,
			Message: "Không thể xuất tệp yêu cầu vật tư.",
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
			Message: "Không thể đọc thông tin tệp yêu cầu vật tư.",
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

// IssueMaterialRequest godoc
// @Summary Issue a draft material request
// @Description Atomically assigns the next request number and adds requested quantities to Material Profile Reality
// @Tags material-requests
// @Accept json
// @Produce json
// @Param id path string true "Material Request ID"
// @Success 200 {object} types.Response{data=types.IssueMaterialRequestResponse} "Material request issued"
// @Failure 401 {object} types.Response "Session is not authorized"
// @Failure 403 {object} types.Response "Only the requester can issue"
// @Failure 404 {object} types.Response "Material request not found"
// @Failure 409 {object} types.Response "Material request is not a draft"
// @Failure 500 {object} types.Response "Internal server error"
// @Security BearerAuth
// @Router /materials-request/{id}/issue [post]
func (h *MaterialRequestHandler) IssueMaterialRequest(ctx *gin.Context) {
	user, ok := ctx.Value("user").(*types.User)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, types.Response{Status: false, Message: types.ErrUnauthorized.Error()})
		return
	}
	result, err := h.issuer.Issue(ctx, ctx.Param("id"), user.ID)
	if err != nil {
		h.logger.Error("Failed to issue material request: " + err.Error())
		status, message := issueErrorResponse(err)
		ctx.JSON(status, types.Response{
			Status:  false,
			Message: message,
		})
		return
	}

	ctx.JSON(http.StatusOK, types.Response{
		Status:  true,
		Message: "Đã ban hành yêu cầu vật tư.",
		Data:    result,
	})
}

func issueErrorResponse(err error) (int, string) {
	switch {
	case errors.Is(err, materialrequest.ErrRequesterMismatch):
		return http.StatusForbidden, err.Error()
	case errors.Is(err, types.ErrMaterialRequestNotFound):
		return http.StatusNotFound, err.Error()
	case errors.Is(err, materialrequest.ErrDraftRequired):
		return http.StatusConflict, err.Error()
	case errors.Is(err, types.ErrUnauthorized):
		return http.StatusUnauthorized, err.Error()
	default:
		return http.StatusInternalServerError, "Không thể ban hành yêu cầu vật tư."
	}
}

// UpdateMaterialRequest godoc
// @Summary Update a material request
// @Description Update the details of an existing material request
// @Tags material-requests
// @Accept json
// @Produce json
// @Param request body types.MaterialRequestUpdate true "Material request update data"
// @Success 200 {object} types.Response "Material request updated successfully"
// @Failure 400 {object} types.Response "Invalid request data"
// @Failure 500 {object} types.Response "Internal server error"
// @Security BearerAuth
// @Router /materials-request/update [post]
func (h *MaterialRequestHandler) UpdateMaterialRequest(ctx *gin.Context) {
	req := types.MaterialRequestUpdate{}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind JSON")
		ctx.JSON(http.StatusBadRequest, types.Response{
			Status:  false,
			Message: "Thông tin cập nhật yêu cầu vật tư không hợp lệ.",
		})
		return
	}

	err := h.materialRequestService.UpdateMaterialsRequest(ctx, &req)
	if err != nil {
		h.logger.Error("Failed to update material request: " + err.Error())
		ctx.JSON(http.StatusInternalServerError, types.Response{
			Status:  false,
			Message: "Không thể cập nhật yêu cầu vật tư.",
		})
		return
	}

	ctx.JSON(http.StatusOK, types.Response{
		Status:  true,
		Message: "Đã cập nhật yêu cầu vật tư.",
	})
}

// CancelMaterialRequest godoc
// @Summary Cancel a material request
// @Description Cancel an existing material request
// @Tags material-requests
// @Accept json
// @Produce json
// @Param id path string true "Material Request ID"
// @Success 200 {object} types.Response "Material request canceled successfully"
// @Failure 400 {object} types.Response "Invalid request - ID is required"
// @Failure 500 {object} types.Response "Internal server error"
// @Security BearerAuth
// @Router /materials-request/cancel/{id} [post]
func (h *MaterialRequestHandler) CancelMaterialRequest(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		h.logger.Warn("CancelMaterialRequest: Missing ID parameter")
		ctx.JSON(http.StatusBadRequest, types.Response{
			Status:  false,
			Message: "Thiếu mã yêu cầu vật tư.",
		})
		return
	}

	err := h.materialRequestService.DeleteMaterialsRequest(ctx, id)
	if err != nil {
		h.logger.Error("Failed to cancel material request: " + err.Error())
		ctx.JSON(http.StatusInternalServerError, types.Response{
			Status:  false,
			Message: "Không thể hủy yêu cầu vật tư.",
		})
		return
	}

	ctx.JSON(http.StatusOK, types.Response{
		Status:  true,
		Message: "Đã hủy yêu cầu vật tư.",
	})
}
