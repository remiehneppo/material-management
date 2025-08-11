package repository

import (
	"context"

	"github.com/remiehneppo/material-management/internal/database"
	"github.com/remiehneppo/material-management/types"
	"go.mongodb.org/mongo-driver/v2/bson"
)

var _ MaterialsRequestRepository = &materialsRequestRepository{}

type MaterialsRequestRepository interface {
	Save(ctx context.Context, materialsRequest *types.MaterialRequest) error
	FindByID(ctx context.Context, id string) (*types.MaterialRequest, error)
	Filter(ctx context.Context, filter *types.MaterialRequestFilter) ([]*types.MaterialRequest, error)
	Update(ctx context.Context, id string, materialsRequest *types.MaterialRequest) error
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

func (r *materialsRequestRepository) Save(ctx context.Context, materialsRequest *types.MaterialRequest) error {
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
		// filter to get all requests with equipment_machinery_ids has element value == filter.EquipmentMachineryID
		bsonFilter["equipment_machinery_ids"] = bson.M{"$in": []string{filter.EquipmentMachineryID}}
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

func (r *materialsRequestRepository) Update(ctx context.Context, id string, materialsRequest *types.MaterialRequest) error {
	return r.database.Update(ctx, r.collection, id, materialsRequest)
}
