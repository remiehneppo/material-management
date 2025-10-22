package service

import (
	"context"

	"github.com/remiehneppo/material-management/internal/repository"
	"github.com/remiehneppo/material-management/types"
)

type EquipmentMachineryService interface {
	CreateEquipmentMachinery(ctx context.Context, req *types.CreateEquipmentMachineryReq) (string, error)
	FilterEquipmentMachinery(ctx context.Context, req *types.EquipmentMachineryFilter) ([]*types.EquipmentMachinery, error)
}

type equipmentMachineryService struct {
	equipmentMachineryRepo repository.EquipmentMachineryRepo
}

func NewEquipmentMachineryService(equipmentMachineryRepo repository.EquipmentMachineryRepo) EquipmentMachineryService {
	return &equipmentMachineryService{
		equipmentMachineryRepo: equipmentMachineryRepo,
	}
}

func (s *equipmentMachineryService) CreateEquipmentMachinery(ctx context.Context, req *types.CreateEquipmentMachineryReq) (string, error) {
	equipmentMachinery := &types.EquipmentMachinery{
		Name:   req.Name,
		Sector: req.Sector,
	}

	id, err := s.equipmentMachineryRepo.Save(
		ctx,
		equipmentMachinery,
	)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (s *equipmentMachineryService) FilterEquipmentMachinery(ctx context.Context, req *types.EquipmentMachineryFilter) ([]*types.EquipmentMachinery, error) {
	return s.equipmentMachineryRepo.Filter(ctx, req)
}
