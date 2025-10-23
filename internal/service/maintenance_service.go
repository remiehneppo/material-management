package service

import (
	"context"
	"time"

	"github.com/remiehneppo/material-management/internal/repository"
	"github.com/remiehneppo/material-management/types"
)

type MaintenanceService interface {
	GetMaintenance(ctx context.Context, id string) (*types.Maintenance, error)
	GetMaintenanceByIDs(ctx context.Context, ids []string) (map[string]*types.Maintenance, error)
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

func (s *maintenanceService) GetMaintenanceByIDs(ctx context.Context, ids []string) (map[string]*types.Maintenance, error) {
	maintenances, err := s.maintenanceRepo.FindByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	result := make(map[string]*types.Maintenance)
	for _, m := range maintenances {
		result[m.ID] = m
	}
	return result, nil
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
		ProjectCode:       maintenance.ProjectCode,
		MaintenanceTier:   maintenance.MaintenanceTier,
		MaintenanceNumber: maintenance.MaintenanceNumber,
	})
	if err != nil {
		return "", err
	}
	if len(maintenances) > 0 {
		return "", types.ErrDuplicateMaintenance
	}
	return s.maintenanceRepo.Save(ctx, &types.Maintenance{
		Project:           maintenance.Project,
		ProjectCode:       maintenance.ProjectCode,
		MaintenanceTier:   maintenance.MaintenanceTier,
		MaintenanceNumber: maintenance.MaintenanceNumber,
		Year:              time.Now().Year(),
	},
	)
}
