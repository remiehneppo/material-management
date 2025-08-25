package repository

import (
	"context"
	"regexp"

	"github.com/remiehneppo/material-management/internal/database"
	"github.com/remiehneppo/material-management/types"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type EquipmentMachineryRepo interface {
	Save(ctx context.Context, equipmentMachinery *types.EquipmentMachinery) (string, error)
	SaveMany(ctx context.Context, equipmentMachineries []*types.EquipmentMachinery) ([]string, error)
	FindByID(ctx context.Context, id string) (*types.EquipmentMachinery, error)
	FindByIDs(ctx context.Context, ids []string) (map[string]*types.EquipmentMachinery, error)
	Filter(ctx context.Context, filter *types.EquipmentMachineryFilter) ([]*types.EquipmentMachinery, error)
}

type equipmentMachineryRepo struct {
	database   database.Database
	collection string
}

func NewEquipmentMachineryRepo(db database.Database) EquipmentMachineryRepo {
	return &equipmentMachineryRepo{
		database:   db,
		collection: "equipment_machineries",
	}
}

func (r *equipmentMachineryRepo) Save(ctx context.Context, equipmentMachinery *types.EquipmentMachinery) (string, error) {
	return r.database.Save(ctx, r.collection, equipmentMachinery)
}

func (r *equipmentMachineryRepo) SaveMany(ctx context.Context, equipmentMachineries []*types.EquipmentMachinery) ([]string, error) {
	data := make([]interface{}, len(equipmentMachineries))
	for i, em := range equipmentMachineries {
		data[i] = em
	}
	return r.database.SaveMany(ctx, r.collection, data)
}

func (r *equipmentMachineryRepo) FindByID(ctx context.Context, id string) (*types.EquipmentMachinery, error) {
	equipmentMachinery := &types.EquipmentMachinery{}
	err := r.database.FindByID(ctx, r.collection, id, equipmentMachinery)
	if err != nil {
		return nil, err
	}
	return equipmentMachinery, nil
}

func (r *equipmentMachineryRepo) Filter(ctx context.Context, filter *types.EquipmentMachineryFilter) ([]*types.EquipmentMachinery, error) {
	var equipmentMachineries []*types.EquipmentMachinery
	bsonFilter := bson.M{}
	if filter.Name != "" {
		bsonFilter["name"] = bson.M{"$regex": regexp.QuoteMeta(filter.Name), "$options": "i"}
	}
	if filter.Sector != "" {
		bsonFilter["sector"] = filter.Sector
	}
	// sort by increase order
	sort := bson.M{"order": 1}
	err := r.database.Query(ctx, r.collection, bsonFilter, 0, 0, sort, &equipmentMachineries)
	if err != nil {
		return nil, err
	}
	return equipmentMachineries, nil
}

func (r *equipmentMachineryRepo) FindByIDs(ctx context.Context, ids []string) (map[string]*types.EquipmentMachinery, error) {
	objIds := make([]bson.ObjectID, len(ids))
	for i, id := range ids {
		objId, err := bson.ObjectIDFromHex(id)
		if err != nil {
			return nil, err
		}
		objIds[i] = objId
	}

	var equipmentMachineries []*types.EquipmentMachinery
	err := r.database.Query(ctx, r.collection, bson.M{"_id": bson.M{"$in": objIds}}, 0, 0, nil, &equipmentMachineries)
	if err != nil {
		return nil, err
	}

	result := make(map[string]*types.EquipmentMachinery)
	for _, em := range equipmentMachineries {
		result[em.ID] = em
	}
	return result, nil
}
