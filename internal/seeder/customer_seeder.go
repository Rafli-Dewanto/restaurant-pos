package seeder

import (
	"cakestore/internal/domain/entity"
	"cakestore/internal/repository"
	"cakestore/utils"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

type CustomerSeeder struct {
	repo   repository.CustomerRepository
	logger *logrus.Logger
}

func NewCustomerSeeder(repo repository.CustomerRepository, logger *logrus.Logger) *CustomerSeeder {
	return &CustomerSeeder{
		repo:   repo,
		logger: logger,
	}
}

func (s *CustomerSeeder) SeedAdmin(email, password string) error {
	// Check if admin already exists
	existingAdmin, err := s.repo.GetByEmail(email)
	if err == nil && existingAdmin != nil {
		s.logger.Info("Admin user already exists")
		return nil
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Errorf("Error hashing password: %v", err)
		return err
	}

	// Create admin user
	admin := &entity.Customer{
		Name:      "Admin",
		Email:     email,
		Password:  string(hashedPassword),
		Address:   "Admin Address",
		Role:      utils.RoleAdmin,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.repo.Create(admin); err != nil {
		s.logger.Errorf("Error creating admin user: %v", err)
		return err
	}

	s.logger.Info("Admin user created successfully")
	return nil
}
