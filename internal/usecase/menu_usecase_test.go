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

type MockMenuRepository struct {
	mock.Mock
}

func (m *MockMenuRepository) GetAll(params *model.MenuQueryParams) (*model.PaginationResponse[[]entity.Menu], error) {
	args := m.Called(params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.PaginationResponse[[]entity.Menu]), args.Error(1)
}

func (m *MockMenuRepository) GetByID(id int64) (*entity.Menu, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Menu), args.Error(1)
}

func (m *MockMenuRepository) Create(menu *entity.Menu) error {
	args := m.Called(menu)
	return args.Error(0)
}

func (m *MockMenuRepository) UpdateMenu(menu *entity.Menu) error {
	args := m.Called(menu)
	return args.Error(0)
}

func (m *MockMenuRepository) SoftDelete(id int64) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockMenuRepository) DecreaseStock(menuID int64, quantity int) error {
	args := m.Called(menuID, quantity)
	return args.Error(0)
}

func TestMenuUseCase_GetAllMenus(t *testing.T) {
	logger := logrus.New()
	mockMenuRepo := new(MockMenuRepository)
	mockCache := new(database.MockRedisCacheService)
	useCase := NewMenuUseCase(mockMenuRepo, logger, mockCache)

	t.Run("success", func(t *testing.T) {
		expectedResponse := &model.PaginationResponse[[]entity.Menu]{
			Data: []entity.Menu{
				{
					ID:    1,
					Title: "Test Menu",
				},
			},
			Total: 1,
			Page:  1,
		}
		params := &model.MenuQueryParams{Page: 1, Limit: 10}
		mockCache.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("not found"))
		mockMenuRepo.On("GetAll", params).Return(expectedResponse, nil).Once()
		mockCache.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

		menus, err := useCase.GetAllMenus(params)

		assert.NoError(t, err)
		assert.NotNil(t, menus)
		assert.Equal(t, len(expectedResponse.Data), len(menus.Data))
		mockMenuRepo.AssertExpectations(t)
	})

	t.Run("error", func(t *testing.T) {
		params := &model.MenuQueryParams{Page: 1, Limit: 10}
		mockCache.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("not found"))
		mockMenuRepo.On("GetAll", params).Return(nil, errors.New("error")).Once()

		menus, err := useCase.GetAllMenus(params)

		assert.Error(t, err)
		assert.Nil(t, menus)
		mockMenuRepo.AssertExpectations(t)
	})
}

func TestMenuUseCase_GetMenuByID(t *testing.T) {
	logger := logrus.New()
	mockMenuRepo := new(MockMenuRepository)
	mockCache := new(database.MockRedisCacheService)
	useCase := NewMenuUseCase(mockMenuRepo, logger, mockCache)

	t.Run("success", func(t *testing.T) {
		expectedMenu := &entity.Menu{
			ID:    1,
			Title: "Test Menu",
		}
		mockCache.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("not found"))
		mockMenuRepo.On("GetByID", int64(1)).Return(expectedMenu, nil).Once()
		mockCache.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

		menu, err := useCase.GetMenuByID(1)

		assert.NoError(t, err)
		assert.NotNil(t, menu)
		assert.Equal(t, expectedMenu.ID, menu.ID)
		mockMenuRepo.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		mockCache.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("not found"))
		mockMenuRepo.On("GetByID", int64(1)).Return(nil, errors.New("not found")).Once()

		menu, err := useCase.GetMenuByID(1)

		assert.Error(t, err)
		assert.Nil(t, menu)
		mockMenuRepo.AssertExpectations(t)
	})
}
