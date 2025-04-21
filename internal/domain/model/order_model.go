package model

import (
	"cakestore/internal/domain/entity"
	"time"
)

type OrderItemRequest struct {
	CakeID   int     `json:"cake_id" validate:"required"`
	Title    string  `json:"title" validate:"required"`
	Quantity int     `json:"quantity" validate:"required,min=1"`
	Price    float64 `json:"price" validate:"required,min=0"`
}

type CreateOrderRequest struct {
	Items []OrderItemRequest `json:"items" validate:"required,min=1,dive"`
}

type OrderItemResponse struct {
	ID       int       `json:"id"`
	Cake     CakeModel `json:"cake"`
	Quantity int       `json:"quantity"`
	Price    float64   `json:"price"`
}

type OrderResponse struct {
	ID         int                 `json:"id"`
	Customer   CustomerResponse    `json:"customer"`
	Status     string              `json:"status"`
	TotalPrice float64             `json:"total_price"`
	Address    string              `json:"delivery_address"`
	Items      []OrderItemResponse `json:"items"`
	CreatedAt  string              `json:"created_at"`
	UpdatedAt  string              `json:"updated_at"`
}

type UpdateOrderStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=pending paid preparing delivered cancelled"`
}

func OrderToResponse(order *entity.Order) *OrderResponse {
	itemResponses := make([]OrderItemResponse, len(order.Items))
	for i, item := range order.Items {
		itemResponses[i] = OrderItemResponse{
			ID:       item.ID,
			Cake:     *CakeToResponse(&item.Cake),
			Quantity: item.Quantity,
			Price:    item.Price,
		}
	}

	return &OrderResponse{
		ID: order.ID,
		Customer: CustomerResponse{
			ID:      order.Customer.ID,
			Name:    order.Customer.Name,
			Email:   order.Customer.Email,
			Address: order.Customer.Address,
		},
		Status:     string(order.Status),
		TotalPrice: order.TotalPrice,
		Address:    order.Address,
		Items:      itemResponses,
		CreatedAt:  order.CreatedAt.Format(time.RFC3339),
		UpdatedAt:  order.UpdatedAt.Format(time.RFC3339),
	}
}
