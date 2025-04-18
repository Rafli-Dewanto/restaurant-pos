package model

type AddCart struct {
	CakeID   int `json:"cake_id" validate:"required"`
	Quantity int `json:"quantity" validate:"required,min=1"`
}
