package utils

import (
	"cakestore/internal/domain/model"
	"math"
)

func CreatePaginationMeta(page, perPage int, total int64) *model.PaginatedMeta {
	lastPage := int(math.Ceil(float64(total) / float64(perPage)))
	return &model.PaginatedMeta{
		Total:       total,
		CurrentPage: int64(page),
		PerPage:     int64(perPage),
		LastPage:    lastPage,
		HasNextPage: int64(page) < int64(lastPage),
		HasPrevPage: page > 1,
	}
}
