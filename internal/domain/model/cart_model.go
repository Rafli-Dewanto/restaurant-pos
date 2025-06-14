package model

import (
	"cakestore/internal/domain/entity"
	"database/sql"
	"time"
)

type CartModel struct {
	ID         int64     `json:"id" validate:"required"`
	CustomerID int64     `json:"customer_id"`
	MenuID     int64     `json:"menu_id"`
	Quantity   int64     `json:"quantity"`
	Price      float64   `json:"price"`
	Subtotal   float64   `json:"subtotal"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type AddCart struct {
	MenuID   int64 `json:"menu_id" validate:"required"`
	Quantity int64 `json:"quantity" validate:"required,min=1"`
}

func ToCartEntity(m *CartModel) *entity.Cart {
	return &entity.Cart{
		ID:         m.ID,
		CustomerID: m.CustomerID,
		MenuID:     m.MenuID,
		Quantity:   m.Quantity,
		Price:      m.Price,
		Subtotal:   m.Subtotal,
		CreatedAt:  m.CreatedAt,
		UpdatedAt:  m.UpdatedAt,
		DeletedAt:  sql.NullTime{},
	}
}

func ToCartModel(e *entity.Cart) *CartModel {
	return &CartModel{
		ID:         e.ID,
		CustomerID: e.CustomerID,
		MenuID:     e.MenuID,
		Quantity:   e.Quantity,
		Price:      e.Price,
		Subtotal:   e.Subtotal,
		CreatedAt:  e.CreatedAt,
		UpdatedAt:  e.UpdatedAt,
	}
}
