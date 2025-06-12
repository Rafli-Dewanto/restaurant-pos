package entity

import (
	"time"

	"gorm.io/gorm"
)

type Inventory struct {
	ID              uint           `gorm:"primaryKey"`
	Name            string         `gorm:"type:varchar(100);not null"`
	Quantity        float64        `gorm:"not null"`
	Unit            string         `gorm:"type:varchar(50);not null"`
	MinimumStock    float64        `gorm:"not null"`
	ReorderPoint    float64        `gorm:"not null"`
	UnitPrice       float64        `gorm:"not null"`
	LastRestockDate time.Time      `gorm:"not null"`
	CreatedAt       time.Time      `gorm:"not null"`
	UpdatedAt       time.Time      `gorm:"not null"`
	DeletedAt       gorm.DeletedAt `gorm:"index"`
}

func (i *Inventory) TableName() string {
	return "inventories"
}
