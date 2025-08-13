package service

import (
	"context"
	"time"

	"github.com/remiehneppo/material-management/internal/repository"
	"github.com/remiehneppo/material-management/types"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type MaintenanceService interface {
	GetMaintenance(ctx context.Context, id string) (*types.Maintenance, error)
	GetMaintenances(ctx context.Context, req *types.MaintenanceFilter) ([]*types.Maintenance, error)
	CreateMaintenance(ctx context.Context, maintenance *types.CreateMaintenanceRequest) (string, error)
}

type maintenanceService struct {
	maintenanceRepo repository.MaintenanceRepository
}

func NewMaintenanceService(maintenanceRepo repository.MaintenanceRepository) MaintenanceService {
	return &maintenanceService{
		maintenanceRepo: maintenanceRepo,
	}
}

func (s *maintenanceService) GetMaintenance(ctx context.Context, id string) (*types.Maintenance, error) {
	maintenance, err := s.maintenanceRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return maintenance, nil
}

func (s *maintenanceService) GetMaintenances(ctx context.Context, req *types.MaintenanceFilter) ([]*types.Maintenance, error) {
	maintenances, err := s.maintenanceRepo.Filter(ctx, req)
	if err != nil {
		return nil, err
	}
	return maintenances, nil
}

func (s *maintenanceService) CreateMaintenance(ctx context.Context, maintenance *types.CreateMaintenanceRequest) (string, error) {
	maintenances, err := s.maintenanceRepo.Filter(ctx, &types.MaintenanceFilter{
		Project:           maintenance.Project,
		MaintenanceTier:   maintenance.MaintenanceTier,
		MaintenanceNumber: maintenance.MaintenanceNumber,
	})
	if err != nil {
		if err != mongo.ErrNoDocuments {
			return "", err
		}
	}
	if len(maintenances) > 0 {
		return "", types.ErrDuplicateMaintenance
	}
	return s.maintenanceRepo.Save(ctx, &types.Maintenance{
		Project:           maintenance.Project,
		MaintenanceTier:   maintenance.MaintenanceTier,
		MaintenanceNumber: maintenance.MaintenanceNumber,
		Year:              time.Now().Year(),
	},
	)
}
