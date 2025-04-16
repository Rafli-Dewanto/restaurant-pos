package entity

import (
	"database/sql"
	"time"

	"gorm.io/gorm"
)

type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "pending"
	PaymentStatusSuccess   PaymentStatus = "success"
	PaymentStatusFailed    PaymentStatus = "failed"
	PaymentStatusExpired   PaymentStatus = "expired"
	PaymentStatusCancelled PaymentStatus = "cancelled"
)

type Payment struct {
	ID           int           `gorm:"column:id;primaryKey"`
	OrderID      int           `gorm:"column:order_id"`
	Order        Order         `gorm:"foreignKey:OrderID"`
	Amount       float64       `gorm:"column:amount"`
	Status       PaymentStatus `gorm:"column:status"`
	PaymentToken string        `gorm:"column:payment_token"`
	PaymentURL   string        `gorm:"column:payment_url"`
	CreatedAt    time.Time     `gorm:"column:created_at"`
	UpdatedAt    time.Time     `gorm:"column:updated_at"`
	DeletedAt    sql.NullTime  `gorm:"column:deleted_at"`
}

func (p *Payment) TableName() string {
	return "payments"
}

func (p *Payment) BeforeCreate(tx *gorm.DB) {
	p.Status = PaymentStatusPending
}
