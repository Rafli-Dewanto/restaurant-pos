package entity

import (
	"database/sql"
	"time"
)

type Customer struct {
	ID        int64        `gorm:"column:id;primaryKey"`
	Name      string       `gorm:"column:name"`
	Email     string       `gorm:"column:email;unique"`
	Password  string       `gorm:"column:password"`
	Address   string       `gorm:"column:address"`
	Role      string       `gorm:"column:role;default:customer"`
	CreatedAt time.Time    `gorm:"column:created_at"`
	UpdatedAt time.Time    `gorm:"column:updated_at"`
	DeletedAt sql.NullTime `gorm:"column:deleted_at"`
}

func (c *Customer) TableName() string {
	return "customers"
}
