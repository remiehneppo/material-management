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
