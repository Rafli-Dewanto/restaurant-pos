package entity

import (
	"database/sql"
	"time"
)

type WishList struct {
	ID         int          `gorm:"primaryKey"`
	CustomerID int          `gorm:"not null"`
	CakeID     int          `gorm:"column:cake_id"`
	CreatedAt  time.Time    `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt  time.Time    `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt  sql.NullTime `gorm:"column:deleted_at"`

	Customer Customer `gorm:"foreignKey:CustomerID"`
	Cake     Cake     `gorm:"foreignKey:CakeID"`
}

func (a *WishList) TableName() string {
	return "wishlists"
}
