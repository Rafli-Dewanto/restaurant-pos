package entity

import (
	"database/sql"
	"time"
)

type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusPaid      OrderStatus = "paid"
	OrderStatusPreparing OrderStatus = "preparing"
	OrderStatusDelivered OrderStatus = "delivered"
	OrderStatusCancelled OrderStatus = "cancelled"
)

type Order struct {
	ID         int64        `gorm:"column:id;primaryKey;autoIncrement"`
	CustomerID int64        `gorm:"column:customer_id"`
	Customer   Customer     `gorm:"foreignKey:CustomerID"`
	Status     OrderStatus  `gorm:"column:status"`
	TotalPrice float64      `gorm:"column:total_price"`
	Address    string       `gorm:"column:delivery_address"`
	Items      []OrderItem  `gorm:"foreignKey:OrderID"`
	CreatedAt  time.Time    `gorm:"column:created_at"`
	UpdatedAt  time.Time    `gorm:"column:updated_at"`
	DeletedAt  sql.NullTime `gorm:"column:deleted_at"`
}

type OrderItem struct {
	ID        int64        `gorm:"column:id;primaryKey;autoIncrement"`
	OrderID   int64        `gorm:"column:order_id"`
	CakeID    int64        `gorm:"column:cake_id"`
	Cake      Cake         `gorm:"foreignKey:CakeID"`
	Quantity  int64        `gorm:"column:quantity"`
	Price     float64      `gorm:"column:price"`
	CreatedAt time.Time    `gorm:"column:created_at"`
	UpdatedAt time.Time    `gorm:"column:updated_at"`
	DeletedAt sql.NullTime `gorm:"column:deleted_at"`
}

func (o *Order) TableName() string {
	return "orders"
}

func (oi *OrderItem) TableName() string {
	return "order_items"
}
