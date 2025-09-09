package seeder

import (
	"cakestore/internal/constants"
	"cakestore/internal/domain/entity"
	"cakestore/internal/domain/model"
	"cakestore/internal/repository"
	"fmt"
	"math/rand"
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
	// Seed only if no data exists to avoid re-seeding 100,000 records every time
	menuData, err := s.repo.GetAll(&model.MenuQueryParams{Limit: 1}) // Just check if any record exists
	if err != nil {
		s.logger.Errorf("Error checking for existing menus: %v", err)
		return err
	}

	if len(menuData.Data) > 0 {
		s.logger.Info("Menus already exist, skipping seeding.")
		return nil
	}

	const numberOfMenusToSeed = 1000
	menus := make([]entity.Menu, 0, numberOfMenusToSeed) // Pre-allocate slice capacity

	// Seed random source for more varied dummy data
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Define categories to cycle through
	categories := []string{
		constants.BirthdayCake,
		constants.WeddingCake,
		constants.CupCake,
		constants.Cookies,
		constants.Other,
	}

	s.logger.Infof("Starting to generate %d dummy menu items...", numberOfMenusToSeed)

	for i := 0; i < numberOfMenusToSeed; i++ {
		title := fmt.Sprintf("Dummy Cake %d", i+1)
		description := fmt.Sprintf("Delicious dummy cake number %d for all occasions.", i+1)
		imageURL := fmt.Sprintf("https://dummyimage.com/600x400/000/fff&text=Cake+%d", i+1) // Generic dummy image URL

		menu := entity.Menu{
			Title:       title,
			Description: description,
			Rating:      float64(r.Intn(100)+300) / 100.0, // Random rating between 3.0 and 4.0
			Image:       imageURL,
			Price:       float64(r.Intn(100000) + 20000),     // Random price between 20,000 and 120,000
			Quantity:    int64(r.Intn(50) + 10),              // Random quantity between 10 and 60
			Category:    categories[r.Intn(len(categories))], // Cycle through categories
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		menus = append(menus, menu)

		// Log progress periodically
		if (i+1)%10000 == 0 {
			s.logger.Infof("Generated %d / %d menu items...", i+1, numberOfMenusToSeed)
		}
	}

	s.logger.Infof("Finished generating %d dummy menu items. Starting database insertion...", numberOfMenusToSeed)

	// --- IMPORTANT: CONSIDER BATCH INSERTION HERE ---
	// Inserting 100,000 records one by one can be very slow.
	// It's highly recommended to implement a batch insert method in your MenuRepository.
	// For example: `s.repo.BulkCreate(menus)`
	// If you don't have a batch insert, the loop below will work but will be slow.

	for i, m := range menus {
		if err := s.repo.Create(&m); err != nil {
			s.logger.Errorf("Error seeding menu %s (item %d/%d): %v", m.Title, i+1, numberOfMenusToSeed, err)
			return err
		}
		// Log insertion progress periodically
		if (i+1)%5000 == 0 {
			s.logger.Infof("Inserted %d / %d menu items into the database...", i+1, numberOfMenusToSeed)
		}
	}

	s.logger.Infof("Successfully seeded %d menus.", numberOfMenusToSeed)
	return nil
}
