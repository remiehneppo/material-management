package repository

import (
	"context"

	"github.com/remiehneppo/material-management/internal/database"
	"github.com/remiehneppo/material-management/types"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type EquipmentMachineryRepo interface {
	Save(ctx context.Context, equipmentMachinery *types.EquipmentMachinery) error
	FindByID(ctx context.Context, id string) (*types.EquipmentMachinery, error)
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

func (r *equipmentMachineryRepo) Save(ctx context.Context, equipmentMachinery *types.EquipmentMachinery) error {
	return r.database.Save(ctx, r.collection, equipmentMachinery)
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
		bsonFilter["name"] = bson.M{"$regex": filter.Name, "$options": "i"}
	}
	if filter.Sector != "" {
		bsonFilter["sector"] = filter.Sector
	}
	err := r.database.Query(ctx, r.collection, bsonFilter, 0, 0, nil, &equipmentMachineries)
	if err != nil {
		return nil, err
	}
	return equipmentMachineries, nil
}
