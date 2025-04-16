package repository

import (
	"cakestore/internal/domain/entity"
	"errors"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type CustomerRepository interface {
	Create(customer *entity.Customer) error
	GetByID(id int) (*entity.Customer, error)
	GetByEmail(email string) (*entity.Customer, error)
	Update(customer *entity.Customer) error
	Delete(id int) error
}

type customerRepository struct {
	db     *gorm.DB
	logger *logrus.Logger
}

func NewCustomerRepository(db *gorm.DB, logger *logrus.Logger) CustomerRepository {
	return &customerRepository{
		db:     db,
		logger: logger,
	}
}

func (r *customerRepository) Create(customer *entity.Customer) error {
	if err := r.db.Create(customer).Error; err != nil {
		r.logger.Errorf("Error creating customer: %v", err)
		return err
	}
	return nil
}

func (r *customerRepository) GetByID(id int) (*entity.Customer, error) {
	var customer entity.Customer
	if err := r.db.First(&customer, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("customer not found")
		}
		r.logger.Errorf("Error getting customer by ID: %v", err)
		return nil, err
	}
	return &customer, nil
}

func (r *customerRepository) GetByEmail(email string) (*entity.Customer, error) {
	var customer entity.Customer
	if err := r.db.Where("email = ?", email).First(&customer).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("customer not found")
		}
		r.logger.Errorf("Error getting customer by email: %v", err)
		return nil, err
	}
	return &customer, nil
}

func (r *customerRepository) Update(customer *entity.Customer) error {
	if err := r.db.Save(customer).Error; err != nil {
		r.logger.Errorf("Error updating customer: %v", err)
		return err
	}
	return nil
}

func (r *customerRepository) Delete(id int) error {
	result := r.db.Delete(&entity.Customer{}, id)
	if result.Error != nil {
		r.logger.Errorf("Error deleting customer: %v", result.Error)
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("customer not found")
	}
	return nil
}
