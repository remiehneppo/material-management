package service

import (
	"context"
	"fmt"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/remiehneppo/material-management/internal/repository"
	"github.com/remiehneppo/material-management/types"
	"github.com/remiehneppo/material-management/utils"
	"github.com/xuri/excelize/v2"
)

var _ MaterialsProfileService = &materialsProfileService{}

type MaterialsProfileService interface {
	GetMaterialsProfile(ctx context.Context, id string) (*types.MaterialsProfileResponse, error)
	GetMaterialsProfiles(ctx context.Context, req *types.MaterialsProfileFilterRequest) ([]*types.MaterialsProfileResponse, error)
	UpdateMaterialsEstimateProfile(ctx context.Context, request *types.UpdateMaterialsEstimateProfileRequest) error
	UploadEstimateSheet(ctx context.Context, request *types.UploadEstimateSheetRequest) error
	PaginatedMaterialsProfiles(ctx context.Context, request *types.PaginatedRequest) ([]*types.MaterialsProfileResponse, int64, error)
	//UpdateMaterialsRealityProfile(ctx context.Context, request *types.UpdateMaterialsRealityProfileRequest) error
}

type materialsProfileService struct {
	materialsProfileRepo   repository.MaterialsProfileRepository
	maintenanceRepo        repository.MaintenanceRepository
	equipmentMachineryRepo repository.EquipmentMachineryRepo
	uploadService          UploadService
}

func NewMaterialsProfileService(
	materialsProfileRepo repository.MaterialsProfileRepository,
	maintenanceRepo repository.MaintenanceRepository,
	equipmentMachineryRepo repository.EquipmentMachineryRepo,
	uploadService UploadService,
) MaterialsProfileService {
	return &materialsProfileService{
		materialsProfileRepo:   materialsProfileRepo,
		maintenanceRepo:        maintenanceRepo,
		equipmentMachineryRepo: equipmentMachineryRepo,
		uploadService:          uploadService,
	}
}

func (s *materialsProfileService) GetMaterialsProfile(ctx context.Context, id string) (*types.MaterialsProfileResponse, error) {
	materialsProfile, err := s.materialsProfileRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	maintenance, err := s.maintenanceRepo.FindByID(ctx, materialsProfile.MaintenanceInstanceID)
	if err != nil {
		return nil, err
	}
	equipmentMachinery, err := s.equipmentMachineryRepo.FindByID(ctx, materialsProfile.EquipmentMachineryID)
	if err != nil {
		return nil, err
	}
	res := types.MaterialsProfileResponse{
		ID:                 materialsProfile.ID,
		Project:            maintenance.Project,
		ProjectCode:        maintenance.ProjectCode,
		MaintenanceTier:    maintenance.MaintenanceTier,
		MaintenanceNumber:  maintenance.MaintenanceNumber,
		Year:               maintenance.Year,
		Sector:             materialsProfile.Sector,
		EquipmentMachinery: equipmentMachinery.Name,
		IndexPath:          utils.IndexPathToString(materialsProfile.Index),
	}
	return &res, nil
}

