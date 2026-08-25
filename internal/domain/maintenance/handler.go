package maintenance

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/remiehneppo/material-management/types"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type Handler struct{ collection *mongo.Collection }

func NewHandler(db *mongo.Database) *Handler {
	return &Handler{collection: db.Collection("maintenances")}
}

// Get godoc
// @Summary Get a maintenance instance
// @Tags maintenance
// @Produce json
// @Param id path string true "Maintenance ID"
// @Success 200 {object} types.Response{data=types.Maintenance}
// @Security BearerAuth
// @Router /maintenance/{id} [get]
func (h *Handler) Get(ctx *gin.Context) {
	id, err := bson.ObjectIDFromHex(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, types.Response{Status: false, Message: "Mã dự án không hợp lệ."})
		return
	}
	var item types.Maintenance
	if err := h.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&item); err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, mongo.ErrNoDocuments) {
			status = http.StatusNotFound
		}
		ctx.JSON(status, types.Response{Status: false, Message: "Không tìm thấy dự án."})
		return
	}
	ctx.JSON(http.StatusOK, types.Response{Status: true, Message: "Đã tải thông tin dự án.", Data: item})
}

// Filter godoc
// @Summary Filter maintenance instances
// @Tags maintenance
// @Accept json
// @Produce json
// @Param request body types.MaintenanceFilter true "Filter"
// @Success 200 {object} types.Response{data=[]types.Maintenance}
// @Security BearerAuth
// @Router /maintenance/filter [post]
func (h *Handler) Filter(ctx *gin.Context) {
	var request types.MaintenanceFilter
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, types.Response{Status: false, Message: "Bộ lọc dự án không hợp lệ."})
		return
	}
	filter := bson.M{}
	if request.ProjectCode != "" {
		filter["project_code"] = request.ProjectCode
	}
	if request.MaintenanceTier != "" {
		filter["maintenance_tier"] = request.MaintenanceTier
	}
	if request.MaintenanceNumber != "" {
		filter["maintenance_number"] = request.MaintenanceNumber
	}
	cursor, err := h.collection.Find(ctx, filter)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, types.Response{Status: false, Message: "Không thể tải danh sách dự án."})
		return
	}
	var items []types.Maintenance
	if err := cursor.All(ctx, &items); err != nil {
		ctx.JSON(http.StatusInternalServerError, types.Response{Status: false, Message: "Không thể đọc danh sách dự án."})
		return
	}
	ctx.JSON(http.StatusOK, types.Response{Status: true, Message: "Đã tải danh sách dự án.", Data: items})
}

// Create godoc
// @Summary Create a maintenance instance
// @Tags maintenance
// @Accept json
// @Produce json
// @Param request body types.CreateMaintenanceRequest true "Maintenance"
// @Success 200 {object} types.Response{data=string}
// @Security BearerAuth
// @Router /maintenance [post]
func (h *Handler) Create(ctx *gin.Context) {
	var request types.CreateMaintenanceRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, types.Response{Status: false, Message: "Thông tin dự án không hợp lệ."})
		return
	}
	key := bson.M{"project_code": request.ProjectCode, "maintenance_tier": request.MaintenanceTier, "maintenance_number": request.MaintenanceNumber}
	if err := h.collection.FindOne(ctx, key).Err(); !errors.Is(err, mongo.ErrNoDocuments) {
		ctx.JSON(http.StatusConflict, types.Response{Status: false, Message: types.ErrDuplicateMaintenance.Error()})
		return
	}
	result, err := h.collection.InsertOne(ctx, types.Maintenance{Project: request.Project, ProjectCode: request.ProjectCode, MaintenanceTier: request.MaintenanceTier, MaintenanceNumber: request.MaintenanceNumber, Year: time.Now().Year()})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, types.Response{Status: false, Message: "Không thể tạo dự án."})
		return
	}
	ctx.JSON(http.StatusOK, types.Response{Status: true, Message: "Đã tạo dự án.", Data: result.InsertedID.(bson.ObjectID).Hex()})
}
