package service

import (
	"context"

	"github.com/remiehneppo/material-management/internal/repository"
	"github.com/remiehneppo/material-management/types"
)

var _ MaterialsProfileService = &materialsProfileService{}

type MaterialsProfileService interface {
	GetMaterialsProfile(ctx context.Context, id string) (*types.MaterialsProfile, error)
	GetMaterialsProfiles(ctx context.Context, req *types.MaterialsProfileFilterRequest) ([]*types.MaterialsProfile, error)
	UpdateMaterialsEstimateProfile(ctx context.Context, request *types.UpdateMaterialsEstimateProfileRequest) error
	UpdateMaterialsRealityProfile(ctx context.Context, request *types.UpdateMaterialsRealityProfileRequest) error
}

type materialsProfileService struct {
	materialsProfileRepo repository.MaterialsProfileRepository
	maintenanceRepo      repository.MaintenanceRepository
}

func NewMaterialsProfileService(materialsProfileRepo repository.MaterialsProfileRepository, maintenanceRepo repository.MaintenanceRepository) MaterialsProfileService {
	return &materialsProfileService{
		materialsProfileRepo: materialsProfileRepo,
		maintenanceRepo:      maintenanceRepo,
	}
}

func (s *materialsProfileService) GetMaterialsProfile(ctx context.Context, id string) (*types.MaterialsProfile, error) {
	materialsProfile, err := s.materialsProfileRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return materialsProfile, nil
}

func (s *materialsProfileService) GetMaterialsProfiles(ctx context.Context, request *types.MaterialsProfileFilterRequest) ([]*types.MaterialsProfile, error) {

	filter := &types.MaterialsProfileFilter{
		Sector: request.Sector,
	}
	materialsProfiles, err := s.materialsProfileRepo.Filter(ctx, filter)
	if err != nil {
		return nil, err
	}
	return materialsProfiles, nil
}

func (s *materialsProfileService) UpdateMaterialsEstimateProfile(ctx context.Context, request *types.UpdateMaterialsEstimateProfileRequest) error {
	panic("UpdateMaterialsEstimateProfile not implemented")
}

func (s *materialsProfileService) UpdateMaterialsRealityProfile(ctx context.Context, request *types.UpdateMaterialsRealityProfileRequest) error {
	if request.UpdateType != types.UPDATE_TYPE_NEW && request.UpdateType != types.UPDATE_TYPE_MODIFY {
		return types.ErrInvalidUpdateType
	}

	panic("UpdateMaterialsRealityProfile not implemented")
}

func (s *materialsProfileService) getMaintenanceIDs(ctx context.Context, request *types.MaterialsProfileFilterRequest) ([]string, error) {
	ids := make([]string, 0)
	if request.MaintenanceIDs != nil {
		ids = append(ids, request.MaintenanceIDs...)
	}
	if request.Project != "" || request.MaintenanceTier != "" || request.MaintenanceNumber != "" {
		filter := &types.MaintenanceFilter{
			Project:           request.Project,
			MaintenanceTier:   request.MaintenanceTier,
			MaintenanceNumber: request.MaintenanceNumber,
		}
		maintenances, err := s.maintenanceRepo.Filter(ctx, filter)
		if err != nil {
			return nil, err
		}
		for _, maintenance := range maintenances {
			ids = append(ids, maintenance.ID)
		}
	}

	return ids, nil
}

func (s *materialsProfileService) getEquipmentMachineryIDs(ctx context.Context, request *types.MaterialsProfileFilterRequest) ([]string, error) {
	ids := make([]string, 0)
	if request.EquipmentMachineryIDs != nil {
		ids = append(ids, request.EquipmentMachineryIDs...)
	}
	if request.MaintenanceIDs != nil || request.Project != "" || request.MaintenanceTier != "" || request.MaintenanceNumber != "" {
		filter := &types.MaintenanceFilter{
			Project:           request.Project,
			MaintenanceTier:   request.MaintenanceTier,
			MaintenanceNumber: request.MaintenanceNumber,
		}
		maintenances, err := s.maintenanceRepo.Filter(ctx, filter)
		if err != nil {
			return nil, err
		}
		for _, maintenance := range maintenances {
			if maintenance.EquipmentMachineryID != "" {
				ids = append(ids, maintenance.EquipmentMachineryID)
			}
		}
	}

	return ids, nil
}
