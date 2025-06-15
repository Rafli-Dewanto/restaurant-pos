package seeder

import (
	"cakestore/internal/constants"
	"cakestore/internal/domain/entity"
	"cakestore/internal/domain/model"
	"cakestore/internal/repository"
	"time"

	"github.com/sirupsen/logrus"
)

type MenuSeeder struct {
	repo   repository.MenuRepository
	logger *logrus.Logger
}

func NewMenuSeeder(repo repository.MenuRepository, logger *logrus.Logger) *MenuSeeder {
	return &MenuSeeder{
		repo:   repo,
		logger: logger,
	}
}

func (s *MenuSeeder) SeedMenus() error {
	menus := []entity.Menu{
		{
			Title:       "Chocolate Fudge Cake",
			Description: "Rich and moist chocolate cake with fudge frosting",
			Rating:      4.5,
			Image:       "https://images.unsplash.com/photo-1586985289906-406988974504?w=900&auto=format&fit=crop&q=60&ixlib=rb-4.0.3&ixid=M3wxMjA3fDB8MHxzZWFyY2h8Mnx8Y2hvY29sYXRlJTIwY2FrZXxlbnwwfHwwfHx8MA%3D%3D",
			Price:       80000,
			Quantity:    20,
			Category:    constants.BirthdayCake,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			Title:       "Vanilla Bean Cheesecake",
			Description: "Creamy cheesecake with real vanilla beans",
			Rating:      4.8,
			Image:       "https://images.unsplash.com/photo-1568051243857-068aa3ea934d?w=900&auto=format&fit=crop&q=60&ixlib=rb-4.0.3&ixid=M3wxMjA3fDB8MHxzZWFyY2h8Nnx8dmFuaWxsYSUyMGNha2V8ZW58MHx8MHx8fDA%3D",
			Price:       90000,
			Quantity:    20,
			Category:    constants.WeddingCake,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			Title:       "Red Velvet Delight",
			Description: "Classic red velvet cake with cream cheese frosting",
			Rating:      4.7,
			Image:       "https://images.unsplash.com/photo-1586788680434-30d324b2d46f?w=900&auto=format&fit=crop&q=60&ixlib=rb-4.0.3&ixid=M3wxMjA3fDB8MHxzZWFyY2h8NHx8cmVkJTIwdmVsdmV0fGVufDB8fDB8fHww",
			Price:       120000,
			Quantity:    20,
			Category:    constants.CupCake,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			Title:       "Matcha Cookies",
			Description: "Soft and chewy cookie with matcha flavor",
			Rating:      4.5,
			Image:       "https://teakandthyme.com/wp-content/uploads/2023/09/matcha-white-chocolate-cookies-DSC_5105-1x1-1200.jpg",
			Price:       120000,
			Quantity:    20,
			Category:    constants.Cookies,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			Title:       "Strawberry Cupcake",
			Description: "Sweet and creamy cupcake with fresh strawberries",
			Rating:      4.9,
			Image:       "https://images.unsplash.com/photo-1563729784474-d77dbb933a9e?w=900&auto=format&fit=crop&q=60&ixlib=rb-4.1.0&ixid=M3wxMjA3fDB8MHxzZWFyY2h8Mnx8Y2FrZXxlbnwwfHwwfHx8MA%3D%3D",
			Price:       35000,
			Quantity:    30,
			Category:    constants.CupCake,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			Title:       "Tiramisu",
			Description: "Classic Italian dessert with layers of coffee-soaked ladyfingers and creamy mascarpone filling",
			Rating:      4.8,
			Image:       "https://images.unsplash.com/photo-1571115177098-24ec42ed204d?q=80&w=3177&auto=format&fit=crop&ixlib=rb-4.1.0&ixid=M3wxMjA3fDB8MHxwaG90by1wYWdlfHx8fGVufDB8fHx8fA%3D%3D",
			Price:       120000,
			Quantity:    20,
			Category:    constants.Other,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	// if already exists return
	menuData, err := s.repo.GetAll(&model.MenuQueryParams{})
	if err != nil {
		s.logger.Errorf("Error getting menus: %v", err)
		return err
	}

	for _, menu := range menuData.Data {
		if menu.Title == menus[0].Title {
			s.logger.Info("Menus already exist")
			return nil
		}
	}

	for _, m := range menus {
		if err := s.repo.Create(&m); err != nil {
			s.logger.Errorf("Error seeding menus %s: %v", m.Title, err)
			return err
		}
	}

	s.logger.Info("Menus seeded successfully")
	return nil
}
