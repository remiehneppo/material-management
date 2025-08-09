package repository

import (
	"context"

	"github.com/remiehneppo/material-management/internal/database"
	"github.com/remiehneppo/material-management/types"
)

var _ MaterialsProfileRepository = &materialsProfileRepository{}

type MaterialsProfileRepository interface {
	Save(ctx context.Context, materialsProfile *types.MaterialsProfile) error
	FindByID(ctx context.Context, id string) (*types.MaterialsProfile, error)
	Filter(ctx context.Context, filter *types.MaterialsProfileFilter) ([]*types.MaterialsProfile, error)
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

func (r *materialsProfileRepository) Save(ctx context.Context, materialsProfile *types.MaterialsProfile) error {
	return r.database.Save(ctx, r.collection, materialsProfile)
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
	err := r.database.Query(ctx, r.collection, filter, 0, 0, nil, &materialsProfiles)
	if err != nil {
		return nil, err
	}
	return materialsProfiles, nil
}
