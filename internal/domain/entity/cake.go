package entity

import (
	"database/sql"
	"time"
)

type Cake struct {
	ID          int          `gorm:"column:id;primaryKey"`
	Title       string       `gorm:"column:title"`
	Description string       `gorm:"column:description"`
	Price       float64      `gorm:"column:price"`
	Category    string       `gorm:"column:category"`
	Rating      float64      `gorm:"column:rating"`
	Image       string       `gorm:"column:image"`
	CreatedAt   time.Time    `gorm:"column:created_at"`
	UpdatedAt   time.Time    `gorm:"column:updated_at"`
	DeletedAt   sql.NullTime `gorm:"column:deleted_at"`
}

func (a *Cake) TableName() string {
	return "cakes"
}