func (s *materialsProfileService) GetMaterialsProfiles(ctx context.Context, request *types.MaterialsProfileFilterRequest) ([]*types.MaterialsProfileResponse, error) {

	filter := &types.MaterialsProfileFilter{
		Sector: request.Sector,
	}
	maintenanceIDs, err := s.getMaintenanceIDs(ctx, request)
	if err != nil {
		return nil, err
	}
	if len(maintenanceIDs) > 0 {
		filter.MaintenanceInstanceIDs = maintenanceIDs
	}
	equipmentMachineryIDs, err := s.getEquipmentMachineryIDs(ctx, request)
	if err != nil {
		return nil, err
	}
	if len(equipmentMachineryIDs) > 0 {
		filter.EquipmentMachineryIDs = equipmentMachineryIDs
	}

	materialsProfiles, err := s.materialsProfileRepo.Filter(ctx, filter)
	if err != nil {
		return nil, err
	}
	uniqueMaintenanceIDs := make(map[string]struct{})
	uniqueEquipmentMachineryIDs := make(map[string]struct{})
	materialsProfilesByID := make(map[string]*types.MaterialsProfile)
	for _, materialProfile := range materialsProfiles {
		uniqueMaintenanceIDs[materialProfile.MaintenanceInstanceID] = struct{}{}
		uniqueEquipmentMachineryIDs[materialProfile.EquipmentMachineryID] = struct{}{}
		materialsProfilesByID[materialProfile.ID] = materialProfile
	}
	maintenances, err := s.maintenanceRepo.FindByIDs(ctx, utils.MapKeys(uniqueMaintenanceIDs))
	if err != nil {
		return nil, err
	}
	equipmentMachineries, err := s.equipmentMachineryRepo.FindByIDs(ctx, utils.MapKeys(uniqueEquipmentMachineryIDs))
	if err != nil {
		return nil, err
	}
	res := make([]*types.MaterialsProfileResponse, len(materialsProfiles))
	for i, mp := range materialsProfiles {
		res[i] = &types.MaterialsProfileResponse{
			ID:                 mp.ID,
			Project:            maintenances[mp.MaintenanceInstanceID].Project,
			ProjectCode:        maintenances[mp.MaintenanceInstanceID].ProjectCode,
			MaintenanceTier:    maintenances[mp.MaintenanceInstanceID].MaintenanceTier,
			MaintenanceNumber:  maintenances[mp.MaintenanceInstanceID].MaintenanceNumber,
			Year:               maintenances[mp.MaintenanceInstanceID].Year,
			Sector:             mp.Sector,
			EquipmentMachinery: equipmentMachineries[mp.EquipmentMachineryID].Name,
			IndexPath:          utils.IndexPathToString(materialsProfilesByID[mp.ID].Index),
			Estimate:           materialsProfilesByID[mp.ID].Estimate,
			Reality:            materialsProfilesByID[mp.ID].Reality,
		}
	}
	return res, nil
}

func (s *materialsProfileService) UpdateMaterialsEstimateProfile(ctx context.Context, request *types.UpdateMaterialsEstimateProfileRequest) error {
	// TODO: Implement UpdateMaterialsEstimateProfile
	return types.ErrNotImplemented
}

