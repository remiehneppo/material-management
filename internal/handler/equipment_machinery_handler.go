package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/remiehneppo/material-management/internal/service"
	"github.com/remiehneppo/material-management/types"
)

type EquipmentMachineryHandler interface {
	CreateEquipmentMachinery(ctx *gin.Context)
	FilterEquipmentMachinery(ctx *gin.Context)
}

type equipmentMachineryHandler struct {
	equipmentMachineryService service.EquipmentMachineryService
}

func NewEquipmentMachineryHandler(equipmentMachineryService service.EquipmentMachineryService) EquipmentMachineryHandler {
	return &equipmentMachineryHandler{
		equipmentMachineryService: equipmentMachineryService,
	}
}

// CreateEquipmentMachinery godoc
// @Summary Create a new equipment machinery
// @Description Create a new equipment machinery with name, sector, and order
// @Tags equipment-machinery
// @Accept json
// @Produce json
// @Param request body types.CreateEquipmentMachineryReq true "Equipment machinery creation request"
// @Success 201 {object} types.Response{data=string} "Equipment machinery created successfully"
// @Failure 400 {object} types.Response "Invalid request data"
// @Failure 500 {object} types.Response "Failed to create equipment machinery"
// @Security BearerAuth
// @Router /equipment-machinery [post]
func (h *equipmentMachineryHandler) CreateEquipmentMachinery(ctx *gin.Context) {
	var req types.CreateEquipmentMachineryReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, types.Response{
			Status:  false,
			Message: "Invalid request data: " + err.Error(),
		})
		return
	}

	id, err := h.equipmentMachineryService.CreateEquipmentMachinery(ctx, &req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, types.Response{
			Status:  false,
			Message: "Failed to create equipment machinery: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, types.Response{
		Status: true,
		Data:   id,
	})
}

// FilterEquipmentMachinery godoc
// @Summary Filter equipment machinery
// @Description Filter and retrieve equipment machinery based on sector and other criteria
// @Tags equipment-machinery
// @Accept json
// @Produce json
// @Param request body types.EquipmentMachineryFilter true "Equipment machinery filter request"
// @Success 200 {object} types.Response{data=[]types.EquipmentMachinery} "Equipment machinery filtered successfully"
// @Failure 400 {object} types.Response "Invalid filter parameters"
// @Failure 500 {object} types.Response "Failed to filter equipment machinery"
// @Security BearerAuth
// @Router /equipment-machinery/filter [post]
func (h *equipmentMachineryHandler) FilterEquipmentMachinery(ctx *gin.Context) {
	var filter types.EquipmentMachineryFilter
	if err := ctx.ShouldBindJSON(&filter); err != nil {
		ctx.JSON(http.StatusBadRequest, types.Response{
			Status:  false,
			Message: "Invalid filter parameters: " + err.Error(),
		})
		return
	}

	results, err := h.equipmentMachineryService.FilterEquipmentMachinery(ctx, &filter)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, types.Response{
			Status:  false,
			Message: "Failed to filter equipment machinery: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, types.Response{
		Status: true,
		Data:   results,
	})
}
