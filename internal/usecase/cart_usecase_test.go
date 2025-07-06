package usecase

import (
	"cakestore/internal/database"
	"cakestore/internal/domain/entity"
	"cakestore/internal/domain/model"
	"errors"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockCartRepository struct {
	mock.Mock
}

func (m *MockCartRepository) Create(cart *entity.Cart) error {
	args := m.Called(cart)
	return args.Error(0)
}

func (m *MockCartRepository) GetByID(id int64) (*entity.Cart, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Cart), args.Error(1)
}

func (m *MockCartRepository) GetByCustomerID(customerID int64, params *model.PaginationQuery) (*model.PaginationResponse[[]model.UserCartResponse], error) {
	args := m.Called(customerID, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.PaginationResponse[[]model.UserCartResponse]), args.Error(1)
}

func (m *MockCartRepository) GetByCustomerIDAndMenuID(customerID, menuID int64) (*entity.Cart, error) {
	args := m.Called(customerID, menuID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Cart), args.Error(1)
}

func (m *MockCartRepository) Update(cart *entity.Cart) error {
	args := m.Called(cart)
	return args.Error(0)
}

func (m *MockCartRepository) RemoveItem(customerID, cartID int64) error {
	args := m.Called(customerID, cartID)
	return args.Error(0)
}

func (m *MockCartRepository) ClearCustomerCart(customerID int64) error {
	args := m.Called(customerID)
	return args.Error(0)
}

func (m *MockCartRepository) BulkDelete(customerID int64, cartIDs []int64) error {
	args := m.Called(customerID, cartIDs)
	return args.Error(0)
}

func (m *MockCartRepository) Delete(id int64) error {
	args := m.Called(id)
	return args.Error(0)
}

func TestCartUseCase_GetCartByID(t *testing.T) {
	logger := logrus.New()
	mockCartRepo := new(MockCartRepository)
	mockCache := new(database.MockRedisCacheService)
	useCase := NewCartUseCase(mockCartRepo, nil, logger, mockCache)

	t.Run("success", func(t *testing.T) {
		expectedCart := &entity.Cart{
			ID:         1,
			CustomerID: 1,
			MenuID:     1,
			Quantity:   1,
			Price:      10000,
			Subtotal:   10000,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
		mockCache.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("not found"))
		mockCartRepo.On("GetByID", int64(1)).Return(expectedCart, nil).Once()
		mockCache.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

		cart, err := useCase.GetCartByID(1)

		assert.NoError(t, err)
		assert.NotNil(t, cart)
		assert.Equal(t, expectedCart.ID, cart.ID)
		mockCartRepo.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		mockCache.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("not found"))
		mockCartRepo.On("GetByID", int64(1)).Return(nil, errors.New("not found")).Once()

		cart, err := useCase.GetCartByID(1)

		assert.Error(t, err)
		assert.Nil(t, cart)
		mockCartRepo.AssertExpectations(t)
	})
}

func TestCartUseCase_GetCartByCustomerID(t *testing.T) {
	logger := logrus.New()
	mockCartRepo := new(MockCartRepository)
	mockCache := new(database.MockRedisCacheService)
	useCase := NewCartUseCase(mockCartRepo, nil, logger, mockCache)

	t.Run("success", func(t *testing.T) {
		expectedResponse := &model.PaginationResponse[[]model.UserCartResponse]{
			Data: []model.UserCartResponse{
				{
					ID:       1,
					MenuID:   1,
					Quantity: 1,
					Price:    10000,
					Subtotal: 10000,
				},
			},
			Total: 1,
			Page:  1,
		}
		params := &model.PaginationQuery{Page: 1, Limit: 10}
		mockCache.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("not found"))
		mockCartRepo.On("GetByCustomerID", int64(1), params).Return(expectedResponse, nil).Once()
		mockCache.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

		carts, meta, err := useCase.GetCartByCustomerID(1, params)

		assert.NoError(t, err)
		assert.NotNil(t, carts)
		assert.NotNil(t, meta)
		assert.Equal(t, len(expectedResponse.Data), len(carts))
		mockCartRepo.AssertExpectations(t)
	})

	t.Run("error", func(t *testing.T) {
		params := &model.PaginationQuery{Page: 1, Limit: 10}
		mockCache.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("not found"))
		mockCartRepo.On("GetByCustomerID", int64(1), params).Return(nil, errors.New("error")).Once()

		carts, meta, err := useCase.GetCartByCustomerID(1, params)

		assert.Error(t, err)
		assert.Nil(t, carts)
		assert.Nil(t, meta)
		mockCartRepo.AssertExpectations(t)
	})
}
