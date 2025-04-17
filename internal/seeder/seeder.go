package seeder

import (
	"cakestore/internal/repository"

	"github.com/sirupsen/logrus"
)

type Seeder struct {
	customerSeeder *CustomerSeeder
	cakeSeeder     *CakeSeeder
	logger         *logrus.Logger
}

func NewSeeder(customerRepo repository.CustomerRepository, cakeRepo repository.CakeRepository, logger *logrus.Logger) *Seeder {
	return &Seeder{
		customerSeeder: NewCustomerSeeder(customerRepo, logger),
		cakeSeeder:     NewCakeSeeder(cakeRepo, logger),
		logger:         logger,
	}
}

func (s *Seeder) SeedAll() error {
	s.logger.Info("Starting database seeding...")

	// Seed admin user
	if err := s.customerSeeder.SeedAdmin("admin@example.com", "master123"); err != nil {
		s.logger.Errorf("Error seeding admin user: %v", err)
		return err
	}

	// Seed cakes
	if err := s.cakeSeeder.SeedCakes(); err != nil {
		s.logger.Errorf("Error seeding cakes: %v", err)
		return err
	}

	s.logger.Info("Database seeding completed successfully")
	return nil
}
