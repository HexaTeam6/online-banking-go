package model

// Pagination holds pagination parameters for paginated requests.
type Pagination struct {
	Page     int `json:"page" validate:"min=1"`
	PageSize int `json:"page_size" validate:"min=1,max=100"`
}
