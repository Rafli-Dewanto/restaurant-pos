package repository

import (
	"cakestore/internal/domain/entity"
	"cakestore/internal/domain/model"
	"errors"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type CartRepository interface {
	Create(cart *entity.Cart) error
	GetByID(id int) (*entity.Cart, error)
	GetByCustomerID(customerID int, params *model.PaginationQuery) (*model.PaginationResponse, error)
	AddItem(item *entity.CartItem) error
	UpdateItemQuantity(itemID int, quantity int) error
	RemoveItem(itemID int) error
	ClearCart(cartID int) error
}

type cartRepository struct {
	db     *gorm.DB
	logger *logrus.Logger
}

func NewCartRepository(db *gorm.DB, logger *logrus.Logger) CartRepository {
	return &cartRepository{
		db:     db,
		logger: logger,
	}
}

func (r *cartRepository) Create(cart *entity.Cart) error {
	cart.CreatedAt = time.Now()
	cart.UpdatedAt = time.Now()
	return r.db.Create(cart).Error
}

func (r *cartRepository) GetByID(id int) (*entity.Cart, error) {
	var cart entity.Cart
	err := r.db.Preload("Items").Where("deleted_at IS NULL").First(&cart, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("cart not found")
		}
		return nil, err
	}
	return &cart, nil
}

func (r *cartRepository) GetByCustomerID(customerID int, params *model.PaginationQuery) (*model.PaginationResponse, error) {
	var total int64
	var cart entity.Cart

	query := r.db.Model(&entity.Cart{}).Where("deleted_at IS NULL AND customer_id = ?", customerID).Preload("Items")

	if params != nil {
		if params.Page > 0 {
			query = query.Offset((params.Page - 1) * params.Limit)
		}
		if params.Limit > 0 {
			query = query.Limit(params.Limit)
		}
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	err := query.First(&cart).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("cart not found")
		}
		return nil, err
	}

	totalPages := int64(0)
	if params.Limit > 0 {
		totalPages = (total + int64(params.Limit) - 1) / int64(params.Limit)
	}

	return &model.PaginationResponse{
		Data:       cart.Items,
		Total:      total,
		Page:       params.Page,
		PageSize:   params.Limit,
		TotalPages: int(totalPages),
	}, nil
}

func (r *cartRepository) AddItem(item *entity.CartItem) error {
	return r.db.Create(item).Error
}

func (r *cartRepository) UpdateItemQuantity(itemID int, quantity int) error {
	result := r.db.Model(&entity.CartItem{}).Where("id = ?", itemID).Update("quantity", quantity)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("cart item not found")
	}
	return nil
}

func (r *cartRepository) RemoveItem(itemID int) error {
	result := r.db.Delete(&entity.CartItem{}, itemID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("cart item not found")
	}
	return nil
}

func (r *cartRepository) ClearCart(cartID int) error {
	return r.db.Where("cart_id = ?", cartID).Delete(&entity.CartItem{}).Error
}
