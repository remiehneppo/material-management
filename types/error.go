package types

import "errors"

var (
	ErrUsernameInvalid              = errors.New("tên đăng nhập không hợp lệ")
	ErrPasswordInvalid              = errors.New("mật khẩu không hợp lệ")
	ErrInvalidUser                  = errors.New("người dùng không hợp lệ")
	ErrUsernameOrPasswordNotCorrect = errors.New("tên đăng nhập hoặc mật khẩu không đúng")
	ErrInvalidTask                  = errors.New("invalid task")
	ErrInvalidCredentials           = errors.New("thông tin đăng nhập không hợp lệ")
	ErrTaskNotCreatorOrAssignee     = errors.New("task not creator or assignee")
	ErrUserNotFound                 = errors.New("không tìm thấy người dùng")
	ErrUnauthorized                 = errors.New("bạn chưa đăng nhập hoặc không có quyền thực hiện thao tác này")
	ErrPasswordIncorrect            = errors.New("mật khẩu hiện tại không đúng")
)

var (
	ErrInvalidUpdateType                   = errors.New("loại cập nhật không hợp lệ")
	ErrMaintenanceNotFound                 = errors.New("không tìm thấy dự án sửa chữa")
	ErrMaterialRequestNotFound             = errors.New("không tìm thấy yêu cầu vật tư")
	ErrUpdateAfterGotNumOfRequest          = errors.New("không thể sửa yêu cầu vật tư đã được cấp số")
	ErrSomeEquipmentMachineryNotFound      = errors.New("không tìm thấy một hoặc nhiều thiết bị")
	ErrDuplicateMaintenance                = errors.New("dự án sửa chữa này đã tồn tại")
	ErrInvalidSector                       = errors.New("ngành không hợp lệ")
	ErrInvalidMaintenanceTier              = errors.New("cấp sửa chữa không hợp lệ")
	ErrNotImplemented                      = errors.New("chức năng này chưa được hỗ trợ")
	ErrSomeMaterialsProfileNotFound        = errors.New("không tìm thấy một hoặc nhiều hồ sơ vật tư")
	ErrMaterialsProfileSectorMismatch      = errors.New("ngành của hồ sơ vật tư không khớp với yêu cầu")
	ErrMaterialsProfileMaintenanceMismatch = errors.New("dự án của hồ sơ vật tư không khớp với yêu cầu")
	ErrMaterialRequestNotDraft             = errors.New("chỉ có thể thay đổi yêu cầu vật tư chưa ban hành")
	ErrMaterialRequestNotOwner             = errors.New("yêu cầu vật tư thuộc về người lập khác")
	ErrMaterialNotInProfile                = errors.New("vật tư được yêu cầu chưa có trong hồ sơ vật tư")
)
