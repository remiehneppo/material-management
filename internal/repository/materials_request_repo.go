package repository

import (
	"context"

	"github.com/remiehneppo/material-management/internal/database"
	"github.com/remiehneppo/material-management/types"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type MaterialsRequestRepository struct {
	database   *database.MongoDatabase
	collection string
}

func NewMaterialsRequestRepository(db *database.MongoDatabase) *MaterialsRequestRepository {
	return &MaterialsRequestRepository{
		database:   db,
		collection: "materials_requests",
	}
}

func (r *MaterialsRequestRepository) Save(ctx context.Context, materialsRequest *types.MaterialRequest) (string, error) {
	return r.database.Save(ctx, r.collection, materialsRequest)
}

func (r *MaterialsRequestRepository) FindByID(ctx context.Context, id string) (*types.MaterialRequest, error) {
	materialsRequest := &types.MaterialRequest{}
	err := r.database.FindByID(ctx, r.collection, id, materialsRequest)
	if err != nil {
		return nil, err
	}
	return materialsRequest, nil
}

func (r *MaterialsRequestRepository) Filter(ctx context.Context, filter *types.MaterialRequestFilter) ([]*types.MaterialRequest, error) {
	var materialsRequests []*types.MaterialRequest
	bsonFilter := bson.M{}
	if filter.MaintenanceInstanceID != "" {
		bsonFilter["maintenance_instance_id"] = filter.MaintenanceInstanceID
	}
	if filter.EquipmentMachineryID != "" {
		bsonFilter["materials_for_equipment."+filter.EquipmentMachineryID] = bson.M{"$exists": true}
	}
	if filter.Sector != "" {
		bsonFilter["sector"] = filter.Sector
	}
	if filter.NumOfRequest > 0 {
		bsonFilter["num_of_request"] = filter.NumOfRequest
	}
	if filter.RequestedBy != "" {
		bsonFilter["requested_by"] = filter.RequestedBy
	}
	if filter.RequestedAtStart != 0 && filter.RequestedAtEnd != 0 {
		bsonFilter["requested_at"] = bson.M{
			"$gte": filter.RequestedAtStart,
			"$lte": filter.RequestedAtEnd,
		}
	}
	err := r.database.Query(ctx, r.collection, bsonFilter, 0, 0, nil, &materialsRequests)
	if err != nil {
		return nil, err
	}
	return materialsRequests, nil
}

func (r *MaterialsRequestRepository) Paginate(ctx context.Context, filter *types.MaterialRequestFilter, page int64, limit int64) ([]*types.MaterialRequest, int64, error) {
	var materialsRequests []*types.MaterialRequest
	bsonFilter := bson.M{}
	conditions := []bson.M{}

	if filter.MaintenanceInstanceID != "" {
		conditions = append(conditions, bson.M{"maintenance_instance_id": filter.MaintenanceInstanceID})
	}
	if filter.EquipmentMachineryID != "" {
		conditions = append(conditions, bson.M{"materials_for_equipment." + filter.EquipmentMachineryID: bson.M{"$exists": true}})
	}
	if filter.Sector != "" {
		conditions = append(conditions, bson.M{"sector": filter.Sector})
	}
	if filter.NumOfRequest > 0 {
		conditions = append(conditions, bson.M{"num_of_request": filter.NumOfRequest})
	}
	if filter.RequestedBy != "" {
		conditions = append(conditions, bson.M{"requested_by": filter.RequestedBy})
	}
	if filter.RequestedAtStart != 0 && filter.RequestedAtEnd != 0 {
		conditions = append(conditions, bson.M{
			"requested_at": bson.M{
				"$gte": filter.RequestedAtStart,
				"$lte": filter.RequestedAtEnd,
			},
		})
	}
	if len(conditions) > 0 {
		bsonFilter["$and"] = conditions
	}
	total, err := r.database.Count(ctx, r.collection, bsonFilter)
	if err != nil {
		return nil, 0, err
	}
	sort := bson.M{"requested_at": -1}
	err = r.database.Query(ctx, r.collection, bsonFilter, (page-1)*limit, limit, sort, &materialsRequests)
	if err != nil {
		return nil, 0, err
	}
	return materialsRequests, total, nil
}

func (r *MaterialsRequestRepository) UpdateDraft(ctx context.Context, id, requesterUserID string, request *types.MaterialRequestUpdate) (bool, error) {
	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return false, err
	}
	set := bson.M{}
	if request.Description != "" {
		set["description"] = request.Description
	}
	if len(request.MaterialsForEquipment) > 0 {
		set["materials_for_equipment"] = request.MaterialsForEquipment
	}
	if len(set) == 0 {
		return true, nil
	}
	result, err := r.database.DB().Collection(r.collection).UpdateOne(ctx, bson.M{
		"_id": objectID, "status": types.MATERIAL_REQUEST_DRAFT, "num_of_request": 0, "requester_user_id": requesterUserID,
	}, bson.M{"$set": set})
	if err != nil {
		return false, err
	}
	return result.MatchedCount == 1, nil
}

func (r *MaterialsRequestRepository) DeleteDraft(ctx context.Context, id, requesterUserID string) (bool, error) {
	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return false, err
	}
	result, err := r.database.DB().Collection(r.collection).DeleteOne(ctx, bson.M{
		"_id": objectID, "status": types.MATERIAL_REQUEST_DRAFT, "num_of_request": 0, "requester_user_id": requesterUserID,
	})
	if err != nil {
		return false, err
	}
	return result.DeletedCount == 1, nil
}
