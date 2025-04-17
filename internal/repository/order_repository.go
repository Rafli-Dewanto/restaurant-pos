package repository

import (
	"cakestore/internal/domain/entity"
	"errors"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type OrderRepository interface {
	Create(order *entity.Order) error
	GetByID(id int) (*entity.Order, error)
	GetByCustomerID(customerID int) ([]entity.Order, error)
	Update(order *entity.Order) error
	Delete(id int) error
	UpdateStatus(id int, status entity.OrderStatus) error
	// GetPendingOrder retrieves the first pending order from the database for testing purposes
	GetPendingOrder() (int, error)
}

type orderRepository struct {
	db     *gorm.DB
	logger *logrus.Logger
}

func NewOrderRepository(db *gorm.DB, logger *logrus.Logger) OrderRepository {
	return &orderRepository{
		db:     db,
		logger: logger,
	}
}

func (r *orderRepository) Create(order *entity.Order) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(order).Error; err != nil {
			r.logger.Errorf("Error creating order: %v", err)
			return err
		}
		return nil
	})
}

func (r *orderRepository) GetByID(id int) (*entity.Order, error) {
	var order entity.Order
	if err := r.db.Preload("Items.Cake").Preload("Customer").First(&order, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("order not found")
		}
		r.logger.Errorf("Error getting order by ID: %v", err)
		return nil, err
	}
	return &order, nil
}

func (r *orderRepository) GetByCustomerID(customerID int) ([]entity.Order, error) {
	var orders []entity.Order
	if err := r.db.Preload("Customer").Preload("Items.Cake").Where("customer_id = ?", customerID).Find(&orders).Error; err != nil {
		r.logger.Errorf("Error getting orders by customer ID: %v", err)
		return nil, err
	}
	return orders, nil
}

func (r *orderRepository) Update(order *entity.Order) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(order).Error; err != nil {
			r.logger.Errorf("Error updating order: %v", err)
			return err
		}

		for _, item := range order.Items {
			if err := tx.Save(&item).Error; err != nil {
				r.logger.Errorf("Error updating order item: %v", err)
				return err
			}
		}
		return nil
	})
}

func (r *orderRepository) Delete(id int) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("order_id = ?", id).Delete(&entity.OrderItem{}).Error; err != nil {
			r.logger.Errorf("Error deleting order items: %v", err)
			return err
		}

		result := tx.Delete(&entity.Order{}, id)
		if result.Error != nil {
			r.logger.Errorf("Error deleting order: %v", result.Error)
			return result.Error
		}
		if result.RowsAffected == 0 {
			return errors.New("order not found")
		}
		return nil
	})
}

func (r *orderRepository) UpdateStatus(id int, status entity.OrderStatus) error {
	result := r.db.Model(&entity.Order{}).Where("id = ?", id).Update("status", status)
	if result.Error != nil {
		r.logger.Errorf("UpdateStatus repository ~ Error updating order status: %v", result.Error)
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("order not found")
	}
	return nil
}

// GetPendingOrder retrieves the first pending order from the database for testing purposes
func (r *orderRepository) GetPendingOrder() (int, error) {
	var order entity.Order
	if err := r.db.Preload("Items.Cake").Preload("Customer").Where("status = ?", entity.OrderStatusPending).First(&order).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, errors.New("order not found")
		}
		r.logger.Errorf("GetPendingOrder repository ~ Error getting order: %v", err)
		return 0, err
	}
	return order.ID, nil
}
