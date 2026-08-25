package equipment

import (
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"
	"github.com/remiehneppo/material-management/types"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Handler struct{ collection *mongo.Collection }

func NewHandler(db *mongo.Database) *Handler {
	return &Handler{collection: db.Collection("equipment_machineries")}
}

// Create godoc
// @Summary Create Equipment/Machinery
// @Tags equipment-machinery
// @Accept json
// @Produce json
// @Param request body types.CreateEquipmentMachineryReq true "Equipment/Machinery"
// @Success 201 {object} types.Response{data=string}
// @Security BearerAuth
// @Router /equipment-machinery [post]
func (h *Handler) Create(ctx *gin.Context) {
	var request types.CreateEquipmentMachineryReq
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, types.Response{Status: false, Message: "Thông tin thiết bị không hợp lệ."})
		return
	}
	if !containsSector(request.Sector) {
		ctx.JSON(http.StatusBadRequest, types.Response{Status: false, Message: types.ErrInvalidSector.Error()})
		return
	}
	result, err := h.collection.InsertOne(ctx, types.EquipmentMachinery{Name: request.Name, Sector: request.Sector})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, types.Response{Status: false, Message: "Không thể tạo thiết bị."})
		return
	}
	ctx.JSON(http.StatusCreated, types.Response{Status: true, Message: "Đã tạo thiết bị.", Data: result.InsertedID.(bson.ObjectID).Hex()})
}

// Filter godoc
// @Summary Filter Equipment/Machinery
// @Tags equipment-machinery
// @Accept json
// @Produce json
// @Param request body types.EquipmentMachineryFilter true "Filter"
// @Success 200 {object} types.Response{data=[]types.EquipmentMachinery}
// @Security BearerAuth
// @Router /equipment-machinery/filter [post]
func (h *Handler) Filter(ctx *gin.Context) {
	var request types.EquipmentMachineryFilter
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, types.Response{Status: false, Message: "Bộ lọc thiết bị không hợp lệ."})
		return
	}
	filter := bson.M{}
	if request.Name != "" {
		filter["name"] = bson.M{"$regex": regexp.QuoteMeta(request.Name), "$options": "i"}
	}
	if request.Sector != "" {
		filter["sector"] = request.Sector
	}
	cursor, err := h.collection.Find(ctx, filter, options.Find().SetSort(bson.D{{Key: "name", Value: 1}}))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, types.Response{Status: false, Message: "Không thể tải danh sách thiết bị."})
		return
	}
	var items []types.EquipmentMachinery
	if err := cursor.All(ctx, &items); err != nil {
		ctx.JSON(http.StatusInternalServerError, types.Response{Status: false, Message: "Không thể đọc danh sách thiết bị."})
		return
	}
	ctx.JSON(http.StatusOK, types.Response{Status: true, Message: "Đã tải danh sách thiết bị.", Data: items})
}

func containsSector(sector string) bool {
	for _, candidate := range types.SECTOR_LIST {
		if candidate == sector {
			return true
		}
	}
	return false
}
