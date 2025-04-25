package database

import (
	"cakestore/internal/domain/entity"
	"log"

	"gorm.io/gorm"
)

func RunMigrations(db *gorm.DB) error {
	log.Println("🔄 Running database migrations...")
	err := db.AutoMigrate(
		&entity.Cake{},
		&entity.Customer{},
		&entity.Order{},
		&entity.OrderItem{},
		&entity.Payment{},
		&entity.Cart{},
		&entity.WishList{},
	)
	if err != nil {
		return err
	}
	log.Println("✅ Database migrations completed successfully")
	return nil
}