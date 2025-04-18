package entity

import (
	"database/sql"
	"time"
)

type CartItem struct {
	ID       int     `gorm:"column:id;primaryKey"`
	CartID   int     `gorm:"column:cart_id"`
	CakeID   int     `gorm:"column:cake_id"`
	Quantity int     `gorm:"column:quantity"`
	Price    float64 `gorm:"column:price"`
	Subtotal float64 `gorm:"column:subtotal"`
}

type Cart struct {
	ID         int          `gorm:"column:id;primaryKey"`
	CustomerID int          `gorm:"column:customer_id"`
	Items      []CartItem   `gorm:"foreignKey:CartID"`
	Total      float64      `gorm:"column:total"`
	CreatedAt  time.Time    `gorm:"column:created_at"`
	UpdatedAt  time.Time    `gorm:"column:updated_at"`
	DeletedAt  sql.NullTime `gorm:"column:deleted_at"`
}

func (c *Cart) TableName() string {
	return "carts"
}

func (ci *CartItem) TableName() string {
	return "cart_items"
}
