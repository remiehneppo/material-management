package types

import "mime/multipart"

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type PaginatedRequest struct {
	Page  int64 `json:"page" binding:"required"`
	Limit int64 `json:"limit" binding:"required"`
}

type MaterialsProfileFilterRequest struct {
	Sector                 string   `json:"sector"`
	MaintenanceIDs         []string `json:"maintenance_ids"`
	Project                string   `json:"project"`
	MaintenanceTier        string   `json:"maintenance_tier"`
	MaintenanceNumber      string   `json:"maintenance_number"`
	EquipmentMachineryName string   `json:"equipment_machinery_name"`
	EquipmentMachineryIDs  []string `json:"equipment_machinery_ids"`
}

var (
	UPDATE_TYPE_NEW    = "new"
	UPDATE_TYPE_MODIFY = "modify"
)

type UpdateMaterialsRealityProfileRequest struct {
	UpdateType string `json:"update_type" binding:"required"`
}

type UpdateMaterialsEstimateProfileRequest struct {
}

type CreateMaterialRequestReq struct {
	Project               string                           `json:"project" binding:"required"`
	MaintenanceTier       string                           `json:"maintenance_tier" binding:"required"`
	MaintenanceNumber     string                           `json:"maintenance_number" binding:"required"`
	Sector                string                           `json:"sector" binding:"required"`
	Description           string                           `json:"description"`
	MaterialsForEquipment map[string]MaterialsForEquipment `json:"materials_for_equipment" binding:"required"`
}

type MaterialRequestUpdate struct {
	Sector                string                           `json:"sector" bson:"sector"`
	Description           string                           `json:"description" bson:"description"`
	MaterialsForEquipment map[string]MaterialsForEquipment `json:"materials_for_equipment" bson:"materials_for_equipment"`
}

type CreateMaintenanceRequest struct {
	Project           string `json:"project" binding:"required"`
	ProjectName       string `json:"project_code" binding:"required"`
	MaintenanceTier   string `json:"maintenance_tier" binding:"required"`
	MaintenanceNumber string `json:"maintenance_number" binding:"required"`
}

type UploadEstimateSheetRequest struct {
	Project           string                `json:"project" binding:"required"`
	MaintenanceTier   string                `json:"maintenance_tier" binding:"required"`
	MaintenanceNumber string                `json:"maintenance_number" binding:"required"`
	Sheet             *multipart.FileHeader `json:"sheet"`
	SheetName         string                `json:"sheet_name" binding:"required"`
	Sector            string                `json:"sector" binding:"required"`
}

type MaterialRequestExport struct {
	MaterialRequestID string `json:"material_request_id" binding:"required"`
}

type UpdateNumberOfRequestReq struct {
	MaterialRequestID string `json:"material_request_id" binding:"required"`
	NumOfRequest      int    `json:"num_of_request" binding:"required"`
}

type CreateEquipmentMachineryReq struct {
	Name   string `json:"name" binding:"required"`
	Sector string `json:"sector" binding:"required"`
	Order  int    `json:"order" binding:"required"`
}
