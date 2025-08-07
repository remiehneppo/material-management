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
