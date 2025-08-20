package service

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"baliance.com/gooxml/document"
	"github.com/remiehneppo/material-management/internal/repository"
	"github.com/remiehneppo/material-management/types"
	"github.com/remiehneppo/material-management/utils"
)

type MaterialsRequestService interface {
	CreateMaterialsRequest(ctx context.Context, request *types.CreateMaterialRequestReq) (string, error)
	GetMaterialsRequest(ctx context.Context, id string) (*types.MaterialRequestResponse, error)
	FilterMaterialsRequests(ctx context.Context, req *types.MaterialRequestFilter) ([]*types.MaterialRequestResponse, error)
	UpdateMaterialsRequest(ctx context.Context, id string, request *types.MaterialRequestUpdate) error
	DeleteMaterialsRequest(ctx context.Context, id string) error
	UpdateNumberOfRequest(ctx context.Context, req types.UpdateNumberOfRequestReq) error
	// create a docx file and stream to user to download and print
	ExportMaterialsRequest(ctx context.Context, req *types.MaterialRequestExport) (*os.File, error)
}

type materialsRequestService struct {
	materialsRequestRepo   repository.MaterialsRequestRepository
	materialsProfileRepo   repository.MaterialsProfileRepository
	maintenanceRepo        repository.MaintenanceRepository
	equipmentMachineryRepo repository.EquipmentMachineryRepo
	templateRequestPath    string
}

func NewMaterialsRequestService(
	materialsRequestRepo repository.MaterialsRequestRepository,
	materialsProfileRepo repository.MaterialsProfileRepository,
	maintenanceRepo repository.MaintenanceRepository,
	equipmentMachineryRepo repository.EquipmentMachineryRepo,
	templateRequestPath string,
) MaterialsRequestService {
	return &materialsRequestService{
		materialsRequestRepo:   materialsRequestRepo,
		materialsProfileRepo:   materialsProfileRepo,
		maintenanceRepo:        maintenanceRepo,
		equipmentMachineryRepo: equipmentMachineryRepo,
		templateRequestPath:    templateRequestPath,
	}
}

