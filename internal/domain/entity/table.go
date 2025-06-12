package entity

import (
	"time"

	"gorm.io/gorm"
)

type Table struct {
	ID           int64          `gorm:"column:id;primaryKey"`
	TableNumber  int            `gorm:"not null;unique"`
	Capacity     int            `gorm:"not null"`
	IsAvailable  bool           `gorm:"not null;default:true"`
	Reservations []Reservation  `gorm:"foreignKey:TableID;constraint:OnDelete:SET NULL"`
	CreatedAt    time.Time      `gorm:"created_at"`
	UpdatedAt    time.Time      `gorm:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"deleted_at"`
}

func (t *Table) TableName() string {
	return "tables"
}
