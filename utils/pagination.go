package utils

import (
	"cakestore/internal/domain/model"
	"math"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

func GetPaginationFromRequest(c *fiber.Ctx) *model.PaginationQuery {
	page := c.Query("page", "1")
	perPage := c.Query("per_page", "10")

	pageInt, err := strconv.ParseInt(page, 10, 64)
	if err != nil {
		pageInt = 1
	}

	perPageInt, err := strconv.ParseInt(perPage, 10, 64)
	if err != nil {
		perPageInt = 10
	}

	return &model.PaginationQuery{
		Page:  pageInt,
		Limit: perPageInt,
	}
}

func CreatePaginationMeta(page, perPage int64, total int64) *model.PaginatedMeta {
	lastPage := int64(math.Ceil(float64(total) / float64(perPage)))
	return &model.PaginatedMeta{
		Total:       total,
		CurrentPage: int64(page),
		PerPage:     int64(perPage),
		LastPage:    lastPage,
		HasNextPage: int64(page) < int64(lastPage),
		HasPrevPage: page > 1,
	}
}
