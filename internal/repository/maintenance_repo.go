package repository

import (
	"context"

	"github.com/remiehneppo/material-management/internal/database"
	"github.com/remiehneppo/material-management/types"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type MaintenanceRepository interface {
	Save(ctx context.Context, maintenance *types.Maintenance) (string, error)
	FindByID(ctx context.Context, id string) (*types.Maintenance, error)
	FindByIDs(ctx context.Context, ids []string) (map[string]*types.Maintenance, error)
	Filter(ctx context.Context, req *types.MaintenanceFilter) ([]*types.Maintenance, error)
	Update(ctx context.Context, id string, maintenance *types.Maintenance) error
}
type maintenanceRepository struct {
	database   database.Database
	collection string
}

func NewMaintenanceRepository(db database.Database) MaintenanceRepository {
	return &maintenanceRepository{
		database:   db,
		collection: "maintenances",
	}
}

func (r *maintenanceRepository) Save(ctx context.Context, maintenance *types.Maintenance) (string, error) {
	return r.database.Save(ctx, r.collection, maintenance)
}

func (r *maintenanceRepository) FindByID(ctx context.Context, id string) (*types.Maintenance, error) {
	maintenance := &types.Maintenance{}
	err := r.database.FindByID(ctx, r.collection, id, maintenance)
	if err != nil {
		return nil, err
	}
	return maintenance, nil
}

func (r *maintenanceRepository) FindByIDs(ctx context.Context, ids []string) (map[string]*types.Maintenance, error) {
	objIds := make([]bson.ObjectID, len(ids))
	for i, id := range ids {
		objId, err := bson.ObjectIDFromHex(id)
		if err != nil {
			return nil, err
		}
		objIds[i] = objId
	}
	filter := bson.M{"_id": bson.M{"$in": objIds}}
	maintenances := make([]*types.Maintenance, 0)
	err := r.database.Query(ctx, r.collection, filter, 0, 0, nil, &maintenances)
	if err != nil {
		return nil, err
	}

	result := make(map[string]*types.Maintenance)
	for _, m := range maintenances {
		result[m.ID] = m
	}

	return result, nil
}

func (r *maintenanceRepository) Filter(ctx context.Context, req *types.MaintenanceFilter) ([]*types.Maintenance, error) {
	var maintenances []*types.Maintenance
	filter := bson.M{}
	if req.ProjectCode != "" {
		filter["project_code"] = req.ProjectCode
	}
	if req.MaintenanceTier != "" {
		filter["maintenance_tier"] = req.MaintenanceTier
	}
	if req.MaintenanceNumber != "" {
		filter["maintenance_number"] = req.MaintenanceNumber
	}

	err := r.database.Query(ctx, r.collection, filter, 0, 0, nil, &maintenances)
	if err != nil {
		return nil, err
	}

	return maintenances, nil
}

func (r *maintenanceRepository) Update(ctx context.Context, id string, maintenance *types.Maintenance) error {
	err := r.database.Update(ctx, r.collection, id, maintenance)
	if err != nil {
		return err
	}
	return nil
}
