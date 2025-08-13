package service

import (
	"context"
	"time"

	"github.com/remiehneppo/material-management/internal/repository"
	"github.com/remiehneppo/material-management/types"
	"github.com/remiehneppo/material-management/utils"
)

type MaterialsRequestService interface {
	CreateMaterialsRequest(ctx context.Context, request *types.CreateMaterialRequestReq) (string, error)
	GetMaterialsRequest(ctx context.Context, id string) (*types.MaterialRequestResponse, error)
	GetMaterialsRequests(ctx context.Context, req *types.MaterialRequestFilter) ([]*types.MaterialRequestResponse, error)
	UpdateMaterialsRequest(ctx context.Context, id string, request *types.MaterialRequestUpdate) error
	DeleteMaterialsRequest(ctx context.Context, id string) error
	UpdateNumberOfRequest(ctx context.Context, id string, numOfRequest int) error
}

type materialsRequestService struct {
	materialsRequestRepo   repository.MaterialsRequestRepository
	materialsProfileRepo   repository.MaterialsProfileRepository
	maintenanceRepo        repository.MaintenanceRepository
	equipmentMachineryRepo repository.EquipmentMachineryRepo
}

func NewMaterialsRequestService(
	materialsRequestRepo repository.MaterialsRequestRepository,
	materialsProfileRepo repository.MaterialsProfileRepository,
	maintenanceRepo repository.MaintenanceRepository,
	equipmentMachineryRepo repository.EquipmentMachineryRepo,
) MaterialsRequestService {
	return &materialsRequestService{
		materialsRequestRepo:   materialsRequestRepo,
		materialsProfileRepo:   materialsProfileRepo,
		maintenanceRepo:        maintenanceRepo,
		equipmentMachineryRepo: equipmentMachineryRepo,
	}
}

func (s *materialsRequestService) CreateMaterialsRequest(ctx context.Context, request *types.CreateMaterialRequestReq) (string, error) {

	maintenance, err := s.maintenanceRepo.Filter(ctx, &types.MaintenanceFilter{
		Project:           request.Project,
		MaintenanceTier:   request.MaintenanceTier,
		MaintenanceNumber: request.MaintenanceNumber,
	})
	if err != nil {
		return "", err
	}
	if len(maintenance) != 1 {
		return "", types.ErrMaintenanceNotFound
	}
	emIDs := make([]string, 0, len(request.EquipmentMachineryIDs))
	for equipmentID := range request.MaterialsForEquipment {
		emIDs = append(emIDs, equipmentID)
	}
	eqs, err := s.equipmentMachineryRepo.FindByIDs(ctx, emIDs)
	if err != nil {
		return "", err
	}
	if len(eqs) != len(emIDs) {
		return "", types.ErrSomeEquipmentMachineryNotFound
	}

	materialsRequest := &types.MaterialRequest{
		MaintenanceInstanceID: maintenance[0].ID,
		Sector:                request.Sector,
		Description:           request.Description,
		MaterialsForEquipment: request.MaterialsForEquipment,
		RequestedBy:           ctx.Value("user_id").(string),
		RequestedAt:           time.Now().Unix(),
	}
	return s.materialsRequestRepo.Save(ctx, materialsRequest)
}

func (s *materialsRequestService) GetMaterialsRequest(ctx context.Context, id string) (*types.MaterialRequestResponse, error) {
	materialsRequest, err := s.materialsRequestRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	emIDs := make([]string, 0, len(materialsRequest.MaterialsForEquipment))
	for equipmentID := range materialsRequest.MaterialsForEquipment {
		emIDs = append(emIDs, equipmentID)
	}
	maintenance, err := s.maintenanceRepo.FindByID(ctx, materialsRequest.MaintenanceInstanceID)
	if err != nil {
		return nil, err
	}
	equipmentMachineries, err := s.equipmentMachineryRepo.FindByIDs(ctx, emIDs)
	if err != nil {
		return nil, err
	}
	materialsForEquipment := make(map[string]types.MaterialsForEquipmentResponse)
	for emID, equipmentMachinery := range equipmentMachineries {
		materialsForEquipment[equipmentMachinery.ID] = types.MaterialsForEquipmentResponse{
			ConsumableSupplies:     materialsRequest.MaterialsForEquipment[emID].ConsumableSupplies,
			ReplacementMaterials:   materialsRequest.MaterialsForEquipment[emID].ReplacementMaterials,
			EquipmentMachineryName: equipmentMachinery.Name,
		}
	}
	materialsRequestResponse := &types.MaterialRequestResponse{
		ID:                    materialsRequest.ID,
		Project:               maintenance.Project,
		MaintenanceTier:       maintenance.MaintenanceTier,
		MaintenanceNumber:     maintenance.MaintenanceNumber,
		Year:                  maintenance.Year,
		NumOfRequest:          materialsRequest.NumOfRequest,
		Sector:                materialsRequest.Sector,
		Description:           materialsRequest.Description,
		MaterialsForEquipment: materialsForEquipment,
		RequestedBy:           materialsRequest.RequestedBy,
		RequestedAt:           materialsRequest.RequestedAt,
	}
	return materialsRequestResponse, nil
}

