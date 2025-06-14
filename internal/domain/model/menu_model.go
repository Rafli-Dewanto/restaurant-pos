package model

import (
	"cakestore/internal/domain/entity"
)

type MenuModel struct {
	ID          int64   `json:"id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Quantity    int64   `json:"quantity"`
	Category    string  `json:"category"`
	Rating      float64 `json:"rating"`
	ImageURL    string  `json:"image"`
}

type MenuQueryParams struct {
	Page     int64   `form:"page" binding:"omitempty,min=1"`
	PageSize int64   `form:"page_size" binding:"omitempty,min=1,max=100"`
	Limit    int64   `form:"limit" binding:"omitempty,min=1,max=100"`
	Title    string  `form:"title" binding:"omitempty"`
	MinPrice float64 `form:"min_price" binding:"omitempty,min=0"`
	MaxPrice float64 `form:"max_price" binding:"omitempty,min=0"`
	Category string  `form:"category" binding:"omitempty"`
}

type CreateUpdateMenuRequest struct {
	Title       string  `json:"title" validate:"required,min=3,max=100"`
	Description string  `json:"description" validate:"required"`
	Price       float64 `json:"price" validate:"required,gte=0"`
	Quantity    int64   `json:"quantity" validate:"required,min=1"`
	Category    string  `json:"category" validate:"required"`
	Rating      float64 `json:"rating"`
	ImageURL    string  `json:"image" validate:"required,url"`
}

type UserCartResponse struct {
	ID         int64   `json:"id"`
	CustomerID int64   `json:"customer_id"`
	MenuName   string  `json:"name"`
	MenuID     int64   `json:"menu_id"`
	MenuImage  string  `json:"image"`
	Quantity   int64   `json:"quantity"`
	Price      float64 `json:"price"`
	Subtotal   float64 `json:"subtotal"`
	CreatedAt  string  `json:"created_at"`
	UpdatedAt  string  `json:"updated_at"`
}

func ToMenuResponse(menu *entity.Menu) *MenuModel {
	return &MenuModel{
		ID:          menu.ID,
		Title:       menu.Title,
		Description: menu.Description,
		Rating:      float64(menu.Rating),
		ImageURL:    menu.Image,
		Price:       menu.Price,
		Quantity:    menu.Quantity,
		Category:    menu.Category,
	}
}

func ToMenuEntity(menu *MenuModel) *entity.Menu {
	return &entity.Menu{
		ID:          menu.ID,
		Title:       menu.Title,
		Description: menu.Description,
		Price:       menu.Price,
		Category:    menu.Category,
		Rating:      float64(menu.Rating),
		Image:       menu.ImageURL,
	}
}
