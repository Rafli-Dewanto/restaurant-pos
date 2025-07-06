package usecase

import (
	"cakestore/internal/domain/entity"
	"cakestore/internal/domain/model"
	"errors"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockOrderRepository struct {
	mock.Mock
}

func (m *MockOrderRepository) Create(order *entity.Order) error {
	args := m.Called(order)
	return args.Error(0)
}

func (m *MockOrderRepository) GetByID(id int64) (*entity.Order, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Order), args.Error(1)
}

func (m *MockOrderRepository) GetPendingPaymentByOrderID(customerID, orderID int64) (entity.Order, error) {
	args := m.Called(customerID, orderID)
	if args.Get(0) == nil {
		return entity.Order{}, args.Error(1)
	}
	return args.Get(0).(entity.Order), args.Error(1)
}

func (m *MockOrderRepository) GetAll(params *model.PaginationQuery) ([]entity.Order, *model.PaginatedMeta, error) {
	args := m.Called(params)
	if args.Get(0) == nil {
		return nil, nil, args.Error(2)
	}
	return args.Get(0).([]entity.Order), args.Get(1).(*model.PaginatedMeta), args.Error(2)
}

func (m *MockOrderRepository) GetByCustomerID(customerID int64) ([]entity.Order, error) {
	args := m.Called(customerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.Order), args.Error(1)
}

func (m *MockOrderRepository) UpdateStatus(id int64, status entity.OrderStatus) error {
	args := m.Called(id, status)
	return args.Error(0)
}

func (m *MockOrderRepository) Delete(id int64) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockOrderRepository) UpdateFoodStatus(orderID int64, foodStatus entity.FoodStatus) error {
	args := m.Called(orderID, foodStatus)
	return args.Error(0)
}

func (m *MockOrderRepository) GetPendingOrder() (int64, error) {
	args := m.Called()
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockOrderRepository) FindByDateRange(startDate, endDate string) ([]entity.Order, error) {
	args := m.Called(startDate, endDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.Order), args.Error(1)
}

func (m *MockOrderRepository) Update(order *entity.Order) error {
	args := m.Called(order)
	return args.Error(0)
}

func TestOrderUseCase_GetOrderByID(t *testing.T) {
	logger := logrus.New()
	mockOrderRepo := new(MockOrderRepository)
	useCase := NewOrderUseCase(mockOrderRepo, nil, nil, logger, "test")

	t.Run("success", func(t *testing.T) {
		expectedOrder := &entity.Order{
			ID: 1,
		}
		mockOrderRepo.On("GetByID", int64(1)).Return(expectedOrder, nil).Once()

		order, err := useCase.GetOrderByID(1)

		assert.NoError(t, err)
		assert.NotNil(t, order)
		assert.Equal(t, expectedOrder.ID, order.ID)
		mockOrderRepo.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		mockOrderRepo.On("GetByID", int64(1)).Return(nil, errors.New("not found")).Once()

		order, err := useCase.GetOrderByID(1)

		assert.Error(t, err)
		assert.Nil(t, order)
		mockOrderRepo.AssertExpectations(t)
	})
}

func TestOrderUseCase_GetPendingOrder(t *testing.T) {
	logger := logrus.New()
	mockOrderRepo := new(MockOrderRepository)
	useCase := NewOrderUseCase(mockOrderRepo, nil, nil, logger, "test")

	t.Run("success", func(t *testing.T) {
		expectedOrder := entity.Order{
			ID: 1,
		}
		mockOrderRepo.On("GetPendingPaymentByOrderID", int64(1), int64(1)).Return(expectedOrder, nil).Once()

		order, err := useCase.GetPendingOrder(1, 1)

		assert.NoError(t, err)
		assert.NotNil(t, order)
		assert.Equal(t, expectedOrder.ID, order.ID)
		mockOrderRepo.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		mockOrderRepo.On("GetPendingPaymentByOrderID", int64(1), int64(1)).Return(nil, errors.New("not found")).Once()

		order, err := useCase.GetPendingOrder(1, 1)

		assert.Error(t, err)
		assert.Nil(t, order)
		mockOrderRepo.AssertExpectations(t)
	})
}

func TestOrderUseCase_GetAllOrders(t *testing.T) {
	logger := logrus.New()
	mockOrderRepo := new(MockOrderRepository)
	useCase := NewOrderUseCase(mockOrderRepo, nil, nil, logger, "test")

	t.Run("success", func(t *testing.T) {
		expectedResponse := []entity.Order{
			{
				ID: 1,
			},
		}
		meta := &model.PaginatedMeta{
			Total: 1,
		}
		params := &model.PaginationQuery{Page: 1, Limit: 10}
		mockOrderRepo.On("GetAll", params).Return(expectedResponse, meta, nil).Once()

		orders, resultMeta, err := useCase.GetAllOrders(params)

		assert.NoError(t, err)
		assert.NotNil(t, orders)
		assert.NotNil(t, resultMeta)
		assert.Equal(t, len(expectedResponse), len(*orders))
		mockOrderRepo.AssertExpectations(t)
	})

	t.Run("error", func(t *testing.T) {
		params := &model.PaginationQuery{Page: 1, Limit: 10}
		mockOrderRepo.On("GetAll", params).Return(nil, nil, errors.New("error")).Once()

		orders, resultMeta, err := useCase.GetAllOrders(params)

		assert.Error(t, err)
		assert.Nil(t, orders)
		assert.Nil(t, resultMeta)
		mockOrderRepo.AssertExpectations(t)
	})
}

func TestOrderUseCase_GetCustomerOrders(t *testing.T) {
	logger := logrus.New()
	mockOrderRepo := new(MockOrderRepository)
	useCase := NewOrderUseCase(mockOrderRepo, nil, nil, logger, "test")

	t.Run("success", func(t *testing.T) {
		expectedResponse := []entity.Order{
			{
				ID: 1,
			},
		}
		mockOrderRepo.On("GetByCustomerID", int64(1)).Return(expectedResponse, nil).Once()

		orders, err := useCase.GetCustomerOrders(1)

		assert.NoError(t, err)
		assert.NotNil(t, orders)
		assert.Equal(t, len(expectedResponse), len(orders))
		mockOrderRepo.AssertExpectations(t)
	})

	t.Run("error", func(t *testing.T) {
		mockOrderRepo.On("GetByCustomerID", int64(1)).Return(nil, errors.New("error")).Once()

		orders, err := useCase.GetCustomerOrders(1)

		assert.Error(t, err)
		assert.Nil(t, orders)
		mockOrderRepo.AssertExpectations(t)
	})
}