func (s *materialsProfileService) UploadEstimateSheet(ctx context.Context, request *types.UploadEstimateSheetRequest) error {
	if !utils.Contains(types.SECTOR_LIST, request.Sector) {
		return types.ErrInvalidSector
	}
	if !utils.Contains(types.MAINTENANCE_TIER_LIST, request.MaintenanceTier) {
		return types.ErrInvalidMaintenanceTier
	}

	maintenance, err := s.maintenanceRepo.Filter(ctx, &types.MaintenanceFilter{
		ProjectCode:       request.ProjectCode,
		MaintenanceTier:   request.MaintenanceTier,
		MaintenanceNumber: request.MaintenanceNumber,
	})
	if err != nil {
		return err
	}
	if len(maintenance) != 1 {
		return types.ErrMaintenanceNotFound
	}

	saveDir := path.Join(
		request.MaintenanceTier+"_"+
			request.ProjectCode+"_"+
			request.MaintenanceNumber,
		time.Now().Format("2006"),
	)
	saveDir = strings.ReplaceAll(saveDir, " ", "_")

	// file name is materials_estimate_ + full date time
	fileName := fmt.Sprintf("%s_%s", request.Sector, time.Now().Format("2006-01-02"))

	sheetPath, err := s.uploadService.UploadFile(ctx, request.Sheet, saveDir, fileName)
	if err != nil {
		return err
	}

	f, err := excelize.OpenFile(sheetPath)
	if err != nil {
		return err
	}
	defer f.Close()
	rows, err := f.GetRows(request.SheetName)
	if err != nil {
		return err
	}

	var materialsProfilesMap = make(map[string]*types.MaterialsProfile)
	var equipmentNameToID = make(map[string]string)

	indexRegex := regexp.MustCompile(`^\d+(\.\d+)*$`)
	currentEquipmentMachineryName := ""
	currentMaterialType := ""
	lastIndexStr := ""
	for _, row := range rows[1:] {
		indexCell := strings.TrimSpace(row[0])
		titleCell := strings.TrimSpace(row[1])
		// check indexCell match regex like "1.1", "2.3.4", etc
		indexStr := indexRegex.FindString(indexCell)
		if indexStr != "" {
			lastIndexStr = indexStr
			currentEquipmentMachineryName = titleCell
			eqs, err := s.equipmentMachineryRepo.Filter(ctx, &types.EquipmentMachineryFilter{
				Name:   currentEquipmentMachineryName,
				Sector: request.Sector,
			})
			if err != nil {
				return err
			}
			if len(eqs) == 0 {
				// Equipment not found, create new one
				eqID, err := s.equipmentMachineryRepo.Save(ctx, &types.EquipmentMachinery{
					Name:   currentEquipmentMachineryName,
					Sector: request.Sector,
				})
				if err != nil {
					return err
				}
				equipmentNameToID[currentEquipmentMachineryName] = eqID

			} else {
				equipmentNameToID[currentEquipmentMachineryName] = eqs[0].ID
			}
			s.ensureMaterialsProfile(ctx, currentEquipmentMachineryName, materialsProfilesMap, maintenance[0].ID, equipmentNameToID[currentEquipmentMachineryName], request.Sector, lastIndexStr)
			currentMaterialType = ""
		}
		if strings.Contains(strings.ToLower(titleCell), types.LABEL_REPLACEMENT) {
			currentMaterialType = types.LABEL_REPLACEMENT
			// s.ensureMaterialsProfile(ctx, currentEquipmentMachineryName, materialsProfilesMap, maintenance[0].ID, equipmentNameToID[currentEquipmentMachineryName], request.Sector, lastIndexStr)
		}
		if strings.Contains(strings.ToLower(titleCell), types.LABEL_CONSUMABLE) {
			currentMaterialType = types.LABEL_CONSUMABLE
			// s.ensureMaterialsProfile(ctx, currentEquipmentMachineryName, materialsProfilesMap, maintenance[0].ID, equipmentNameToID[currentEquipmentMachineryName], request.Sector, lastIndexStr)
		}
		if currentMaterialType == types.LABEL_CONSUMABLE && indexCell == "-" {
			materialQuantity := 0.0
			if len(row) < 3 {
				continue // Skip rows that do not have enough columns
			}
			materialUnit := strings.ToLower(strings.TrimSpace(row[2]))
			if len(row) >= 4 {
				materialQuantity, err = strconv.ParseFloat(strings.TrimSpace(row[3]), 64)
				if err != nil {
					return err
				}
			}
			materialsProfilesMap[currentEquipmentMachineryName].Estimate.ConsumableSupplies[titleCell] = types.Material{
				Name:     titleCell,
				Unit:     materialUnit,
				Quantity: materialQuantity,
			}

		}
		if currentMaterialType == types.LABEL_REPLACEMENT && indexCell == "-" {
			if len(row) < 4 {
				continue // Skip rows that do not have enough columns
			}
			materialUnit := strings.ToLower(strings.TrimSpace(row[2]))
			materialQuantity, err := strconv.ParseFloat(strings.TrimSpace(row[3]), 64)
			if err != nil {
				return err
			}
			materialsProfilesMap[currentEquipmentMachineryName].Estimate.ReplacementMaterials[titleCell] = types.Material{
				Name:     titleCell,
				Unit:     materialUnit,
				Quantity: materialQuantity,
			}
		}

	}

	materialsProfileList := make([]*types.MaterialsProfile, 0, len(materialsProfilesMap))
	for _, materialsProfile := range materialsProfilesMap {
		materialsProfileList = append(materialsProfileList, materialsProfile)
	}

	_, err = s.materialsProfileRepo.SaveMany(ctx, materialsProfileList)
	if err != nil {
		return err
	}

	return nil

}

