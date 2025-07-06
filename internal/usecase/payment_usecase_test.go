package usecase

import (
	"cakestore/internal/domain/entity"
	"errors"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockPaymentRepository struct {
	mock.Mock
}

func (m *MockPaymentRepository) CreatePayment(payment *entity.Payment) error {
	args := m.Called(payment)
	return args.Error(0)
}

func (m *MockPaymentRepository) GetPaymentByOrderID(orderID int64) (*entity.Payment, error) {
	args := m.Called(orderID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Payment), args.Error(1)
}

func (m *MockPaymentRepository) UpdatePayment(payment *entity.Payment) error {
	args := m.Called(payment)
	return args.Error(0)
}

func (m *MockPaymentRepository) GetPendingPayment() (int64, error) {
	args := m.Called()
	return args.Get(0).(int64), args.Error(1)
}

func TestPaymentUseCase_GetPaymentByOrderID(t *testing.T) {
	logger := logrus.New()
	mockPaymentRepo := new(MockPaymentRepository)
	useCase := NewPaymentUseCase("http://test.com", mockPaymentRepo, logger, "test")

	t.Run("success", func(t *testing.T) {
		expectedPayment := &entity.Payment{
			ID:      1,
			OrderID: 1,
		}
		order := &entity.Order{ID: 1}
		mockPaymentRepo.On("GetPaymentByOrderID", order.ID).Return(expectedPayment, nil).Once()

		payment, err := useCase.GetPaymentByOrderID(order)

		assert.NoError(t, err)
		assert.NotNil(t, payment)
		assert.Equal(t, expectedPayment.ID, payment.ID)
		mockPaymentRepo.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		order := &entity.Order{ID: 1}
		mockPaymentRepo.On("GetPaymentByOrderID", order.ID).Return(nil, errors.New("not found")).Once()

		payment, err := useCase.GetPaymentByOrderID(order)

		assert.Error(t, err)
		assert.Nil(t, payment)
		mockPaymentRepo.AssertExpectations(t)
	})
}
