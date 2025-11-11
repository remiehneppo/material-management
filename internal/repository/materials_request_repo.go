package repository

import (
	"context"

	"github.com/remiehneppo/material-management/internal/database"
	"github.com/remiehneppo/material-management/types"
	"go.mongodb.org/mongo-driver/v2/bson"
)

var _ MaterialsRequestRepository = &materialsRequestRepository{}

type MaterialsRequestRepository interface {
	Save(ctx context.Context, materialsRequest *types.MaterialRequest) (string, error)
	FindByID(ctx context.Context, id string) (*types.MaterialRequest, error)
	Filter(ctx context.Context, filter *types.MaterialRequestFilter) ([]*types.MaterialRequest, error)
	Paginate(ctx context.Context, filter *types.MaterialRequestFilter, page int64, limit int64) ([]*types.MaterialRequest, int64, error)
	GetMaterialsRequestByMaintenanceInstanceIDAndNumOfRequest(ctx context.Context, maintenanceInstanceID string, numOfRequest int) (*types.MaterialRequest, error)
	Update(ctx context.Context, id string, materialsRequest *types.MaterialRequest) error
	Delete(ctx context.Context, id string) error
}

type materialsRequestRepository struct {
	database   database.Database
	collection string
}

func NewMaterialsRequestRepository(db database.Database) MaterialsRequestRepository {
	return &materialsRequestRepository{
		database:   db,
		collection: "materials_requests",
	}
}

func (r *materialsRequestRepository) Save(ctx context.Context, materialsRequest *types.MaterialRequest) (string, error) {
	return r.database.Save(ctx, r.collection, materialsRequest)
}

func (r *materialsRequestRepository) FindByID(ctx context.Context, id string) (*types.MaterialRequest, error) {
	materialsRequest := &types.MaterialRequest{}
	err := r.database.FindByID(ctx, r.collection, id, materialsRequest)
	if err != nil {
		return nil, err
	}
	return materialsRequest, nil
}

func (r *materialsRequestRepository) Filter(ctx context.Context, filter *types.MaterialRequestFilter) ([]*types.MaterialRequest, error) {
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

func (r *materialsRequestRepository) Paginate(ctx context.Context, filter *types.MaterialRequestFilter, page int64, limit int64) ([]*types.MaterialRequest, int64, error) {
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
	err = r.database.Query(ctx, r.collection, bsonFilter, page*limit, limit, nil, &materialsRequests)
	if err != nil {
		return nil, 0, err
	}
	return materialsRequests, total, nil
}

func (r *materialsRequestRepository) GetMaterialsRequestByMaintenanceInstanceIDAndNumOfRequest(ctx context.Context, maintenanceInstanceID string, numOfRequest int) (*types.MaterialRequest, error) {
	bsonFilter := bson.M{
		"maintenance_instance_id": maintenanceInstanceID,
		"num_of_request":          numOfRequest,
	}
	materialsRequest := []*types.MaterialRequest{}
	err := r.database.Query(ctx, r.collection, bsonFilter, 0, 0, nil, &materialsRequest)
	if err != nil {
		return nil, err
	}
	if len(materialsRequest) == 0 {
		return nil, nil
	}
	return materialsRequest[0], nil
}

func (r *materialsRequestRepository) Update(ctx context.Context, id string, materialsRequest *types.MaterialRequest) error {
	materialsRequest.ID = ""
	return r.database.Update(ctx, r.collection, id, materialsRequest)
}

func (r *materialsRequestRepository) Delete(ctx context.Context, id string) error {
	return r.database.Delete(ctx, r.collection, id)
}
