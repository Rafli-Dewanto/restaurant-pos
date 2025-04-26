package entity

import (
	"database/sql"
	"time"
)

type Cart struct {
	ID         int64        `gorm:"column:id;primaryKey"`
	CustomerID int64        `gorm:"column:customer_id"`
	CakeID     int64        `gorm:"column:cake_id"`
	Quantity   int64        `gorm:"column:quantity"`
	Price      float64      `gorm:"column:price"`
	Subtotal   float64      `gorm:"column:subtotal"`
	CreatedAt  time.Time    `gorm:"column:created_at"`
	UpdatedAt  time.Time    `gorm:"column:updated_at"`
	DeletedAt  sql.NullTime `gorm:"column:deleted_at"`
}

func (c *Cart) TableName() string {
	return "carts"
}
