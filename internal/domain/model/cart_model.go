package model

import (
	"cakestore/internal/domain/entity"
	"database/sql"
	"time"
)

type CartModel struct {
	ID         int       `json:"id" validate:"required"`
	CustomerID int       `json:"customer_id"`
	CakeID     int       `json:"cake_id"`
	Quantity   int       `json:"quantity"`
	Price      float64   `json:"price"`
	Subtotal   float64   `json:"subtotal"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type AddCart struct {
	CakeID   int `json:"cake_id" validate:"required"`
	Quantity int `json:"quantity" validate:"required,min=1"`
}

func ToCartEntity(m *CartModel) *entity.Cart {
	return &entity.Cart{
		ID:         m.ID,
		CustomerID: m.CustomerID,
		CakeID:     m.CakeID,
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
		CakeID:     e.CakeID,
		Quantity:   e.Quantity,
		Price:      e.Price,
		Subtotal:   e.Subtotal,
		CreatedAt:  e.CreatedAt,
		UpdatedAt:  e.UpdatedAt,
	}
}