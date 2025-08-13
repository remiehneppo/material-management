package repository

import (
	"context"

	"github.com/remiehneppo/material-management/internal/database"
	"github.com/remiehneppo/material-management/types"
	"go.mongodb.org/mongo-driver/v2/bson"
)

var _ MaterialsProfileRepository = &materialsProfileRepository{}

type MaterialsProfileRepository interface {
	Save(ctx context.Context, materialsProfile *types.MaterialsProfile) (string, error)
	SaveMany(ctx context.Context, materialsProfiles []*types.MaterialsProfile) ([]string, error)
	FindByID(ctx context.Context, id string) (*types.MaterialsProfile, error)
	Filter(ctx context.Context, filter *types.MaterialsProfileFilter) ([]*types.MaterialsProfile, error)
	UpdateEstimateMaterials(ctx context.Context, id string, estimateMaterials types.MaterialsForEquipment) error
	UpdateRealityMaterials(ctx context.Context, id string, realityMaterials types.MaterialsForEquipment) error
}

type materialsProfileRepository struct {
	database   database.Database
	collection string
}

func NewMaterialsProfileRepository(db database.Database) MaterialsProfileRepository {
	return &materialsProfileRepository{
		database:   db,
		collection: "materials_profiles",
	}
}

func (r *materialsProfileRepository) Save(ctx context.Context, materialsProfile *types.MaterialsProfile) (string, error) {
	return r.database.Save(ctx, r.collection, materialsProfile)
}

func (r *materialsProfileRepository) SaveMany(ctx context.Context, materialsProfiles []*types.MaterialsProfile) ([]string, error) {
	return r.database.SaveMany(ctx, r.collection, materialsProfiles)
}

func (r *materialsProfileRepository) FindByID(ctx context.Context, id string) (*types.MaterialsProfile, error) {
	materialsProfile := &types.MaterialsProfile{}
	err := r.database.FindByID(ctx, r.collection, id, materialsProfile)
	if err != nil {
		return nil, err
	}
	return materialsProfile, nil
}

func (r *materialsProfileRepository) Filter(ctx context.Context, filter *types.MaterialsProfileFilter) ([]*types.MaterialsProfile, error) {
	var materialsProfiles []*types.MaterialsProfile
	bsonFilter := bson.M{}
	if len(filter.MaintenanceInstanceIDs) > 0 {
		bsonFilter["maintenance_instance_id"] = bson.M{"$in": filter.MaintenanceInstanceIDs}
	}
	if len(filter.EquipmentMachineryIDs) > 0 {
		bsonFilter["equipment_machinery_id"] = bson.M{"$in": filter.EquipmentMachineryIDs}
	}
	if filter.Sector != "" {
		bsonFilter["sector"] = filter.Sector
	}
	err := r.database.Query(ctx, r.collection, bsonFilter, 0, 0, nil, &materialsProfiles)
	if err != nil {
		return nil, err
	}
	return materialsProfiles, nil
}

func (r *materialsProfileRepository) UpdateEstimateMaterials(ctx context.Context, id string, estimateMaterials types.MaterialsForEquipment) error {
	materialsProfile, err := r.FindByID(ctx, id)
	if err != nil {
		return err
	}
	materialsProfile.Estimate = estimateMaterials
	err = r.database.Update(ctx, r.collection, id, materialsProfile)
	if err != nil {
		return err
	}
	return nil
}

func (r *materialsProfileRepository) UpdateRealityMaterials(ctx context.Context, id string, realityMaterials types.MaterialsForEquipment) error {
	materialsProfile, err := r.FindByID(ctx, id)
	if err != nil {
		return err
	}
	materialsProfile.Reality = realityMaterials
	err = r.database.Update(ctx, r.collection, id, materialsProfile)
	if err != nil {
		return err
	}
	return nil
}
