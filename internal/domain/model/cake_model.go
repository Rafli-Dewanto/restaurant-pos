package model

import (
	"cakestore/internal/domain/entity"
)

type CakeModel struct {
	ID          int     `json:"id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Category    string  `json:"category"`
	Rating      float64 `json:"rating"`
	ImageURL    string  `json:"image_url"`
}

type CakeQueryParams struct {
	Page     int     `form:"page" binding:"omitempty,min=1"`
	PageSize int     `form:"page_size" binding:"omitempty,min=1,max=100"`
	Limit    int     `form:"limit" binding:"omitempty,min=1,max=100"`
	Title    string  `form:"title" binding:"omitempty"`
	MinPrice float64 `form:"min_price" binding:"omitempty,min=0"`
	MaxPrice float64 `form:"max_price" binding:"omitempty,min=0"`
	Category string  `form:"category" binding:"omitempty"`
}

type CreateUpdateCakeRequest struct {
	Title       string  `json:"title" validate:"required,min=3,max=100"`
	Description string  `json:"description" validate:"required"`
	Price       float64 `json:"price" validate:"required,gte=0"`
	Category    string  `json:"category" validate:"required"`
	Rating      float64 `json:"rating" validate:"required,gte=0,lte=10"`
	ImageURL    string  `json:"image" validate:"required,url"`
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
