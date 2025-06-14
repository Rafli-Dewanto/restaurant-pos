package model

import (
	"cakestore/internal/domain/entity"
	"time"
)

type OrderItemRequest struct {
	MenuID   int64   `json:"menu_id" validate:"required"`
	Title    string  `json:"title" validate:"required"`
	Quantity int64   `json:"quantity" validate:"required,min=1"`
	Price    float64 `json:"price" validate:"required,min=0"`
}

type UpdateFoodStatusRequest struct {
	FoodStatus string `json:"food_status" validate:"required,oneof=pending cooking ready delivered cancelled"`
}

type CreateOrderRequest struct {
	Items []OrderItemRequest `json:"items" validate:"required,min=1,dive"`
}

type OrderItemResponse struct {
	ID       int64     `json:"id"`
	Menu     MenuModel `json:"menu"`
	Quantity int64     `json:"quantity"`
	Price    float64   `json:"price"`
}

type OrderResponse struct {
	ID         int64               `json:"id"`
	Customer   CustomerResponse    `json:"customer"`
	Status     string              `json:"status"`
	TotalPrice float64             `json:"total_price"`
	Address    string              `json:"delivery_address"`
	FoodStatus string              `json:"food_status"`
	Items      []OrderItemResponse `json:"items"`
	CreatedAt  string              `json:"created_at"`
	UpdatedAt  string              `json:"updated_at"`
}

type UpdateOrderStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=pending paid preparing delivered cancelled"`
}

func ToOrderResponse(order *entity.Order) *OrderResponse {
	itemResponses := make([]OrderItemResponse, len(order.Items))
	for i, item := range order.Items {
		itemResponses[i] = OrderItemResponse{
			ID:       item.ID,
			Menu:     *ToMenuResponse(&item.Menu),
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
		FoodStatus: string(order.FoodStatus),
		Address:    order.Address,
		Items:      itemResponses,
		CreatedAt:  order.CreatedAt.Format(time.RFC3339),
		UpdatedAt:  order.UpdatedAt.Format(time.RFC3339),
	}
}

func ToOrderEntity(order *OrderResponse) *entity.Order {
	var item entity.OrderItem
	for _, itemResponse := range order.Items {
		item = entity.OrderItem{
			MenuID:   itemResponse.Menu.ID,
			Quantity: itemResponse.Quantity,
			Price:    itemResponse.Price,
		}
	}
	return &entity.Order{
		ID:         order.ID,
		CustomerID: order.Customer.ID,
		Customer:   entity.Customer{},
		Status:     entity.OrderStatus(order.Status),
		TotalPrice: order.TotalPrice,
		FoodStatus: entity.FoodStatus(order.FoodStatus),
		Address:    order.Address,
		Items:      []entity.OrderItem{item},
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
}
