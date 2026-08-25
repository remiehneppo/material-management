package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/remiehneppo/material-management/internal/domain/materialprofile"
	"github.com/remiehneppo/material-management/internal/logger"
	"github.com/remiehneppo/material-management/internal/service"
	"github.com/remiehneppo/material-management/types"
)

type MaterialProfileHandler struct {
	materialProfileService *service.MaterialsProfileService
	logger                 *logger.Logger
	catalog                *materialprofile.Catalog
	importer               *materialprofile.Importer
}

func NewMaterialProfileHandler(materialProfileService *service.MaterialsProfileService, catalog *materialprofile.Catalog, importer *materialprofile.Importer, logger *logger.Logger) *MaterialProfileHandler {
	return &MaterialProfileHandler{
		materialProfileService: materialProfileService,
		logger:                 logger,
		catalog:                catalog,
		importer:               importer,
	}
}

// UpsertEstimatedMaterial godoc
// @Summary Persist a material in Material Profile Estimate
// @Tags materials-profiles
// @Accept json
// @Produce json
// @Param id path string true "Material Profile ID"
// @Param request body materialprofile.UpsertMaterialRequest true "Material"
// @Success 200 {object} types.Response
// @Security BearerAuth
// @Router /materials-profiles/{id}/materials [post]
func (h *MaterialProfileHandler) UpsertEstimatedMaterial(ctx *gin.Context) {
	var request materialprofile.UpsertMaterialRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, types.Response{Status: false, Message: "Thông tin vật tư không hợp lệ."})
		return
	}
	if err := h.catalog.UpsertEstimatedMaterial(ctx, ctx.Param("id"), request); err != nil {
		ctx.JSON(http.StatusBadRequest, types.Response{Status: false, Message: err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, types.Response{Status: true, Message: "Đã cập nhật vật tư dự toán."})
}

// GetMaterialsProfileByID godoc
// @Summary Get materials profile by ID
// @Description Retrieve a specific materials profile using its ID
// @Tags materials-profiles
// @Accept json
// @Produce json
// @Param id path string true "Materials Profile ID"
// @Success 200 {object} types.Response{data=types.MaterialsProfile} "Materials profile retrieved successfully"
// @Failure 400 {object} types.Response "Invalid request"
// @Failure 404 {object} types.Response "Materials profile not found"
// @Security BearerAuth
// @Router /materials-profiles/{id} [get]
func (h *MaterialProfileHandler) GetMaterialsProfileByID(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		h.logger.Warn("GetMaterialsProfileByID: Missing ID parameter")
		ctx.JSON(http.StatusBadRequest, types.Response{
			Status:  false,
			Message: "Thiếu mã hồ sơ vật tư.",
		})
		return
	}

	materialsProfile, err := h.materialProfileService.GetMaterialsProfile(ctx, id)
	if err != nil {
		h.logger.Error("GetMaterialsProfileByID: Failed to retrieve materials profile", "id", id, "error", err)
		ctx.JSON(http.StatusNotFound, types.Response{
			Status:  false,
			Message: "Không tìm thấy hồ sơ vật tư.",
		})
		return
	}

	ctx.JSON(http.StatusOK, types.Response{
		Status:  true,
		Message: "Đã tải hồ sơ vật tư.",
		Data:    materialsProfile,
	})
}

// FilterMaterialsProfiles godoc
// @Summary Filter materials profiles
// @Description Retrieve materials profiles based on filter criteria
// @Tags materials-profiles
// @Accept json
// @Produce json
// @Param filter body types.MaterialsProfileFilterRequest true "Materials profile filter"
// @Success 200 {object} types.Response{data=[]types.MaterialsProfile} "Materials profiles retrieved successfully"
// @Failure 400 {object} types.Response "Invalid request"
// @Failure 500 {object} types.Response "Internal server error"
// @Security BearerAuth
// @Router /materials-profiles [post]
func (h *MaterialProfileHandler) FilterMaterialsProfiles(ctx *gin.Context) {
	request := &types.MaterialsProfileFilterRequest{}
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, types.Response{
			Status:  false,
			Message: "Bộ lọc hồ sơ vật tư không hợp lệ.",
		})
		return
	}

	materialsProfiles, err := h.materialProfileService.GetMaterialsProfiles(ctx, request)
	if err != nil {
		h.logger.Error("FilterMaterialsProfiles: Failed to retrieve materials profiles", "error", err)
		ctx.JSON(http.StatusInternalServerError, types.Response{
			Status:  false,
			Message: "Không thể tải danh sách hồ sơ vật tư.",
		})
		return
	}

	ctx.JSON(http.StatusOK, types.Response{
		Status:  true,
		Message: "Đã tải danh sách hồ sơ vật tư.",
		Data:    materialsProfiles,
	})
}

