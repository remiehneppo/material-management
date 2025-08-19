package handler

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/remiehneppo/material-management/internal/logger"
	"github.com/remiehneppo/material-management/internal/service"
	"github.com/remiehneppo/material-management/types"
)

type MaterialProfileHandler interface {
	GetMaterialsProfileByID(ctx *gin.Context)
	FilterMaterialsProfiles(ctx *gin.Context)
	UpdateMaterialsEstimateProfileBySheet(ctx *gin.Context)
}

type materialProfileHandler struct {
	materialProfileService service.MaterialsProfileService
	logger                 *logger.Logger
}

func NewMaterialProfileHandler(materialProfileService service.MaterialsProfileService, logger *logger.Logger) MaterialProfileHandler {
	return &materialProfileHandler{
		materialProfileService: materialProfileService,
		logger:                 logger,
	}
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
func (h *materialProfileHandler) GetMaterialsProfileByID(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		h.logger.Warn("GetMaterialsProfileByID: Missing ID parameter")
		ctx.JSON(http.StatusBadRequest, types.Response{
			Status:  false,
			Message: "ID is required",
		})
		return
	}

	materialsProfile, err := h.materialProfileService.GetMaterialsProfile(ctx, id)
	if err != nil {
		h.logger.Error("GetMaterialsProfileByID: Failed to retrieve materials profile", "id", id, "error", err)
		ctx.JSON(http.StatusNotFound, types.Response{
			Status:  false,
			Message: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, types.Response{
		Status:  true,
		Message: "Materials profile retrieved successfully",
		Data:    materialsProfile,
	})
}

// FilterMaterialsProfiles godoc
// @Summary Filter materials profiles
// @Description Retrieve materials profiles based on filter criteria
// @Tags materials-profiles
// @Accept json
// @Produce json
// @Param sector query string false "Sector filter"
// @Param project query string false "Project filter"
// @Param maintenance_tier query string false "Maintenance tier filter"
// @Param maintenance_number query string false "Maintenance number filter"
// @Param equipment_machinery_name query string false "Equipment machinery name filter"
// @Success 200 {object} types.Response{data=[]types.MaterialsProfile} "Materials profiles retrieved successfully"
// @Failure 400 {object} types.Response "Invalid request"
// @Failure 500 {object} types.Response "Internal server error"
// @Security BearerAuth
// @Router /materials-profiles [get]
func (h *materialProfileHandler) FilterMaterialsProfiles(ctx *gin.Context) {
	request := &types.MaterialsProfileFilterRequest{}
	if err := ctx.ShouldBindQuery(request); err != nil {
		h.logger.Warn("FilterMaterialsProfiles: Invalid query parameters", "error", err)
		ctx.JSON(http.StatusBadRequest, types.Response{
			Status:  false,
			Message: err.Error(),
		})
		return
	}

	materialsProfiles, err := h.materialProfileService.GetMaterialsProfiles(ctx, request)
	if err != nil {
		h.logger.Error("FilterMaterialsProfiles: Failed to retrieve materials profiles", "error", err)
		ctx.JSON(http.StatusInternalServerError, types.Response{
			Status:  false,
			Message: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, types.Response{
		Status:  true,
		Message: "Materials profiles retrieved successfully",
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
func (h *materialProfileHandler) UpdateMaterialsEstimateProfileBySheet(ctx *gin.Context) {
	var request types.UploadEstimateSheetRequest

	// Get file from form
	file, err := ctx.FormFile("file")
	if err != nil {
		h.logger.Warn("UpdateMaterialsEstimateProfileBySheet: Failed to get file from form", "error", err)
		ctx.JSON(http.StatusBadRequest, types.Response{
			Status:  false,
			Message: "File is required: " + err.Error(),
		})
		return
	}

	// Get JSON request data from form
	requestStr := ctx.PostForm("request")
	if requestStr == "" {
		h.logger.Warn("UpdateMaterialsEstimateProfileBySheet: Missing request data")
		ctx.JSON(http.StatusBadRequest, types.Response{
			Status:  false,
			Message: "Request data is required",
		})
		return
	}

	if err := json.Unmarshal([]byte(requestStr), &request); err != nil {
		h.logger.Warn("UpdateMaterialsEstimateProfileBySheet: Failed to parse request data", "error", err)
		ctx.JSON(http.StatusBadRequest, types.Response{
			Status:  false,
			Message: "Invalid request data: " + err.Error(),
		})
		return
	}

	// Assign the file to the request
	request.Sheet = file

	// Upload and process the sheet
	if err := h.materialProfileService.UploadEstimateSheet(ctx, &request); err != nil {
		h.logger.Error("UpdateMaterialsEstimateProfileBySheet: Failed to upload estimate sheet", "error", err)
		ctx.JSON(http.StatusInternalServerError, types.Response{
			Status:  false,
			Message: err.Error(),
		})
		return
	}

	h.logger.Info("UpdateMaterialsEstimateProfileBySheet: Successfully updated materials estimate profile", "project", request.Project, "sector", request.Sector)
	ctx.JSON(http.StatusOK, types.Response{
		Status:  true,
		Message: "Materials estimate profile updated successfully",
	})
}
