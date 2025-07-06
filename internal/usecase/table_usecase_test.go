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

type MockTableRepository struct {
	mock.Mock
}

func (m *MockTableRepository) Create(table *entity.Table) error {
	args := m.Called(table)
	return args.Error(0)
}

func (m *MockTableRepository) GetByID(id uint) (*entity.Table, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Table), args.Error(1)
}

func (m *MockTableRepository) GetAll() ([]entity.Table, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.Table), args.Error(1)
}

func (m *MockTableRepository) Update(table *entity.Table) error {
	args := m.Called(table)
	return args.Error(0)
}

func (m *MockTableRepository) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockTableRepository) GetAvailableTables(reserveTime time.Time, duration time.Duration) ([]entity.Table, error) {
	args := m.Called(reserveTime, duration)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.Table), args.Error(1)
}

func (m *MockTableRepository) UpdateAvailability(id uint, isAvailable bool) error {
	args := m.Called(id, isAvailable)
	return args.Error(0)
}

func (m *MockTableRepository) Count() (int64, error) {
	args := m.Called()
	return args.Get(0).(int64), args.Error(1)
}

func TestTableUseCase_GetByID(t *testing.T) {
	logger := logrus.New()
	mockTableRepo := new(MockTableRepository)
	mockCache := new(database.MockRedisCacheService)
	useCase := NewTableUseCase(mockTableRepo, logger, mockCache)

	t.Run("success", func(t *testing.T) {
		expectedTable := &entity.Table{
			ID: 1,
		}
		mockCache.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("not found"))
		mockTableRepo.On("GetByID", uint(1)).Return(expectedTable, nil).Once()
		mockCache.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

		table, err := useCase.GetByID(1)

		assert.NoError(t, err)
		assert.NotNil(t, table)
		assert.Equal(t, uint(expectedTable.ID), table.ID)
		mockTableRepo.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		mockCache.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("not found"))
		mockTableRepo.On("GetByID", uint(1)).Return(nil, errors.New("not found")).Once()

		table, err := useCase.GetByID(1)

		assert.Error(t, err)
		assert.Nil(t, table)
		mockTableRepo.AssertExpectations(t)
	})
}

func TestTableUseCase_GetAll(t *testing.T) {
	logger := logrus.New()
	mockTableRepo := new(MockTableRepository)
	mockCache := new(database.MockRedisCacheService)
	useCase := NewTableUseCase(mockTableRepo, logger, mockCache)

	t.Run("success", func(t *testing.T) {
		expectedResponse := []entity.Table{
			{
				ID: 1,
			},
		}
		params := &model.TableQueryParams{}
		mockCache.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("not found"))
		mockTableRepo.On("GetAll").Return(expectedResponse, nil).Once()
		mockCache.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

		tables, err := useCase.GetAll(params)

		assert.NoError(t, err)
		assert.NotNil(t, tables)
		assert.Equal(t, len(expectedResponse), len(tables.Data))
		mockTableRepo.AssertExpectations(t)
	})

	t.Run("error", func(t *testing.T) {
		params := &model.TableQueryParams{}
		mockCache.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("not found"))
		mockTableRepo.On("GetAll").Return(nil, errors.New("error")).Once()

		tables, err := useCase.GetAll(params)

		assert.Error(t, err)
		assert.Nil(t, tables)
		mockTableRepo.AssertExpectations(t)
	})
}

func TestTableUseCase_GetAvailableTables(t *testing.T) {
	logger := logrus.New()
	mockTableRepo := new(MockTableRepository)
	mockCache := new(database.MockRedisCacheService)
	useCase := NewTableUseCase(mockTableRepo, logger, mockCache)

	t.Run("success", func(t *testing.T) {
		expectedResponse := []entity.Table{
			{
				ID: 1,
			},
		}
		reserveTime := time.Now()
		duration := time.Hour
		mockCache.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("not found"))
		mockTableRepo.On("GetAvailableTables", reserveTime, duration).Return(expectedResponse, nil).Once()
		mockCache.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

		tables, err := useCase.GetAvailableTables(reserveTime, duration)

		assert.NoError(t, err)
		assert.NotNil(t, tables)
		assert.Equal(t, len(expectedResponse), len(tables))
		mockTableRepo.AssertExpectations(t)
	})

	t.Run("error", func(t *testing.T) {
		reserveTime := time.Now()
		duration := time.Hour
		mockCache.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("not found"))
		mockTableRepo.On("GetAvailableTables", reserveTime, duration).Return(nil, errors.New("error")).Once()

		tables, err := useCase.GetAvailableTables(reserveTime, duration)

		assert.Error(t, err)
		assert.Nil(t, tables)
		mockTableRepo.AssertExpectations(t)
	})
}
