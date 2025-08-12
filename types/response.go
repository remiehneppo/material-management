package types

type Response struct {
	Status  bool        `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// PaginatedData holds pagination metadata and the actual items
type PaginatedData struct {
	Total int64       `json:"total"`
	Limit int64       `json:"limit"`
	Page  int64       `json:"page"`
	Items interface{} `json:"items"`
}

// PaginatedResponse for paginated API responses
type PaginatedResponse struct {
	Status  bool          `json:"status"`
	Message string        `json:"message"`
	Data    PaginatedData `json:"data"`
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type MaterialsForEquipmentResponse struct {
	EquipmentMachineryName string              `json:"equipment_machinery_name"`
	ConsumableSupplies     map[string]Material `json:"consumable_supplies"`
	ReplacementMaterials   map[string]Material `json:"replacement_materials"`
}

type MaterialRequestResponse struct {
	ID                    string                                   `json:"id"`
	Project               string                                   `json:"project"`
	MaintenanceTier       string                                   `json:"maintenance_tier"`
	MaintenanceNumber     string                                   `json:"maintenance_number"`
	Year                  int                                      `json:"year"`
	Sector                string                                   `json:"sector"`
	Description           string                                   `json:"description"`
	MaterialsForEquipment map[string]MaterialsForEquipmentResponse `json:"materials_for_equipment"`
	RequestedBy           string                                   `json:"requested_by"`
	RequestedAt           int64                                    `json:"requested_at"`
	NumOfRequest          int                                      `json:"num_of_request"`
}