// UpdateMaterialsEstimateProfileBySheet godoc
// @Summary Update materials estimate profile by uploading a sheet
// @Description Upload an Excel sheet to update materials estimate profile
// @Tags materials-profiles
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "Excel file to upload"
// @Param request formData string true "JSON request data containing project, maintenance_tier, maintenance_number, sheet_name, sector"
// @Success 200 {object} types.Response "Materials estimate profile updated successfully"
// @Failure 400 {object} types.Response "Invalid request"
// @Failure 500 {object} types.Response "Internal server error"
// @Security BearerAuth
// @Router /materials-profiles/upload-estimate [post]
func (h *MaterialProfileHandler) UpdateMaterialsEstimateProfileBySheet(ctx *gin.Context) {
	var request types.UploadEstimateSheetRequest

	// Get file from form
	file, err := ctx.FormFile("file")
	if err != nil {
		h.logger.Warn("UpdateMaterialsEstimateProfileBySheet: Failed to get file from form", "error", err)
		ctx.JSON(http.StatusBadRequest, types.Response{
			Status:  false,
			Message: "Vui lòng chọn tệp Excel cần tải lên.",
		})
		return
	}

	// Get JSON request data from form
	requestStr := ctx.PostForm("request")
	if requestStr == "" {
		h.logger.Warn("UpdateMaterialsEstimateProfileBySheet: Missing request data")
		ctx.JSON(http.StatusBadRequest, types.Response{
			Status:  false,
			Message: "Thiếu thông tin nhập dữ liệu dự toán.",
		})
		return
	}

	if err := json.Unmarshal([]byte(requestStr), &request); err != nil {
		h.logger.Warn("UpdateMaterialsEstimateProfileBySheet: Failed to parse request data", "error", err)
		ctx.JSON(http.StatusBadRequest, types.Response{
			Status:  false,
			Message: "Thông tin nhập dữ liệu dự toán không hợp lệ.",
		})
		return
	}

	opened, err := file.Open()
	if err != nil {
		ctx.JSON(http.StatusBadRequest, types.Response{Status: false, Message: "Không thể mở tệp Excel."})
		return
	}
	defer opened.Close()
	if err := h.importer.Import(ctx, request.MaintenanceInstanceID, request.Sector, request.SheetName, opened); err != nil {
		h.logger.Error("UpdateMaterialsEstimateProfileBySheet: Failed to upload estimate sheet", "error", err)
		ctx.JSON(http.StatusInternalServerError, types.Response{
			Status:  false,
			Message: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, types.Response{
		Status:  true,
		Message: "Đã nhập dữ liệu dự toán từ tệp Excel.",
	})
}

// CreateNewMaterialsProfile godoc
// @Summary Create a new materials profile
// @Description Create a new materials profile for a specific maintenance instance and equipment machinery
// @Tags materials-profiles
// @Accept json
// @Produce json
// @Param materials_profile body types.CreateMaterialProfileReq true "Materials profile creation request"
// @Success 200 {object} types.Response{data=string} "Materials profile created successfully"
// @Failure 400 {object} types.Response "Invalid request"
// @Failure 500 {object} types.Response "Internal server error"
// @Security BearerAuth
// @Router /materials-profiles/create [post]
func (h *MaterialProfileHandler) CreateNewMaterialsProfile(ctx *gin.Context) {
	var request types.CreateMaterialProfileReq
	if err := ctx.ShouldBindJSON(&request); err != nil {
		h.logger.Warn("CreateNewMaterialsProfile: Invalid request data", "error", err)
		ctx.JSON(http.StatusBadRequest, types.Response{
			Status:  false,
			Message: "Thông tin hồ sơ vật tư không hợp lệ.",
		})
		return
	}

	materialsProfileID, err := h.materialProfileService.CreateMaterialsProfile(ctx, &request)
	if err != nil {
		h.logger.Error("CreateNewMaterialsProfile: Failed to create materials profile", "error", err)
		ctx.JSON(http.StatusInternalServerError, types.Response{
			Status:  false,
			Message: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, types.Response{
		Status:  true,
		Message: "Đã tạo hồ sơ vật tư.",
		Data:    materialsProfileID,
	})
}

// PaginatedMaterialsProfiles godoc
// @Summary Get paginated materials profiles
// @Description Retrieve materials profiles with pagination
// @Tags materials-profiles
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Number of items per page" default(10)
// @Success 200 {object} types.Response{data=object} "Paginated materials profiles retrieved successfully"
// @Failure 400 {object} types.Response "Invalid request"
// @Failure 500 {object} types.Response "Internal server error"
// @Security BearerAuth
// @Router /materials-profiles/paginated [get]
func (h *MaterialProfileHandler) PaginatedMaterialsProfiles(ctx *gin.Context) {
	page := ctx.DefaultQuery("page", "1")
	limit := ctx.DefaultQuery("limit", "10")

	var request types.PaginatedRequest
	var err error
	request.Page, err = strconv.ParseInt(page, 10, 64)
	if err != nil || request.Page <= 0 {
		h.logger.Warn("PaginatedMaterialsProfiles: Invalid page parameter", "page", page)
		ctx.JSON(http.StatusBadRequest, types.Response{
			Status:  false,
			Message: "Số trang không hợp lệ.",
		})
		return
	}
	request.Limit, err = strconv.ParseInt(limit, 10, 64)
	if err != nil || request.Limit <= 0 {
		h.logger.Warn("PaginatedMaterialsProfiles: Invalid limit parameter", "limit", limit)
		ctx.JSON(http.StatusBadRequest, types.Response{
			Status:  false,
			Message: "Số bản ghi trên mỗi trang không hợp lệ.",
		})
		return
	}

	materialsProfiles, total, err := h.materialProfileService.PaginatedMaterialsProfiles(ctx, &request)
	if err != nil {
		h.logger.Error("PaginatedMaterialsProfiles: Failed to retrieve paginated materials profiles", "error", err)
		ctx.JSON(http.StatusInternalServerError, types.Response{
			Status:  false,
			Message: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, types.PaginatedResponse{
		Status:  true,
		Message: "Đã tải danh sách hồ sơ vật tư.",
		Data: types.PaginatedData{
			Total: total,
			Page:  request.Page,
			Limit: request.Limit,
			Items: materialsProfiles,
		},
	})
}
