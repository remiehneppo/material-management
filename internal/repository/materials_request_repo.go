package repository

import (
	"context"

	"github.com/remiehneppo/material-management/internal/database"
	"github.com/remiehneppo/material-management/types"
)

var _ MaterialsRequestRepository = &materialsRequestRepository{}

type MaterialsRequestRepository interface {
	Save(ctx context.Context, materialsRequest *types.MaterialRequest) error
	FindByID(ctx context.Context, id string) (*types.MaterialRequest, error)
	Filter(ctx context.Context, filter *types.MaterialRequestFilter) ([]*types.MaterialRequest, error)
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
	err := r.database.Query(ctx, r.collection, filter, 0, 0, nil, &materialsRequests)
	if err != nil {
		return nil, err
	}
	return materialsRequests, nil
}
