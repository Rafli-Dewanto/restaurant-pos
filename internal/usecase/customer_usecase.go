package usecase

import (
	"cakestore/internal/domain/entity"
	"cakestore/internal/domain/model"
	"cakestore/internal/repository"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

type CustomerUseCase interface {
	Register(request *model.CreateCustomerRequest) (*entity.Customer, error)
	Login(request *model.LoginRequest) (*string, error)
	GetCustomerByID(id int) (*entity.Customer, error)
	UpdateCustomer(id int, request *model.CreateCustomerRequest) error
}

type customerUseCase struct {
	repo      repository.CustomerRepository
	logger    *logrus.Logger
	jwtSecret string
}

func NewCustomerUseCase(repo repository.CustomerRepository, logger *logrus.Logger, jwtSecret string) CustomerUseCase {
	return &customerUseCase{
		repo:      repo,
		logger:    logger,
		jwtSecret: jwtSecret,
	}
}

func (uc *customerUseCase) Register(request *model.CreateCustomerRequest) (*entity.Customer, error) {
	// Check if email already exists
	existingCustomer, err := uc.repo.GetByEmail(request.Email)
	if err == nil && existingCustomer != nil {
		return nil, errors.New("email already registered")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		uc.logger.Errorf("Error hashing password: %v", err)
		return nil, err
	}

	customer := &entity.Customer{
		Name:      request.Name,
		Email:     request.Email,
		Password:  string(hashedPassword),
		Address:   request.Address,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := uc.repo.Create(customer); err != nil {
		uc.logger.Errorf("Error creating customer: %v", err)
		return nil, err
	}

	return customer, nil
}

func (uc *customerUseCase) Login(request *model.LoginRequest) (*string, error) {
	customer, err := uc.repo.GetByEmail(request.Email)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(customer.Password), []byte(request.Password)); err != nil {
		return nil, errors.New("invalid email or password")
	}

	// Generate JWT token
	claims := jwt.MapClaims{
		"customer_id": customer.ID,
		"email":       customer.Email,
		"exp":         time.Now().Add(time.Hour * 24).Unix(), // Token expires in 24 hours
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(uc.jwtSecret))
	if err != nil {
		uc.logger.Errorf("Error generating token: %v", err)
		return nil, err
	}

	return &tokenString, nil
}

func (uc *customerUseCase) GetCustomerByID(id int) (*entity.Customer, error) {
	customer, err := uc.repo.GetByID(id)
	if err != nil {
		uc.logger.Errorf("Error getting customer by ID: %v", err)
		return nil, err
	}
	return customer, nil
}

func (uc *customerUseCase) UpdateCustomer(id int, request *model.CreateCustomerRequest) error {
	customer, err := uc.repo.GetByID(id)
	if err != nil {
		return err
	}

	// Update fields
	customer.Name = request.Name
	customer.Address = request.Address
	customer.UpdatedAt = time.Now()

	// Update password if provided
	if request.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
		if err != nil {
			uc.logger.Errorf("Error hashing password: %v", err)
			return err
		}
		customer.Password = string(hashedPassword)
	}

	if err := uc.repo.Update(customer); err != nil {
		uc.logger.Errorf("Error updating customer: %v", err)
		return err
	}

	return nil
}
