package model

type PaginationQuery struct {
	Page   int `json:"page"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

type PaginatedMeta struct {
	CurrentPage int64 `json:"current_page"`
	Total       int64 `json:"total"`
	PerPage     int64 `json:"per_page"`
	LastPage    int   `json:"last_page"`
	HasNextPage bool  `json:"has_next_page"`
	HasPrevPage bool  `json:"has_prev_page"`
}
