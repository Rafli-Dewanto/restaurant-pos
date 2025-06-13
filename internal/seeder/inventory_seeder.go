package seeder

import (
	"cakestore/internal/domain/entity"
	"cakestore/internal/repository"
	"time"

	"github.com/sirupsen/logrus"
)

type InventorySeeder struct {
	repo   repository.InventoryRepository
	logger *logrus.Logger
}

func NewInventorySeeder(repo repository.InventoryRepository, logger *logrus.Logger) *InventorySeeder {
	return &InventorySeeder{
		repo:   repo,
		logger: logger,
	}
}

func (s *InventorySeeder) Seed() error {
	count, err := s.repo.Count()
	if err != nil {
		s.logger.Errorf("Error checking inventory count: %v", err)
		return err
	}

	if count > 0 {
		s.logger.Info("Inventory already seeded, skipping...")
		return nil
	}

	ingredients := []entity.Inventory{
		{
			Name:            "All-Purpose Flour",
			Quantity:        50000,
			Unit:            "grams",
			MinimumStock:    5000,
			ReorderPoint:    10000,
			UnitPrice:       40000,
			LastRestockDate: time.Now(),
		},
		{
			Name:            "Granulated Sugar",
			Quantity:        40000,
			Unit:            "grams",
			MinimumStock:    4000,
			ReorderPoint:    8000,
			UnitPrice:       32000,
			LastRestockDate: time.Now(),
		},
		{
			Name:            "Butter",
			Quantity:        20000,
			Unit:            "grams",
			MinimumStock:    2000,
			ReorderPoint:    4000,
			UnitPrice:       60000,
			LastRestockDate: time.Now(),
		},
		{
			Name:            "Eggs",
			Quantity:        1000,
			Unit:            "pieces",
			MinimumStock:    100,
			ReorderPoint:    200,
			UnitPrice:       3000,
			LastRestockDate: time.Now(),
		},
		{
			Name:            "Milk",
			Quantity:        30000,
			Unit:            "ml",
			MinimumStock:    3000,
			ReorderPoint:    6000,
			UnitPrice:       34000,
			LastRestockDate: time.Now(),
		},
		{
			Name:            "Vanilla Extract",
			Quantity:        2000,
			Unit:            "ml",
			MinimumStock:    200,
			ReorderPoint:    400,
			UnitPrice:       15000,
			LastRestockDate: time.Now(),
		},
		{
			Name:            "Baking Powder",
			Quantity:        5000,
			Unit:            "grams",
			MinimumStock:    500,
			ReorderPoint:    1000,
			UnitPrice:       10000,
			LastRestockDate: time.Now(),
		},
		{
			Name:            "Salt",
			Quantity:        3000,
			Unit:            "grams",
			MinimumStock:    300,
			ReorderPoint:    600,
			UnitPrice:       10000,
			LastRestockDate: time.Now(),
		},
	}

	for _, ingredient := range ingredients {
		if err := s.repo.Create(&ingredient); err != nil {
			s.logger.Errorf("Error seeding ingredient %s: %v", ingredient.Name, err)
			return err
		}
	}

	s.logger.Info("Inventory seeding completed successfully.")
	return nil
}
