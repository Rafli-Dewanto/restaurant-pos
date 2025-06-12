package model

import (
	"time"
)

type CreateInventoryRequest struct {
	Name         string  `json:"name" validate:"required"`
	Quantity     float64 `json:"quantity" validate:"required,min=0"`
	Unit         string  `json:"unit" validate:"required"`
	MinimumStock float64 `json:"minimum_stock" validate:"required,min=0"`
	ReorderPoint float64 `json:"reorder_point" validate:"required,min=0"`
	UnitPrice    float64 `json:"unit_price" validate:"required,min=0"`
}

type UpdateInventoryRequest struct {
	Name         string  `json:"name" validate:"omitempty"`
	Quantity     float64 `json:"quantity" validate:"omitempty,min=0"`
	Unit         string  `json:"unit" validate:"omitempty"`
	MinimumStock float64 `json:"minimum_stock" validate:"omitempty,min=0"`
	ReorderPoint float64 `json:"reorder_point" validate:"omitempty,min=0"`
	UnitPrice    float64 `json:"unit_price" validate:"omitempty,min=0"`
}

type InventoryResponse struct {
	ID              uint      `json:"id"`
	Name            string    `json:"name"`
	Quantity        float64   `json:"quantity"`
	Unit            string    `json:"unit"`
	MinimumStock    float64   `json:"minimum_stock"`
	ReorderPoint    float64   `json:"reorder_point"`
	UnitPrice       float64   `json:"unit_price"`
	LastRestockDate time.Time `json:"last_restock_date"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type InventoryQueryParams struct {
	Page   int64  `json:"page" validate:"required,min=1"`
	Limit  int64  `json:"limit" validate:"required,min=1"`
	Search string `json:"search"`
}
