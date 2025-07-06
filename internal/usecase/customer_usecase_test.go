package usecase

import (
	"cakestore/internal/database"
	"cakestore/internal/domain/entity"
	"cakestore/internal/domain/model"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/sirupsen/logrus"
)

type MockCustomerRepository struct {
	mock.Mock
}

func (m *MockCustomerRepository) Create(customer *entity.Customer) error {
	args := m.Called(customer)
	return args.Error(0)
}

func (m *MockCustomerRepository) GetByEmail(email string) (*entity.Customer, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Customer), args.Error(1)
}

func (m *MockCustomerRepository) GetByID(id int64) (*entity.Customer, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Customer), args.Error(1)
}

func (m *MockCustomerRepository) Update(customer *entity.Customer) error {
	args := m.Called(customer)
	return args.Error(0)
}

func (m *MockCustomerRepository) GetEmployees() ([]entity.Customer, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.Customer), args.Error(1)
}

func (m *MockCustomerRepository) GetEmployeeByID(id int64) (*entity.Customer, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Customer), args.Error(1)
}

func (m *MockCustomerRepository) UpdateEmployee(id int64, request *model.UpdateUserRequest, role string) error {
	args := m.Called(id, request, role)
	return args.Error(0)
}

func (m *MockCustomerRepository) DeleteEmployee(id int64) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockCustomerRepository) Delete(id int64) error {
	args := m.Called(id)
	return args.Error(0)
}

func TestCustomerUseCase_GetCustomerByID(t *testing.T) {
	logger := logrus.New()
	mockCustomerRepo := new(MockCustomerRepository)
	mockCache := new(database.MockRedisCacheService)
	useCase := NewCustomerUseCase(mockCustomerRepo, logger, "secret", mockCache)

	t.Run("success", func(t *testing.T) {
		expectedCustomer := &entity.Customer{
			ID:        1,
			Name:      "Test User",
			Email:     "test@example.com",
			Address:   "123 Test St",
			Role:      "customer",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		mockCache.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("not found"))
		mockCustomerRepo.On("GetByID", int64(1)).Return(expectedCustomer, nil).Once()
		mockCache.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

		customer, err := useCase.GetCustomerByID(1)

		assert.NoError(t, err)
		assert.NotNil(t, customer)
		assert.Equal(t, expectedCustomer.ID, customer.ID)
		mockCustomerRepo.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		mockCache.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("not found"))
		mockCustomerRepo.On("GetByID", int64(1)).Return(nil, errors.New("not found")).Once()

		customer, err := useCase.GetCustomerByID(1)

		assert.Error(t, err)
		assert.Nil(t, customer)
		mockCustomerRepo.AssertExpectations(t)
	})
}

func TestCustomerUseCase_GetEmployees(t *testing.T) {
	logger := logrus.New()
	mockCustomerRepo := new(MockCustomerRepository)
	mockCache := new(database.MockRedisCacheService)
	useCase := NewCustomerUseCase(mockCustomerRepo, logger, "secret", mockCache)

	t.Run("success", func(t *testing.T) {
		expectedEmployees := []entity.Customer{
			{
				ID:        1,
				Name:      "Test Employee",
				Email:     "employee@example.com",
				Address:   "123 Test St",
				Role:      "employee",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		}
		mockCache.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("not found"))
		mockCustomerRepo.On("GetEmployees").Return(expectedEmployees, nil).Once()
		mockCache.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

		employees, err := useCase.GetEmployees()

		assert.NoError(t, err)
		assert.NotNil(t, employees)
		assert.Equal(t, len(expectedEmployees), len(employees))
		mockCustomerRepo.AssertExpectations(t)
	})

	t.Run("error", func(t *testing.T) {
		mockCache.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("not found"))
		mockCustomerRepo.On("GetEmployees").Return(nil, errors.New("error")).Once()

		employees, err := useCase.GetEmployees()

		assert.Error(t, err)
		assert.Nil(t, employees)
		mockCustomerRepo.AssertExpectations(t)
	})
}

func TestCustomerUseCase_GetEmployeeByID(t *testing.T) {
	logger := logrus.New()
	mockCustomerRepo := new(MockCustomerRepository)
	mockCache := new(database.MockRedisCacheService)
	useCase := NewCustomerUseCase(mockCustomerRepo, logger, "secret", mockCache)

	t.Run("success", func(t *testing.T) {
		expectedEmployee := &entity.Customer{
			ID:        1,
			Name:      "Test Employee",
			Email:     "employee@example.com",
			Address:   "123 Test St",
			Role:      "employee",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		mockCache.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("not found"))
		mockCustomerRepo.On("GetEmployeeByID", int64(1)).Return(expectedEmployee, nil).Once()
		mockCache.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

		employee, err := useCase.GetEmployeeByID(1)

		assert.NoError(t, err)
		assert.NotNil(t, employee)
		assert.Equal(t, expectedEmployee.ID, employee.ID)
		mockCustomerRepo.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		mockCache.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("not found"))
		mockCustomerRepo.On("GetEmployeeByID", int64(1)).Return(nil, errors.New("not found")).Once()

		employee, err := useCase.GetEmployeeByID(1)

		assert.Error(t, err)
		assert.Nil(t, employee)
		mockCustomerRepo.AssertExpectations(t)
	})
}