func (s *materialsProfileService) getMaintenanceIDs(ctx context.Context, request *types.MaterialsProfileFilterRequest) ([]string, error) {
	ids := make([]string, 0)
	if request.MaintenanceIDs != nil {
		ids = append(ids, request.MaintenanceIDs...)
	}
	if request.ProjectCode != "" || request.MaintenanceTier != "" || request.MaintenanceNumber != "" {
		filter := &types.MaintenanceFilter{
			ProjectCode:       request.ProjectCode,
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

func (s *materialsProfileService) PaginatedMaterialsProfiles(ctx context.Context, request *types.PaginatedRequest) ([]*types.MaterialsProfileResponse, int64, error) {
	filter := &types.MaterialsProfileFilter{}
	materialsProfiles, total, err := s.materialsProfileRepo.Paginate(ctx, filter, request.Page, request.Limit)
	if err != nil {
		return nil, 0, err
	}
	maintenanceInstanceIds := make([]string, 0)
	for _, profile := range materialsProfiles {
		maintenanceInstanceIds = append(maintenanceInstanceIds, profile.MaintenanceInstanceID)
	}
	maintenanceInstanceIds = utils.RemoveDuplicates(maintenanceInstanceIds)
	equipmentMachineryIds := make([]string, 0)
	for _, profile := range materialsProfiles {
		equipmentMachineryIds = append(equipmentMachineryIds, profile.EquipmentMachineryID)
	}
	equipmentMachineryIds = utils.RemoveDuplicates(equipmentMachineryIds)

	maintenances, err := s.maintenanceRepo.FindByIDs(ctx, maintenanceInstanceIds)
	if err != nil {
		return nil, 0, err
	}
	equipmentMachineries, err := s.equipmentMachineryRepo.FindByIDs(ctx, equipmentMachineryIds)
	if err != nil {
		return nil, 0, err
	}

	// Map materials profiles to response format
	var responses []*types.MaterialsProfileResponse

	for _, profile := range materialsProfiles {
		maintenance, ok := maintenances[profile.MaintenanceInstanceID]
		if !ok {
			continue
		}
		equipmentMachinery, ok := equipmentMachineries[profile.EquipmentMachineryID]
		if !ok {
			continue
		}
		response := &types.MaterialsProfileResponse{
			ID:                 profile.ID,
			Project:            maintenance.Project,
			ProjectCode:        maintenance.ProjectCode,
			MaintenanceTier:    maintenance.MaintenanceTier,
			MaintenanceNumber:  maintenance.MaintenanceNumber,
			Year:               maintenance.Year,
			Sector:             profile.Sector,
			EquipmentMachinery: equipmentMachinery.Name,
			IndexPath:          utils.IndexPathToString(profile.Index),
			Estimate:           profile.Estimate,
			Reality:            profile.Reality,
		}
		responses = append(responses, response)
	}

	return responses, total, nil
}

func (s *materialsProfileService) getEquipmentMachineryIDs(ctx context.Context, request *types.MaterialsProfileFilterRequest) ([]string, error) {
	ids := make([]string, 0)
	if request.EquipmentMachineryIDs != nil {
		ids = append(ids, request.EquipmentMachineryIDs...)
	}
	if request.EquipmentMachineryName != "" {
		filter := &types.EquipmentMachineryFilter{
			Name:   request.EquipmentMachineryName,
			Sector: request.Sector,
		}
		equipmentMachineries, err := s.equipmentMachineryRepo.Filter(ctx, filter)
		if err != nil {
			return nil, err
		}
		for _, equipmentMachinery := range equipmentMachineries {
			ids = append(ids, equipmentMachinery.ID)
		}
	}

	return ids, nil
}

func (s *materialsProfileService) ensureMaterialsProfile(ctx context.Context, equipmentName string, materialsProfilesMap map[string]*types.MaterialsProfile, maintenanceID, equipmentID, sector, indexPathStr string) {
	if _, exists := materialsProfilesMap[equipmentName]; !exists {
		materialsProfilesFromDb, _ := s.materialsProfileRepo.Filter(ctx, &types.MaterialsProfileFilter{
			MaintenanceInstanceIDs: []string{maintenanceID},
			EquipmentMachineryIDs:  []string{equipmentID},
			Sector:                 sector,
		})
		if len(materialsProfilesFromDb) > 0 {
			materialsProfilesMap[equipmentName] = materialsProfilesFromDb[0]
		} else {
			index, err := utils.StringToIndexPath(indexPathStr)
			if err != nil {
				index = 0
			}
			materialsProfilesMap[equipmentName] = &types.MaterialsProfile{
				MaintenanceInstanceID: maintenanceID,
				EquipmentMachineryID:  equipmentID,
				Sector:                sector,
				Index:                 index,
				Estimate: types.MaterialsForEquipment{
					ReplacementMaterials: make(map[string]types.Material),
					ConsumableSupplies:   make(map[string]types.Material),
				},
				Reality: types.MaterialsForEquipment{
					ReplacementMaterials: make(map[string]types.Material),
					ConsumableSupplies:   make(map[string]types.Material),
				},
			}
		}
	}
}
