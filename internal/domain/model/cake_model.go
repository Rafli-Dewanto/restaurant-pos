package model

import (
	"cakestore/internal/domain/entity"
)

type CakeModel struct {
	ID          int64   `json:"id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Category    string  `json:"category"`
	Rating      float64 `json:"rating"`
	ImageURL    string  `json:"image"`
}

type CakeQueryParams struct {
	Page     int64   `form:"page" binding:"omitempty,min=1"`
	PageSize int64   `form:"page_size" binding:"omitempty,min=1,max=100"`
	Limit    int64   `form:"limit" binding:"omitempty,min=1,max=100"`
	Title    string  `form:"title" binding:"omitempty"`
	MinPrice float64 `form:"min_price" binding:"omitempty,min=0"`
	MaxPrice float64 `form:"max_price" binding:"omitempty,min=0"`
	Category string  `form:"category" binding:"omitempty"`
}

type CreateUpdateCakeRequest struct {
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
	CakeName   string  `json:"name"`
	CakeID     int64   `json:"cake_id"`
	CakeImage  string  `json:"image"`
	Quantity   int64   `json:"quantity"`
	Price      float64 `json:"price"`
	Subtotal   float64 `json:"subtotal"`
	CreatedAt  string  `json:"created_at"`
	UpdatedAt  string  `json:"updated_at"`
}

func CakeToResponse(cake *entity.Cake) *CakeModel {
	return &CakeModel{
		ID:          cake.ID,
		Title:       cake.Title,
		Description: cake.Description,
		Rating:      float64(cake.Rating),
		ImageURL:    cake.Image,
		Price:       cake.Price,
		Category:    cake.Category,
	}
}

func CakeToEntity(cake *CakeModel) *entity.Cake {
	return &entity.Cake{
		ID:          cake.ID,
		Title:       cake.Title,
		Description: cake.Description,
		Price:       cake.Price,
		Category:    cake.Category,
		Rating:      float64(cake.Rating),
		Image:       cake.ImageURL,
	}
}