func (s *materialsRequestService) GetMaterialsRequests(ctx context.Context, req *types.MaterialRequestFilter) ([]*types.MaterialRequestResponse, error) {
	materialsRequests, err := s.materialsRequestRepo.Filter(ctx, req)
	if err != nil {
		return nil, err
	}

	maintenance, err := s.maintenanceRepo.FindByID(ctx, req.MaintenanceInstanceID)
	if err != nil {
		return nil, err
	}

	uniqueEmIDs := make(map[string]struct{})
	for _, materialsRequest := range materialsRequests {
		for emID := range materialsRequest.MaterialsForEquipment {
			uniqueEmIDs[emID] = struct{}{}
		}
	}

	equipmentMachineries, err := s.equipmentMachineryRepo.FindByIDs(ctx, utils.MapKeys(uniqueEmIDs))
	if err != nil {
		return nil, err
	}
	materialsRequestResponses := make([]*types.MaterialRequestResponse, 0, len(materialsRequests))
	for _, materialsRequest := range materialsRequests {
		materialsForEquipment := make(map[string]types.MaterialsForEquipmentResponse)
		for emID, equipmentMachinery := range equipmentMachineries {
			if materials, ok := materialsRequest.MaterialsForEquipment[emID]; ok {
				materialsForEquipment[equipmentMachinery.ID] = types.MaterialsForEquipmentResponse{
					ConsumableSupplies:     materials.ConsumableSupplies,
					ReplacementMaterials:   materials.ReplacementMaterials,
					EquipmentMachineryName: equipmentMachinery.Name,
				}
			}
		}
		materialsRequestResponse := &types.MaterialRequestResponse{
			ID:                    materialsRequest.ID,
			Project:               maintenance.Project,
			MaintenanceTier:       maintenance.MaintenanceTier,
			MaintenanceNumber:     maintenance.MaintenanceNumber,
			Year:                  maintenance.Year,
			Sector:                materialsRequest.Sector,
			Description:           materialsRequest.Description,
			NumOfRequest:          materialsRequest.NumOfRequest,
			MaterialsForEquipment: materialsForEquipment,
			RequestedBy:           materialsRequest.RequestedBy,
			RequestedAt:           materialsRequest.RequestedAt,
		}
		materialsRequestResponses = append(materialsRequestResponses, materialsRequestResponse)
	}
	return materialsRequestResponses, nil
}

func (s *materialsRequestService) UpdateMaterialsRequest(ctx context.Context, id string, request *types.MaterialRequestUpdate) error {
	materialsRequest, err := s.materialsRequestRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if materialsRequest.NumOfRequest != 0 {
		return types.ErrUpdateAfterGotNumOfRequest
	}
	if request.Sector != "" {
		materialsRequest.Sector = request.Sector
	}
	if request.Description != "" {
		materialsRequest.Description = request.Description
	}
	if len(request.MaterialsForEquipment) > 0 {
		emIDs := make([]string, 0, len(request.MaterialsForEquipment))
		for equipmentID := range request.MaterialsForEquipment {
			emIDs = append(emIDs, equipmentID)
		}
		eqs, err := s.equipmentMachineryRepo.FindByIDs(ctx, emIDs)
		if err != nil {
			return err
		}
		if len(eqs) != len(emIDs) {
			return types.ErrSomeEquipmentMachineryNotFound
		}
		materialsRequest.MaterialsForEquipment = request.MaterialsForEquipment
	}

	return s.materialsRequestRepo.Update(ctx, id, materialsRequest)
}

func (s *materialsRequestService) DeleteMaterialsRequest(ctx context.Context, id string) error {
	materialsRequest, err := s.materialsRequestRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if materialsRequest == nil {
		return types.ErrMaterialRequestNotFound
	}

	return s.materialsRequestRepo.Delete(ctx, id)
}

func (s *materialsRequestService) UpdateNumberOfRequest(ctx context.Context, id string, numOfRequest int) error {
	materialsRequest, err := s.materialsRequestRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if materialsRequest == nil {
		return types.ErrMaterialRequestNotFound
	}

	materialsRequest.NumOfRequest = numOfRequest

	emIDs := make([]string, 0, len(materialsRequest.MaterialsForEquipment))
	for emID := range materialsRequest.MaterialsForEquipment {
		emIDs = append(emIDs, emID)
	}

	materialsProfile, err := s.materialsProfileRepo.Filter(
		ctx,
		&types.MaterialsProfileFilter{
			MaintenanceInstanceIDs: []string{materialsRequest.MaintenanceInstanceID},
			EquipmentMachineryIDs:  emIDs,
			Sector:                 materialsRequest.Sector,
		},
	)
	if err != nil {
		return err
	}
	if len(materialsProfile) != len(emIDs) {
		return types.ErrSomeEquipmentMachineryNotFound
	}
	for _, profile := range materialsProfile {
		for _, emID := range emIDs {
			materials := materialsRequest.MaterialsForEquipment[emID]
			for _, consumableSupplies := range materials.ConsumableSupplies {
				cs, ok := profile.Reality.ConsumableSupplies[consumableSupplies.Name]
				if !ok {
					profile.Reality.ConsumableSupplies[consumableSupplies.Name] = consumableSupplies
				}
				cs.Quantity += consumableSupplies.Quantity
			}
		}
		err = s.materialsProfileRepo.UpdateRealityMaterials(ctx, profile.ID, profile.Reality)
	}

	return s.materialsRequestRepo.Update(ctx, id, materialsRequest)
}
