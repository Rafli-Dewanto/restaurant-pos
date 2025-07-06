package usecase

import (
	"cakestore/internal/constants"
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
	Register(request *model.CreateCustomerRequest, role string) (*entity.Customer, error)
	Login(request *model.LoginRequest) (*string, error)
	GetCustomerByID(id int64) (*entity.Customer, error)
	UpdateCustomer(id int64, request *model.UpdateUserRequest) error
	GetEmployees() ([]entity.Customer, error)
	GetEmployeeByID(id int64) (*entity.Customer, error)
	UpdateEmployee(id int64, request *model.UpdateUserRequest, role string) error
	DeleteEmployee(id int64) error
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

func (uc *customerUseCase) Register(request *model.CreateCustomerRequest, role string) (*entity.Customer, error) {
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

	if role == "" {
		role = constants.RoleCustomer
	}

	customer := &entity.Customer{
		Name:      request.Name,
		Email:     request.Email,
		Password:  string(hashedPassword),
		Address:   request.Address,
		Role:      role,
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
		"role":        customer.Role,
		"exp":         time.Now().Add(time.Hour * 24).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(uc.jwtSecret))
	if err != nil {
		uc.logger.Errorf("Error generating token: %v", err)
		return nil, err
	}

	return &tokenString, nil
}

func (uc *customerUseCase) GetCustomerByID(id int64) (*entity.Customer, error) {
	start := time.Now()
	defer func() {
		uc.logger.Infof("GetCustomerByID took %v", time.Since(start))
	}()

	customer, err := uc.repo.GetByID(id)
	if err != nil {
		uc.logger.Errorf("Error getting customer by ID: %v", err)
		return nil, err
	}
	return customer, nil
}

func (uc *customerUseCase) UpdateCustomer(id int64, request *model.UpdateUserRequest) error {
	customer, err := uc.repo.GetByID(id)
	if err != nil {
		return err
	}

	customer.Name = request.Name
	customer.Address = request.Address
	customer.Email = request.Email
	customer.UpdatedAt = time.Now()

	if err := uc.repo.Update(customer); err != nil {
		uc.logger.Errorf("Error updating customer: %v", err)
		return err
	}

	return nil
}

func (uc *customerUseCase) GetEmployees() ([]entity.Customer, error) {
	start := time.Now()
	defer func() {
		uc.logger.Infof("GetEmployees took %v", time.Since(start))
	}()

	employees, err := uc.repo.GetEmployees()
	if err != nil {
		uc.logger.Errorf("Error getting employees: %v", err)
		return nil, err
	}
	return employees, nil
}

func (uc *customerUseCase) GetEmployeeByID(id int64) (*entity.Customer, error) {
	start := time.Now()
	defer func() {
		uc.logger.Infof("GetEmployeeByID took %v", time.Since(start))
	}()

	employee, err := uc.repo.GetEmployeeByID(id)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return nil, err
		}
		uc.logger.Errorf("Error getting employee by ID: %v", err)
		return nil, err
	}
	return employee, nil
}

func (uc *customerUseCase) UpdateEmployee(id int64, request *model.UpdateUserRequest, role string) error {
	employee, err := uc.repo.GetEmployeeByID(id)
	if err != nil {
		return err
	}

	// Update fields
	employee.Name = request.Name
	employee.Email = request.Email
	employee.Address = request.Address
	employee.UpdatedAt = time.Now()
	if role != "" {
		employee.Role = role
	}

	if err := uc.repo.UpdateEmployee(id, request, role); err != nil {
		uc.logger.Errorf("Error updating employee: %v", err)
		return err
	}

	return nil
}

func (uc *customerUseCase) DeleteEmployee(id int64) error {
	if err := uc.repo.DeleteEmployee(id); err != nil {
		uc.logger.Errorf("Error deleting employee: %v", err)
		return err
	}
	return nil
}
