package model

import (
	"cakestore/internal/domain/entity"
	"time"
)

type CreateTableRequest struct {
	TableNumber int `json:"table_number" validate:"required"`
	Capacity    int `json:"capacity" validate:"required"`
}

type UpdateTableRequest struct {
	TableNumber int  `json:"table_number"`
	Capacity    int  `json:"capacity"`
	IsAvailable bool `json:"is_available"`
}

type TableResponse struct {
	ID          uint      `json:"id"`
	TableNumber int       `json:"table_number"`
	Capacity    int       `json:"capacity"`
	IsAvailable bool      `json:"is_available"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type TableQueryParams struct {
	Page        int64 `json:"page"`
	Limit       int64 `json:"limit"`
	Capacity    int   `json:"capacity"`
	IsAvailable *bool `json:"is_available"`
}

func ToTableResponse(table *entity.Table) *TableResponse {
	return &TableResponse{
		ID:          uint(table.ID),
		TableNumber: table.TableNumber,
		Capacity:    table.Capacity,
		IsAvailable: table.IsAvailable,
		CreatedAt:   table.CreatedAt,
		UpdatedAt:   table.UpdatedAt,
	}
}
