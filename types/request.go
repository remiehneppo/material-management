package types

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
