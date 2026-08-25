package repository

import (
	"context"

	"github.com/remiehneppo/material-management/internal/database"
	"github.com/remiehneppo/material-management/types"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type MaterialsProfileRepository struct {
	database   *database.MongoDatabase
	collection string
}

func NewMaterialsProfileRepository(db *database.MongoDatabase) *MaterialsProfileRepository {
	return &MaterialsProfileRepository{
		database:   db,
		collection: "materials_profiles",
	}
}

func (r *MaterialsProfileRepository) Save(ctx context.Context, materialsProfile *types.MaterialsProfile) (string, error) {
	return r.database.Save(ctx, r.collection, materialsProfile)
}

func (r *MaterialsProfileRepository) SaveMany(ctx context.Context, materialsProfiles []*types.MaterialsProfile) ([]string, error) {
	data := make([]interface{}, len(materialsProfiles))
	for i, mp := range materialsProfiles {
		data[i] = mp
	}
	return r.database.SaveMany(ctx, r.collection, data)
}

func (r *MaterialsProfileRepository) UpdateMany(ctx context.Context, materialsProfileIds []string, materialsProfiles []*types.MaterialsProfile) error {
	data := make([]interface{}, len(materialsProfiles))
	for i, mp := range materialsProfiles {
		mp.ID = ""
		data[i] = mp
	}
	return r.database.UpdateMany(ctx, r.collection, materialsProfileIds, data)
}

func (r *MaterialsProfileRepository) FindByID(ctx context.Context, id string) (*types.MaterialsProfile, error) {
	materialsProfile := &types.MaterialsProfile{}
	err := r.database.FindByID(ctx, r.collection, id, materialsProfile)
	if err != nil {
		return nil, err
	}
	return materialsProfile, nil
}

func (r *MaterialsProfileRepository) FindByIDs(ctx context.Context, ids []string) (map[string]*types.MaterialsProfile, error) {
	objIds := make([]bson.ObjectID, len(ids))
	for i, id := range ids {
		objId, err := bson.ObjectIDFromHex(id)
		if err != nil {
			return nil, err
		}
		objIds[i] = objId
	}
	filter := bson.M{"_id": bson.M{"$in": objIds}}
	materialsProfiles := make([]*types.MaterialsProfile, 0)
	err := r.database.Query(ctx, r.collection, filter, 0, 0, nil, &materialsProfiles)
	if err != nil {
		return nil, err
	}

	result := make(map[string]*types.MaterialsProfile)
	for _, mp := range materialsProfiles {
		result[mp.ID] = mp
	}

	return result, nil
}

func (r *MaterialsProfileRepository) Filter(ctx context.Context, filter *types.MaterialsProfileFilter) ([]*types.MaterialsProfile, error) {
	var materialsProfiles []*types.MaterialsProfile
	bsonFilter := bson.M{}
	conditions := []bson.M{}
	if len(filter.MaintenanceInstanceIDs) > 0 {
		conditions = append(conditions, bson.M{"maintenance_instance_id": bson.M{"$in": filter.MaintenanceInstanceIDs}})
	}
	if len(filter.EquipmentMachineryIDs) > 0 {
		conditions = append(conditions, bson.M{"equipment_machinery_id": bson.M{"$in": filter.EquipmentMachineryIDs}})
	}
	if filter.Sector != "" {
		conditions = append(conditions, bson.M{"sector": filter.Sector})
	}
	if filter.Index != 0 {
		conditions = append(conditions, bson.M{"index": filter.Index})
	}
	if len(conditions) > 0 {
		bsonFilter["$and"] = conditions
	}
	sort := bson.D{{Key: "index", Value: 1}}
	err := r.database.Query(ctx, r.collection, bsonFilter, 0, 0, sort, &materialsProfiles)
	if err != nil {
		return nil, err
	}
	return materialsProfiles, nil
}

func (r *MaterialsProfileRepository) Paginate(ctx context.Context, filter *types.MaterialsProfileFilter, page int64, limit int64) ([]*types.MaterialsProfile, int64, error) {
	var materialsProfiles []*types.MaterialsProfile
	bsonFilter := bson.M{}
	conditions := []bson.M{}
	if len(filter.MaintenanceInstanceIDs) > 0 {
		conditions = append(conditions, bson.M{"maintenance_instance_id": bson.M{"$in": filter.MaintenanceInstanceIDs}})
	}
	if len(filter.EquipmentMachineryIDs) > 0 {
		conditions = append(conditions, bson.M{"equipment_machinery_id": bson.M{"$in": filter.EquipmentMachineryIDs}})
	}
	if filter.Sector != "" {
		conditions = append(conditions, bson.M{"sector": filter.Sector})
	}
	if len(conditions) > 0 {
		bsonFilter["$and"] = conditions
	}
	sort := bson.D{{Key: "index", Value: 1}}
	total, err := r.database.Count(ctx, r.collection, bsonFilter)
	if err != nil {
		return nil, 0, err
	}
	err = r.database.Query(ctx, r.collection, bsonFilter, page, limit, sort, &materialsProfiles)
	if err != nil {
		return nil, 0, err
	}
	return materialsProfiles, total, nil
}

func (r *MaterialsProfileRepository) UpdateEstimateMaterials(ctx context.Context, id string, estimateMaterials types.MaterialsForEquipment) error {
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
