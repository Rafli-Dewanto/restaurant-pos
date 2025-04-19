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

type PaginationResponse struct {
	Data       interface{} `json:"data"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalPages int         `json:"total_pages"`
}

func ToPaginatedMeta(PaginationResponse *PaginationResponse) *PaginatedMeta {
	return &PaginatedMeta{
		CurrentPage: int64(PaginationResponse.Page),
		Total:       PaginationResponse.Total,
		PerPage:     int64(PaginationResponse.PageSize),
		LastPage:    PaginationResponse.TotalPages,
		HasNextPage: PaginationResponse.Page < PaginationResponse.TotalPages,
		HasPrevPage: PaginationResponse.Page > 1,
	}
}
