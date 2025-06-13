package model

import (
	"time"

	"github.com/go-playground/validator/v10"
)

type ReservationQueryParams struct {
	PaginationQuery
	CustomerID  uint      `query:"customer_id"`
	Status      string    `query:"status"`
	ReserveDate time.Time `query:"reserve_date"`
	TableNumber int       `query:"table_number"`
}

type CreateReservationRequest struct {
	TableID      uint      `json:"table_id"`
	GuestCount   int       `json:"guest_count" validate:"required,min=1"`
	ReserveDate  time.Time `json:"reserve_date" validate:"required,future"`
	SpecialNotes string    `json:"special_notes"`
}

type UpdateReservationRequest struct {
	Status       string    `json:"status" validate:"omitempty,oneof=pending confirmed cancelled completed"`
	TableNumber  int       `json:"table_number" validate:"omitempty,min=1"`
	GuestCount   int       `json:"guest_count" validate:"omitempty,min=1"`
	ReserveDate  time.Time `json:"reserve_date" validate:"omitempty,future"`
	SpecialNotes string    `json:"special_notes"`
}

type ReservationResponse struct {
	ID           uint             `json:"id"`
	CustomerID   uint             `json:"customer_id"`
	Customer     CustomerResponse `json:"customer"`
	TableNumber  int              `json:"table_number"`
	GuestCount   int              `json:"guest_count"`
	ReserveDate  time.Time        `json:"reserve_date"`
	Status       string           `json:"status"`
	SpecialNotes string           `json:"special_notes"`
	CreatedAt    time.Time        `json:"created_at"`
	UpdatedAt    time.Time        `json:"updated_at"`
}

func (r *CreateReservationRequest) Validate() error {
	validate := validator.New()

	// Register custom validation for future dates
	_ = validate.RegisterValidation("future", func(fl validator.FieldLevel) bool {
		date, ok := fl.Field().Interface().(time.Time)
		if !ok {
			return false
		}
		return date.After(time.Now())
	})

	return validate.Struct(r)
}
