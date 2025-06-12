package entity

import (
	"time"

	"gorm.io/gorm"
)

type ReservationStatus string

const (
	ReservationStatusPending   ReservationStatus = "pending"
	ReservationStatusConfirmed ReservationStatus = "confirmed"
	ReservationStatusCancelled ReservationStatus = "cancelled"
	ReservationStatusCompleted ReservationStatus = "completed"
)

type Reservation struct {
	ID           uint              `json:"id" gorm:"primaryKey"`
	CustomerID   uint              `json:"customer_id"`
	Customer     Customer          `json:"customer" gorm:"foreignKey:CustomerID"`
	TableNumber  int               `json:"table_number"`
	GuestCount   int               `json:"guest_count"`
	ReserveDate  time.Time         `json:"reserve_date"`
	Status       ReservationStatus `json:"status"`
	SpecialNotes string            `json:"special_notes"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
	DeletedAt    gorm.DeletedAt    `json:"deleted_at" gorm:"index"`
}
