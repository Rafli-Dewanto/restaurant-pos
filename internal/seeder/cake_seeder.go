package seeder

import (
	"cakestore/internal/constants"
	"cakestore/internal/domain/entity"
	"cakestore/internal/domain/model"
	"cakestore/internal/repository"
	"time"

	"github.com/sirupsen/logrus"
)

type CakeSeeder struct {
	repo   repository.CakeRepository
	logger *logrus.Logger
}

func NewCakeSeeder(repo repository.CakeRepository, logger *logrus.Logger) *CakeSeeder {
	return &CakeSeeder{
		repo:   repo,
		logger: logger,
	}
}

func (s *CakeSeeder) SeedCakes() error {
	cakes := []entity.Cake{
		{
			Title:       "Chocolate Fudge Cake",
			Description: "Rich and moist chocolate cake with fudge frosting",
			Rating:      4.5,
			Image:       "https://images.unsplash.com/photo-1586985289906-406988974504?w=900&auto=format&fit=crop&q=60&ixlib=rb-4.0.3&ixid=M3wxMjA3fDB8MHxzZWFyY2h8Mnx8Y2hvY29sYXRlJTIwY2FrZXxlbnwwfHwwfHx8MA%3D%3D",
			Price:       80000,
			Category:    constants.BirthdayCake,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			Title:       "Vanilla Bean Cheesecake",
			Description: "Creamy cheesecake with real vanilla beans",
			Rating:      4.3,
			Image:       "https://images.unsplash.com/photo-1568051243857-068aa3ea934d?w=900&auto=format&fit=crop&q=60&ixlib=rb-4.0.3&ixid=M3wxMjA3fDB8MHxzZWFyY2h8Nnx8dmFuaWxsYSUyMGNha2V8ZW58MHx8MHx8fDA%3D",
			Price:       90000,
			Category:    constants.WeddingCake,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			Title:       "Red Velvet Delight",
			Description: "Classic red velvet cake with cream cheese frosting",
			Rating:      4.4,
			Image:       "https://images.unsplash.com/photo-1586788680434-30d324b2d46f?w=900&auto=format&fit=crop&q=60&ixlib=rb-4.0.3&ixid=M3wxMjA3fDB8MHxzZWFyY2h8NHx8cmVkJTIwdmVsdmV0fGVufDB8fDB8fHww",
			Price:       120000,
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
			Category:    constants.Cookies,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	// if already exists return
	cakeData, err := s.repo.GetAll(&model.CakeQueryParams{})
	if err != nil {
		s.logger.Errorf("Error getting cakes: %v", err)
		return err
	}
	
	for _, cake := range cakeData.Data {
		if cake.Title == cakes[0].Title {
			s.logger.Info("Cakes already exist")
			return nil
		}
	}

	for _, cake := range cakes {
		if err := s.repo.Create(&cake); err != nil {
			s.logger.Errorf("Error seeding cake %s: %v", cake.Title, err)
			return err
		}
	}

	s.logger.Info("Cakes seeded successfully")
	return nil
}
