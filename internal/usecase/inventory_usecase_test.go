package usecase

import (
	"cakestore/internal/database"
	"cakestore/internal/domain/entity"
	"cakestore/internal/domain/model"
	"errors"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockInventoryRepository struct {
	mock.Mock
}

func (m *MockInventoryRepository) Create(ingredient *entity.Inventory) error {
	args := m.Called(ingredient)
	return args.Error(0)
}

func (m *MockInventoryRepository) GetByID(id uint) (*entity.Inventory, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Inventory), args.Error(1)
}

func (m *MockInventoryRepository) GetAll(params *model.InventoryQueryParams) (*model.PaginationResponse[[]entity.Inventory], error) {
	args := m.Called(params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.PaginationResponse[[]entity.Inventory]), args.Error(1)
}

func (m *MockInventoryRepository) Update(ingredient *entity.Inventory) error {
	args := m.Called(ingredient)
	return args.Error(0)
}

func (m *MockInventoryRepository) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockInventoryRepository) UpdateStock(id uint, quantity float64) error {
	args := m.Called(id, quantity)
	return args.Error(0)
}

func (m *MockInventoryRepository) GetLowStockIngredients() ([]entity.Inventory, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.Inventory), args.Error(1)
}

func (m *MockInventoryRepository) Count() (int64, error) {
	args := m.Called()
	return args.Get(0).(int64), args.Error(1)
}

func TestInventoryUseCase_GetByID(t *testing.T) {
	logger := logrus.New()
	mockInventoryRepo := new(MockInventoryRepository)
	mockCache := new(database.MockRedisCacheService)
	useCase := NewInventoryUseCase(mockInventoryRepo, logger, mockCache)

	t.Run("success", func(t *testing.T) {
		expectedInventory := &entity.Inventory{
			ID:           1,
			Name:         "Flour",
			Quantity:     10,
			Unit:         "kg",
			MinimumStock: 5,
		}
		mockCache.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("not found"))
		mockInventoryRepo.On("GetByID", uint(1)).Return(expectedInventory, nil).Once()
		mockCache.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

		inventory, err := useCase.GetByID(1)

		assert.NoError(t, err)
		assert.NotNil(t, inventory)
		assert.Equal(t, expectedInventory.ID, inventory.ID)
		mockInventoryRepo.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		mockCache.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("not found"))
		mockInventoryRepo.On("GetByID", uint(1)).Return(nil, errors.New("not found")).Once()

		inventory, err := useCase.GetByID(1)

		assert.Error(t, err)
		assert.Nil(t, inventory)
		mockInventoryRepo.AssertExpectations(t)
	})
}

func TestInventoryUseCase_GetAll(t *testing.T) {
	logger := logrus.New()
	mockInventoryRepo := new(MockInventoryRepository)
	mockCache := new(database.MockRedisCacheService)
	useCase := NewInventoryUseCase(mockInventoryRepo, logger, mockCache)

	t.Run("success", func(t *testing.T) {
		expectedResponse := &model.PaginationResponse[[]entity.Inventory]{
			Data: []entity.Inventory{
				{
					ID:           1,
					Name:         "Flour",
					Quantity:     10,
					Unit:         "kg",
					MinimumStock: 5,
				},
			},
			Total: 1,
			Page:  1,
		}
		params := &model.InventoryQueryParams{Page: 1, Limit: 10}
		mockCache.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("not found"))
		mockInventoryRepo.On("GetAll", params).Return(expectedResponse, nil).Once()
		mockCache.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

		inventories, err := useCase.GetAll(params)

		assert.NoError(t, err)
		assert.NotNil(t, inventories)
		assert.Equal(t, len(expectedResponse.Data), len(inventories.Data))
		mockInventoryRepo.AssertExpectations(t)
	})

	t.Run("error", func(t *testing.T) {
		params := &model.InventoryQueryParams{Page: 1, Limit: 10}
		mockCache.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("not found"))
		mockInventoryRepo.On("GetAll", params).Return(nil, errors.New("error")).Once()

		inventories, err := useCase.GetAll(params)

		assert.Error(t, err)
		assert.Nil(t, inventories)
		mockInventoryRepo.AssertExpectations(t)
	})
}

func TestInventoryUseCase_GetLowStockIngredients(t *testing.T) {
	logger := logrus.New()
	mockInventoryRepo := new(MockInventoryRepository)
	mockCache := new(database.MockRedisCacheService)
	useCase := NewInventoryUseCase(mockInventoryRepo, logger, mockCache)

	t.Run("success", func(t *testing.T) {
		expectedIngredients := []entity.Inventory{
			{
				ID:           1,
				Name:         "Sugar",
				Quantity:     2,
				Unit:         "kg",
				MinimumStock: 5,
			},
		}
		mockCache.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("not found"))
		mockInventoryRepo.On("GetLowStockIngredients").Return(expectedIngredients, nil).Once()
		mockCache.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

		ingredients, err := useCase.GetLowStockIngredients()

		assert.NoError(t, err)
		assert.NotNil(t, ingredients)
		assert.Equal(t, len(expectedIngredients), len(ingredients))
		mockInventoryRepo.AssertExpectations(t)
	})

	t.Run("error", func(t *testing.T) {
		mockCache.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("not found"))
		mockInventoryRepo.On("GetLowStockIngredients").Return(nil, errors.New("error")).Once()

		ingredients, err := useCase.GetLowStockIngredients()

		assert.Error(t, err)
		assert.Nil(t, ingredients)
		mockInventoryRepo.AssertExpectations(t)
	})
}
