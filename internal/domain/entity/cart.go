package entity

import (
	"database/sql"
	"time"
)

type Cart struct {
	ID         int          `gorm:"column:id;primaryKey"`
	CustomerID int          `gorm:"column:customer_id"`
	CakeID     int          `gorm:"column:cake_id"`
	Quantity   int          `gorm:"column:quantity"`
	Price      float64      `gorm:"column:price"`
	Subtotal   float64      `gorm:"column:subtotal"`
	CreatedAt  time.Time    `gorm:"column:created_at"`
	UpdatedAt  time.Time    `gorm:"column:updated_at"`
	DeletedAt  sql.NullTime `gorm:"column:deleted_at"`
}

func (c *Cart) TableName() string {
	return "carts"
}
