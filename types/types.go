package types

const (
	USER_ROLE_ADMIN = "admin"
)
const (
	USER_WORKSPACE_ROLE_EXECUTIVE = "executive"
	USER_WORKSPACE_ROLE_HEAD      = "head"
	USER_WORKSPACE_ROLE_DHEAD     = "dhead"
	USER_WORKSPACE_ROLE_ASSISTANT = "assistant"
	USER_WORKSPACE_ROLE_STAFF     = "staff"
)

const (
	USER_MANAGEMENT_LEVEL_EXECUTIVE = 5
	USER_MANAGEMENT_LEVEL_HEAD      = 4
	USER_MANAGEMENT_LEVEL_DHEAD     = 3
	USER_MANAGEMENT_LEVEL_ASSISTANT = 2
	USER_MANAGEMENT_LEVEL_STAFF     = 2
)

var MAPPING_ROLE_TO_MANAGEMENT_LEVEL map[string]int = map[string]int{
	USER_WORKSPACE_ROLE_EXECUTIVE: USER_MANAGEMENT_LEVEL_EXECUTIVE,
	USER_WORKSPACE_ROLE_HEAD:      USER_MANAGEMENT_LEVEL_HEAD,
	USER_WORKSPACE_ROLE_DHEAD:     USER_MANAGEMENT_LEVEL_DHEAD,
	USER_WORKSPACE_ROLE_ASSISTANT: USER_MANAGEMENT_LEVEL_ASSISTANT,
	USER_WORKSPACE_ROLE_STAFF:     USER_MANAGEMENT_LEVEL_STAFF,
}

const (
	DepartmentTechnical      = "DepartmentTechnical"
	DepartmentProductionPlan = "DepartmentProductionPlan"
	DepartmentQuality        = "DepartmentQuality"
	DepartmentMaterial       = "DepartmentMaterial"
)

type Admin struct {
	ID       string `json:"id" bson:"_id,omitempty"`
	Username string `json:"username" bson:"username"`
	Password string `json:"password" bson:"password"`
	Role     string `json:"role" bson:"role"`
}

type User struct {
	ID              string `json:"id" bson:"_id,omitempty"`
	Username        string `json:"username" bson:"username"`
	Password        string `json:"password" bson:"password"`
	FullName        string `json:"full_name" bson:"full_name"`
	ManagementLevel int    `json:"management_level" bson:"management_level"`
	WorkspaceRole   string `json:"workspace_role" bson:"workspace_role"`
	Workspace       string `json:"workspace" bson:"workspace"`
	CreateAt        int64  `json:"created_at" bson:"created_at"`
	UpdateAt        int64  `json:"updated_at" bson:"updated_at"`
}

type Workspace struct {
	ID   string `json:"id" bson:"_id,omitempty"`
	Name string `json:"name" bson:"name"`
}

// Material Management Types

var (
	SECTOR_MECHANICAL           = "Mechanical"
	SECTOR_WEAPONS              = "Weapons"
	SECTOR_HULL_TANK            = "HullTank"
	SECTOR_ELECTRONICS          = "Electronics"
	SECTOR_PROPULSION           = "Propulsion"
	SECTOR_VALVE_PIPE           = "ValvePipe"
	SECTOR_ELECTRONICS_TACTICAL = "ElectronicsTactical"
	SECTOR_DECORATIVE           = "Decorative"
)

var (
	SECTOR_LIST = []string{
		SECTOR_MECHANICAL,
		SECTOR_WEAPONS,
		SECTOR_HULL_TANK,
		SECTOR_ELECTRONICS,
		SECTOR_PROPULSION,
		SECTOR_VALVE_PIPE,
		SECTOR_ELECTRONICS_TACTICAL,
		SECTOR_DECORATIVE,
	}
)

type EquipmentMachinery struct {
	ID     string `json:"id" bson:"_id,omitempty"`
	Name   string `json:"name" bson:"name"`
	Sector string `json:"sector" bson:"sector"`
}

type Material struct {
	Name     string `json:"name" bson:"name"`
	Unit     string `json:"unit" bson:"unit"`
	Quantity int    `json:"quantity" bson:"quantity"`
}

type MaterialsForEquipment struct {
	ConsumableSupplies   []Material `json:"consumable_supplies" bson:"consumable_supplies"`
	ReplacementMaterials []Material `json:"replacement_materials" bson:"replacement_materials"`
}
type MaintainInstance struct {
	ID                   string `json:"id" bson:"_id,omitempty"`
	Vehicle              string `json:"vehicle" bson:"vehicle"`
	MaintenanceTier      string `json:"maintenance_tier" bson:"maintenance_tier"`
	MaintenanceNumber    string `json:"maintenance_number" bson:"maintenance_number"`
	Year                 int    `json:"year" bson:"year"`
	EquipmentMachineryID string `json:"equipment_machinery_id" bson:"equipment_machinery_id"`
}

type MaterialsProfile struct {
	ID                   string                `json:"id" bson:"_id,omitempty"`
	MaintainInstanceID   string                `json:"maintain_instance_id" bson:"maintain_instance_id"`
	EquipmentMachineryID string                `json:"equipment_machinery_id" bson:"equipment_machinery_id"`
	Sector               string                `json:"sector" bson:"sector"`
	Estimate             MaterialsForEquipment `json:"estimate" bson:"estimate"`
	Reality              MaterialsForEquipment `json:"reality" bson:"reality"`
}

type MaterialRequest struct {
	ID                    string                           `json:"id" bson:"_id,omitempty"`
	MaintainInstanceID    string                           `json:"maintain_instance_id" bson:"maintain_instance_id"`
	NumOfRequest          int                              `json:"num_of_request" bson:"num_of_request"`
	EquipmentMachineryID  string                           `json:"equipment_machinery_id" bson:"equipment_machinery_id"`
	Sector                string                           `json:"sector" bson:"sector"`
	MaterialsForEquipment map[string]MaterialsForEquipment `json:"materials_for_equipment" bson:"materials_for_equipment"`
	RequestedBy           string                           `json:"requested_by" bson:"requested_by"`
	RequestedAt           int64                            `json:"requested_at" bson:"requested_at"`
}

type MaterialsProfileFilter struct {
	MaintainInstanceID   string `json:"maintain_instance_id" bson:"maintain_instance_id"`
	EquipmentMachineryID string `json:"equipment_machinery_id" bson:"equipment_machinery_id"`
	Sector               string `json:"sector" bson:"sector"`
}

type MaterialRequestFilter struct {
	MaintainInstanceID   string `json:"maintain_instance_id" bson:"maintain_instance_id"`
	EquipmentMachineryID string `json:"equipment_machinery_id" bson:"equipment_machinery_id"`
	NumOfRequest         int    `json:"num_of_request" bson:"num_of_request"`
	Sector               string `json:"sector" bson:"sector"`
	RequestedBy          string `json:"requested_by" bson:"requested_by"`
	RequestedAtStart     int64  `json:"requested_at_start" bson:"requested_at_start"`
	RequestedAtEnd       int64  `json:"requested_at_end" bson:"requested_at_end"`
}
