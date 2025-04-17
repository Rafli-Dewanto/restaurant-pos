package repository

import (
	"cakestore/internal/constants"
	"cakestore/internal/domain/entity"
	"errors"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type PaymentRepository interface {
	CreatePayment(payment *entity.Payment) error
	GetPaymentByOrderID(orderID int) (*entity.Payment, error)
	UpdatePayment(payment *entity.Payment) error
	// function to retrieve the first pending payment for testing purposes in development mode
	GetPendingPayment() (int, error)
}

type paymentRespositoryImpl struct {
	db  *gorm.DB
	log *logrus.Logger
}

func NewPaymentRepository(db *gorm.DB, log *logrus.Logger) PaymentRepository {
	return &paymentRespositoryImpl{
		db:  db,
		log: log,
	}
}

func (r *paymentRespositoryImpl) CreatePayment(payment *entity.Payment) error {
	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(payment).Error; err != nil {
			r.log.WithError(err).Error("Failed to create payment")
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (r *paymentRespositoryImpl) GetPaymentByOrderID(orderID int) (*entity.Payment, error) {
	var payment entity.Payment
	if err := r.db.Where("order_id = ?", orderID).First(&payment).Error; err != nil {
		r.log.WithError(err).Error("Failed to get payment")
		return nil, err
	}
	return &payment, nil
}

func (r *paymentRespositoryImpl) UpdatePayment(payment *entity.Payment) error {
	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&entity.Payment{}).
			Where("order_id = ?", payment.OrderID).
			Updates(map[string]interface{}{
				"status": payment.Status,
			}).Error; err != nil {
			r.log.WithError(err).Error("Failed to update payment")
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (r *paymentRespositoryImpl) GetPendingPayment() (int, error) {
	var payment entity.Payment
	if err := r.db.
		Preload("Order").
		Where("status = ?", constants.PaymentStatusPending).
		Order("created_at DESC").
		First(&payment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, errors.New("payment not found")
		}
		r.log.WithError(err).Error("Failed to get payment")
		return 0, err
	}
	return payment.ID, nil
}
