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

type PaginateMessagesRequest struct {
	ChatId string `json:"chat_id" binding:"required"`
	Page   int64  `json:"page" binding:"required"`
	Limit  int64  `json:"limit" binding:"required"`
}
