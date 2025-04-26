package model

type PaginationQuery struct {
	Page   int64 `json:"page"`
	Limit  int64 `json:"limit"`
	Offset int64 `json:"offset"`
}

type PaginatedMeta struct {
	CurrentPage int64 `json:"current_page"`
	Total       int64 `json:"total"`
	PerPage     int64 `json:"per_page"`
	LastPage    int64 `json:"last_page"`
	HasNextPage bool  `json:"has_next_page"`
	HasPrevPage bool  `json:"has_prev_page"`
}

type PaginationResponse[T any] struct {
	Data       T     `json:"data"`
	Total      int64 `json:"total"`
	Page       int64 `json:"page"`
	PageSize   int64 `json:"page_size"`
	TotalPages int64 `json:"total_pages"`
}

func ToPaginatedMeta[T any](PaginationResponse *PaginationResponse[T]) *PaginatedMeta {
	return &PaginatedMeta{
		CurrentPage: int64(PaginationResponse.Page),
		Total:       PaginationResponse.Total,
		PerPage:     int64(PaginationResponse.PageSize),
		LastPage:    PaginationResponse.TotalPages,
		HasNextPage: PaginationResponse.Page < PaginationResponse.TotalPages,
		HasPrevPage: PaginationResponse.Page > 1,
	}
}
