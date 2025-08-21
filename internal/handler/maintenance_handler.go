package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/remiehneppo/material-management/internal/service"
	"github.com/remiehneppo/material-management/types"
)

type MaintenanceHandler interface {
	CreateMaintenance(ctx *gin.Context)
	FilterMaintenance(ctx *gin.Context)
}

type maintenanceHandler struct {
	maintenanceService service.MaintenanceService
}

func NewMaintenanceHandler(maintenanceService service.MaintenanceService) MaintenanceHandler {
	return &maintenanceHandler{
		maintenanceService: maintenanceService,
	}
}

// CreateMaintenance godoc
// @Summary Create a new maintenance
// @Description Create a new maintenance record
// @Tags maintenance
// @Accept json
// @Produce json
// @Param request body types.CreateMaintenanceRequest true "Maintenance creation request"
// @Success 200 {object} types.Response{data=string} "Maintenance created successfully"
// @Failure 400 {object} types.Response "Invalid request data"
// @Failure 500 {object} types.Response "Failed to create maintenance"
// @Security BearerAuth
// @Router /maintenance [post]
func (h *maintenanceHandler) CreateMaintenance(ctx *gin.Context) {
	req := types.CreateMaintenanceRequest{}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, types.Response{
			Status:  false,
			Message: "Invalid request data: " + err.Error(),
		})
	}

	id, err := h.maintenanceService.CreateMaintenance(ctx, &req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, types.Response{
			Status:  false,
			Message: "Failed to create maintenance: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, types.Response{
		Status:  true,
		Message: "Maintenance created successfully",
		Data:    id,
	})
}

// FilterMaintenance godoc
// @Summary Filter maintenance records
// @Description Filter and retrieve maintenance records based on query parameters
// @Tags maintenance
// @Accept json
// @Produce json
// @Param sector query string false "Filter by sector"
// @Param project query string false "Filter by project"
// @Success 200 {object} types.Response{data=[]types.Maintenance} "Maintenance filtered successfully"
// @Failure 400 {object} types.Response "Invalid query parameters"
// @Failure 500 {object} types.Response "Failed to filter maintenance"
// @Security BearerAuth
// @Router /maintenance [get]
func (h *maintenanceHandler) FilterMaintenance(ctx *gin.Context) {
	req := types.MaintenanceFilter{}
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, types.Response{
			Status:  false,
			Message: "Invalid query parameters: " + err.Error(),
		})
		return
	}

	maintenances, err := h.maintenanceService.GetMaintenances(ctx, &req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, types.Response{
			Status:  false,
			Message: "Failed to filter maintenance: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, types.Response{
		Status:  true,
		Message: "Maintenance filtered successfully",
		Data:    maintenances,
	})
}
