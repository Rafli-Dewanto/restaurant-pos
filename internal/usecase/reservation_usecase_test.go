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

type MockReservationRepository struct {
	mock.Mock
}

func (m *MockReservationRepository) Create(reservation *entity.Reservation) error {
	args := m.Called(reservation)
	return args.Error(0)
}

func (m *MockReservationRepository) GetByID(id uint) (*entity.Reservation, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Reservation), args.Error(1)
}

func (m *MockReservationRepository) GetAll(params *model.ReservationQueryParams) (*model.PaginationResponse[[]entity.Reservation], error) {
	args := m.Called(params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.PaginationResponse[[]entity.Reservation]), args.Error(1)
}

func (m *MockReservationRepository) AdminGetAllCustomerReservations(params *model.PaginationQuery) (*model.PaginationResponse[[]entity.Reservation], error) {
	args := m.Called(params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.PaginationResponse[[]entity.Reservation]), args.Error(1)
}

func (m *MockReservationRepository) Update(reservation *entity.Reservation) error {
	args := m.Called(reservation)
	return args.Error(0)
}

func (m *MockReservationRepository) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockReservationRepository) CheckTableAvailability(tableID uint, reserveDate time.Time) (bool, error) {
	args := m.Called(tableID, reserveDate)
	return args.Bool(0), args.Error(1)
}

func TestReservationUseCase_GetByID(t *testing.T) {
	logger := logrus.New()
	mockReservationRepo := new(MockReservationRepository)
	mockCache := new(database.MockRedisCacheService)
	useCase := NewReservationUseCase(mockReservationRepo, logger, nil, mockCache)

	t.Run("success", func(t *testing.T) {
		expectedReservation := &entity.Reservation{
			ID: 1,
		}
		mockCache.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("not found"))
		mockReservationRepo.On("GetByID", uint(1)).Return(expectedReservation, nil).Once()
		mockCache.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

		reservation, err := useCase.GetByID(1)

		assert.NoError(t, err)
		assert.NotNil(t, reservation)
		assert.Equal(t, expectedReservation.ID, reservation.ID)
		mockReservationRepo.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		mockCache.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("not found"))
		mockReservationRepo.On("GetByID", uint(1)).Return(nil, errors.New("not found")).Once()

		reservation, err := useCase.GetByID(1)

		assert.Error(t, err)
		assert.Nil(t, reservation)
		mockReservationRepo.AssertExpectations(t)
	})
}

func TestReservationUseCase_GetAll(t *testing.T) {
	logger := logrus.New()
	mockReservationRepo := new(MockReservationRepository)
	mockCache := new(database.MockRedisCacheService)
	useCase := NewReservationUseCase(mockReservationRepo, logger, nil, mockCache)

	t.Run("success", func(t *testing.T) {
		expectedResponse := &model.PaginationResponse[[]entity.Reservation]{
			Data: []entity.Reservation{
				{
					ID: 1,
				},
			},
			Total: 1,
		}
		params := &model.ReservationQueryParams{}
		mockCache.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("not found"))
		mockReservationRepo.On("GetAll", params).Return(expectedResponse, nil).Once()
		mockCache.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

		reservations, err := useCase.GetAll(params)

		assert.NoError(t, err)
		assert.NotNil(t, reservations)
		assert.Equal(t, len(expectedResponse.Data), len(reservations.Data))
		mockReservationRepo.AssertExpectations(t)
	})

	t.Run("error", func(t *testing.T) {
		params := &model.ReservationQueryParams{}
		mockCache.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("not found"))
		mockReservationRepo.On("GetAll", params).Return(nil, errors.New("error")).Once()

		reservations, err := useCase.GetAll(params)

		assert.Error(t, err)
		assert.Nil(t, reservations)
		mockReservationRepo.AssertExpectations(t)
	})
}

func TestReservationUseCase_AdminGetAllCustomerReservations(t *testing.T) {
	logger := logrus.New()
	mockReservationRepo := new(MockReservationRepository)
	mockCache := new(database.MockRedisCacheService)
	useCase := NewReservationUseCase(mockReservationRepo, logger, nil, mockCache)

	t.Run("success", func(t *testing.T) {
		expectedResponse := &model.PaginationResponse[[]entity.Reservation]{
			Data: []entity.Reservation{
				{
					ID: 1,
				},
			},
			Total: 1,
		}
		params := &model.PaginationQuery{}
		mockCache.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("not found"))
		mockReservationRepo.On("AdminGetAllCustomerReservations", params).Return(expectedResponse, nil).Once()
		mockCache.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

		reservations, err := useCase.AdminGetAllCustomerReservations(params)

		assert.NoError(t, err)
		assert.NotNil(t, reservations)
		assert.Equal(t, len(expectedResponse.Data), len(reservations.Data))
		mockReservationRepo.AssertExpectations(t)
	})

	t.Run("error", func(t *testing.T) {
		params := &model.PaginationQuery{}
		mockCache.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("not found"))
		mockReservationRepo.On("AdminGetAllCustomerReservations", params).Return(nil, errors.New("error")).Once()

		reservations, err := useCase.AdminGetAllCustomerReservations(params)

		assert.Error(t, err)
		assert.Nil(t, reservations)
		mockReservationRepo.AssertExpectations(t)
	})
}