func (s *materialsRequestService) CreateMaterialsRequest(ctx context.Context, request *types.CreateMaterialRequestReq) (string, error) {
	// Validate sector
	if !utils.Contains(types.SECTOR_LIST, request.Sector) {
		return "", types.ErrInvalidSector
	}

	// Validate maintenance tier
	if !utils.Contains(types.MAINTENANCE_TIER_LIST, request.MaintenanceTier) {
		return "", types.ErrInvalidMaintenanceTier
	}

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

func (s *materialsRequestService) FilterMaterialsRequests(ctx context.Context, req *types.MaterialRequestFilter) ([]*types.MaterialRequestResponse, error) {
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

func (s *materialsRequestService) UpdateNumberOfRequest(ctx context.Context, req types.UpdateNumberOfRequestReq) error {
	materialsRequest, err := s.materialsRequestRepo.FindByID(ctx, req.MaterialRequestID)
	if err != nil {
		return err
	}

	if materialsRequest == nil {
		return types.ErrMaterialRequestNotFound
	}

	materialsRequest.NumOfRequest = req.NumOfRequest

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
		// Only process materials for the equipment that matches this profile
		if materials, exists := materialsRequest.MaterialsForEquipment[profile.EquipmentMachineryID]; exists {
			// Initialize Reality if it's nil
			if profile.Reality.ConsumableSupplies == nil {
				profile.Reality.ConsumableSupplies = make(map[string]types.Material)
			}
			if profile.Reality.ReplacementMaterials == nil {
				profile.Reality.ReplacementMaterials = make(map[string]types.Material)
			}

			// Update consumable supplies
			for _, consumableSupplies := range materials.ConsumableSupplies {
				if existing, ok := profile.Reality.ConsumableSupplies[consumableSupplies.Name]; ok {
					existing.Quantity += consumableSupplies.Quantity
					profile.Reality.ConsumableSupplies[consumableSupplies.Name] = existing
				} else {
					profile.Reality.ConsumableSupplies[consumableSupplies.Name] = consumableSupplies
				}
			}

			// Update replacement materials
			for _, replacementMaterial := range materials.ReplacementMaterials {
				if existing, ok := profile.Reality.ReplacementMaterials[replacementMaterial.Name]; ok {
					existing.Quantity += replacementMaterial.Quantity
					profile.Reality.ReplacementMaterials[replacementMaterial.Name] = existing
				} else {
					profile.Reality.ReplacementMaterials[replacementMaterial.Name] = replacementMaterial
				}
			}

			err = s.materialsProfileRepo.UpdateRealityMaterials(ctx, profile.ID, profile.Reality)
			if err != nil {
				return err
			}
		}
	}

	return s.materialsRequestRepo.Update(ctx, req.MaterialRequestID, materialsRequest)
}

func (s *materialsRequestService) ExportMaterialsRequest(ctx context.Context, req *types.MaterialRequestExport) (*os.File, error) {
	// create a docx file and stream to user to download and print
	doc, err := document.Open(s.templateRequestPath)
	if err != nil {
		return nil, err
	}

	materialRequest, err := s.materialsRequestRepo.FindByID(ctx, req.MaterialRequestID)
	if err != nil {
		return nil, err
	}

	maintenance, err := s.maintenanceRepo.FindByID(ctx, materialRequest.MaintenanceInstanceID)
	if err != nil {
		return nil, err
	}
	eqIDs := make([]string, 0, len(materialRequest.MaterialsForEquipment))
	for eqID := range materialRequest.MaterialsForEquipment {
		eqIDs = append(eqIDs, eqID)
	}
	eqs, err := s.equipmentMachineryRepo.FindByIDs(ctx, eqIDs)
	if err != nil {
		return nil, err
	}
	if len(eqs) != len(eqIDs) {
		return nil, types.ErrSomeEquipmentMachineryNotFound
	}

	s.replacePlaceholderInDoc(doc, maintenance, materialRequest)

	consumableMaterialsMap := make(map[string]types.Material)

	tables := doc.Tables()
	materialTable := tables[1]
	currentEquipmentIndex := 1
	currentTableIndex := 1

	for eqID, eqMaterials := range materialRequest.MaterialsForEquipment {
		newRow := materialTable.InsertRowBefore(materialTable.Rows()[len(materialTable.Rows())-1])
		indexRun := newRow.AddCell().AddParagraph().AddRun()
		indexRun.Properties().SetBold(true)
		indexRun.AddText(utils.IntToRoman(currentEquipmentIndex))
		titleRun := newRow.AddCell().AddParagraph().AddRun()
		titleRun.Properties().SetBold(true)
		titleRun.AddText(eqs[eqID].Name)
		newRow.AddCell().AddParagraph().AddRun().AddText("")
		newRow.AddCell().AddParagraph().AddRun().AddText("")
		newRow.AddCell().AddParagraph().AddRun().AddText("")
		newRow.AddCell().AddParagraph().AddRun().AddText("")
		newRow.AddCell().AddParagraph().AddRun().AddText("")

		for _, consumable := range eqMaterials.ConsumableSupplies {
			if existing, ok := consumableMaterialsMap[consumable.Name]; ok {
				existing.Quantity += consumable.Quantity
				consumableMaterialsMap[consumable.Name] = existing
			} else {
				consumableMaterialsMap[consumable.Name] = consumable
			}
		}

		for _, replacement := range eqMaterials.ReplacementMaterials {
			newRow := materialTable.InsertRowBefore(materialTable.Rows()[len(materialTable.Rows())-1])
			newRow.AddCell().AddParagraph().AddRun().AddText(fmt.Sprintf("%d", currentTableIndex))
			newRow.AddCell().AddParagraph().AddRun().AddText(replacement.Name)
			newRow.AddCell().AddParagraph().AddRun().AddText(replacement.Unit)
			newRow.AddCell().AddParagraph().AddRun().AddText("")
			newRow.AddCell().AddParagraph().AddRun().AddText(fmt.Sprintf("%.2f", replacement.Quantity))
			newRow.AddCell().AddParagraph().AddRun().AddText("")
			newRow.AddCell().AddParagraph().AddRun().AddText("")
			currentTableIndex += 1
		}

		currentEquipmentIndex += 1
	}

	newRow := materialTable.InsertRowBefore(materialTable.Rows()[len(materialTable.Rows())-1])
	indexRun := newRow.AddCell().AddParagraph().AddRun()
	indexRun.Properties().SetBold(true)
	indexRun.AddText(utils.IntToRoman(currentEquipmentIndex))
	titleRun := newRow.AddCell().AddParagraph().AddRun()
	titleRun.Properties().SetBold(true)
	titleRun.AddText(strings.ToUpper(types.LABEL_CONSUMABLE))
	newRow.AddCell().AddParagraph().AddRun().AddText("")
	newRow.AddCell().AddParagraph().AddRun().AddText("")
	newRow.AddCell().AddParagraph().AddRun().AddText("")
	newRow.AddCell().AddParagraph().AddRun().AddText("")
	newRow.AddCell().AddParagraph().AddRun().AddText("")

	for consumableName, consumable := range consumableMaterialsMap {
		newRow := materialTable.InsertRowBefore(materialTable.Rows()[len(materialTable.Rows())-1])
		newRow.AddCell().AddParagraph().AddRun().AddText(fmt.Sprintf("%d", currentTableIndex))
		newRow.AddCell().AddParagraph().AddRun().AddText(consumableName)
		newRow.AddCell().AddParagraph().AddRun().AddText(consumable.Unit)
		newRow.AddCell().AddParagraph().AddRun().AddText("")
		newRow.AddCell().AddParagraph().AddRun().AddText(fmt.Sprintf("%.2f", consumable.Quantity))
		newRow.AddCell().AddParagraph().AddRun().AddText("")
		newRow.AddCell().AddParagraph().AddRun().AddText("")
		currentTableIndex += 1
	}

	tempDir := os.TempDir()
	saveDir := path.Join(tempDir, "material_request")
	// create dir if not exist
	if err := os.MkdirAll(saveDir, os.ModePerm); err != nil {
		return nil, err
	}
	fileName := path.Join(
		saveDir,
		fmt.Sprintf(
			"%s%s.docx",
			types.MATERIALS_REQUEST_PREFIX,
			time.Now().Local().Format("2006-01-02"),
		),
	)

	doc.SaveToFile(fileName)
	return os.Open(fileName)
}

func (s *materialsRequestService) replacePlaceholderInDoc(doc *document.Document, maintenance *types.Maintenance, materialRequest *types.MaterialRequest) {
	replacements := map[string]string{
		"{project}":     maintenance.ProjectName,
		"{workshop}":    fmt.Sprintf("X. %s", materialRequest.Sector),
		"{team}":        ".....",
		"{description}": materialRequest.Description,
		"{year}":        time.Now().Format("2006"),
	}

	// Replace in paragraphs
	paras := doc.Paragraphs()
	for _, para := range paras {
		for _, run := range para.Runs() {
			s.replaceTextInRun(run, replacements)
		}
	}

	// Replace in tables
	tables := doc.Tables()

	// Handle special number cell in first table
	if len(tables) > 0 {
		numRqCell := tables[0].Rows()[0].Cells()[2]
		for _, para := range numRqCell.Paragraphs() {
			for _, run := range para.Runs() {
				run.ClearContent()
			}
		}
		numRqCell.Paragraphs()[0].Runs()[0].AddText(
			fmt.Sprintf(
				"Số:   /%s/%s/%s",
				maintenance.Project,
				types.ShortSectorList[materialRequest.Sector],
				time.Now().Format("06"),
			),
		)
	}

	// Replace in all table cells
	for _, table := range tables {
		for _, row := range table.Rows() {
			for _, cell := range row.Cells() {
				for _, para := range cell.Paragraphs() {
					for _, run := range para.Runs() {
						s.replaceTextInRun(run, replacements)
					}
				}
			}
		}
	}
}

func (s *materialsRequestService) replaceTextInRun(run document.Run, replacements map[string]string) {
	text := run.Text()
	modified := false

	for placeholder, replacement := range replacements {
		if strings.Contains(text, placeholder) {
			text = strings.ReplaceAll(text, placeholder, replacement)
			modified = true
		}
	}

	if modified {
		run.ClearContent()
		run.AddText(text)
	}
}
