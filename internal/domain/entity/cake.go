package entity

import (
	"database/sql"
	"time"
)

type Cake struct {
	ID          int64        `db:"id" gorm:"column:id;primaryKey"`
	Title       string       `db:"title" gorm:"column:title"`
	Description string       `db:"description" gorm:"column:description"`
	Price       float64      `db:"price" gorm:"column:price"`
	Category    string       `db:"category" gorm:"column:category"`
	Rating      float64      `db:"rating" gorm:"column:rating"`
	Image       string       `db:"image" gorm:"column:image"`
	CreatedAt   time.Time    `db:"created_at" gorm:"column:created_at"`
	UpdatedAt   time.Time    `db:"updated_at" gorm:"column:updated_at"`
	DeletedAt   sql.NullTime `db:"deleted_at" gorm:"column:deleted_at"`
}

func (a *Cake) TableName() string {
	return "cakes"
}
