package types

import "errors"

var (
	ErrInvalidUser              = errors.New("invalid user")
	ErrInvalidTask              = errors.New("invalid task")
	ErrInvalidCredentials       = errors.New("invalid credentials")
	ErrTaskNotCreatorOrAssignee = errors.New("task not creator or assignee")
	ErrUserNotFound             = errors.New("user not found")
	ErrUnauthorized             = errors.New("unauthorized")
)

var (
	ErrInvalidUpdateType              = errors.New("invalid update type")
	ErrMaintenanceNotFound            = errors.New("maintenance not found")
	ErrMaterialRequestNotFound        = errors.New("material request not found")
	ErrUpdateAfterGotNumOfRequest     = errors.New("cannot update after getting number of request")
	ErrSomeEquipmentMachineryNotFound = errors.New("some equipment machinery not found")
	ErrDuplicateMaintenance           = errors.New("duplicate maintenance")
	ErrInvalidSector                  = errors.New("invalid sector")
	ErrInvalidMaintenanceTier         = errors.New("invalid maintenance tier")
	ErrNotImplemented                 = errors.New("feature not implemented")
)
