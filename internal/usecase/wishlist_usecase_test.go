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

type MockWishListRepository struct {
	mock.Mock
}

func (m *MockWishListRepository) Create(wishlist *entity.WishList) error {
	args := m.Called(wishlist)
	return args.Error(0)
}

func (m *MockWishListRepository) GetByCustomerID(customerID int64, params *model.PaginationQuery) ([]entity.Menu, *model.PaginatedMeta, error) {
	args := m.Called(customerID, params)
	if args.Get(0) == nil {
		return nil, nil, args.Error(2)
	}
	return args.Get(0).([]entity.Menu), args.Get(1).(*model.PaginatedMeta), args.Error(2)
}

func (m *MockWishListRepository) GetByCustomerIDAndMenuID(customerID, menuID int64) (*entity.WishList, error) {
	args := m.Called(customerID, menuID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.WishList), args.Error(1)
}

func (m *MockWishListRepository) Delete(customerID, menuID int64) error {
	args := m.Called(customerID, menuID)
	return args.Error(0)
}

func TestWishListUseCase_GetWishList(t *testing.T) {
	logger := logrus.New()
	mockWishListRepo := new(MockWishListRepository)
	mockCache := new(database.MockRedisCacheService)
	useCase := NewWishListUseCase(mockWishListRepo, nil, logger, mockCache)

	t.Run("success", func(t *testing.T) {
		expectedResponse := []entity.Menu{
			{
				ID:    1,
				Title: "Test Menu",
			},
		}
		meta := &model.PaginatedMeta{
			Total: 1,
		}
		params := &model.PaginationQuery{Page: 1, Limit: 10}
		mockCache.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("not found"))
		mockWishListRepo.On("GetByCustomerID", int64(1), params).Return(expectedResponse, meta, nil).Once()
		mockCache.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

		menus, resultMeta, err := useCase.GetWishList(1, params)

		assert.NoError(t, err)
		assert.NotNil(t, menus)
		assert.NotNil(t, resultMeta)
		assert.Equal(t, len(expectedResponse), len(menus))
		mockWishListRepo.AssertExpectations(t)
	})

	t.Run("error", func(t *testing.T) {
		params := &model.PaginationQuery{Page: 1, Limit: 10}
		mockCache.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("not found"))
		mockWishListRepo.On("GetByCustomerID", int64(1), params).Return(nil, nil, errors.New("error")).Once()

		menus, resultMeta, err := useCase.GetWishList(1, params)

		assert.Error(t, err)
		assert.Nil(t, menus)
		assert.Nil(t, resultMeta)
		mockWishListRepo.AssertExpectations(t)
	})
}
