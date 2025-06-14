package repository

import (
	"cakestore/internal/domain/entity"
	"cakestore/internal/domain/model"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type CartRepository interface {
	Create(cart *entity.Cart) error
	GetByID(id int64) (*entity.Cart, error)
	GetByCustomerID(customerID int64, params *model.PaginationQuery) (*model.PaginationResponse[[]model.UserCartResponse], error)
	GetByCustomerIDAndMenuID(customerID int64, menuID int64) (*entity.Cart, error)
	Update(cart *entity.Cart) error
	Delete(cartID int64) error
	RemoveItem(customerID int64, cartID int64) error
	ClearCustomerCart(customerID int64) error
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
	return r.db.Create(cart).Error
}

func (r *cartRepository) GetByID(id int64) (*entity.Cart, error) {
	var cart entity.Cart
	if err := r.db.First(&cart, id).Error; err != nil {
		return nil, err
	}
	return &cart, nil
}

func (r *cartRepository) GetByCustomerID(customerID int64, params *model.PaginationQuery) (*model.PaginationResponse[[]model.UserCartResponse], error) {
	var carts []model.UserCartResponse
	var total int64
	var perPage int64
	var page int64 = 1

	if params.Page > 0 {
		page = int64(params.Page)
	}

	query := r.db.Model(&entity.Cart{}).
		Select("carts.id, menus.title as menu_name, menus.image as menu_image, carts.customer_id, carts.menu_id, carts.quantity, carts.price, carts.subtotal, carts.created_at, carts.updated_at").
		Joins("JOIN menus ON menus.id = carts.menu_id").
		Where("carts.customer_id = ?", customerID)

	query.Count(&total)

	if params.Limit > 0 {
		perPage = int64(params.Limit)
	} else {
		perPage = 10
	}

	offset := (page - 1) * perPage

	if err := query.Offset(int(offset)).Limit(int(perPage)).Scan(&carts).Error; err != nil {
		return nil, err
	}

	return &model.PaginationResponse[[]model.UserCartResponse]{
		Total:      total,
		Data:       carts,
		Page:       page,
		PageSize:   perPage,
		TotalPages: (total + perPage - 1) / perPage, // better calculation for total pages
	}, nil
}

func (r *cartRepository) GetByCustomerIDAndMenuID(customerID int64, menuID int64) (*entity.Cart, error) {
	var cart entity.Cart
	if err := r.db.Where("customer_id = ? AND menu_id = ?", customerID, menuID).First(&cart).Error; err != nil {
		return nil, err
	}
	return &cart, nil
}

func (r *cartRepository) Update(cart *entity.Cart) error {
	return r.db.Save(cart).Error
}

func (r *cartRepository) Delete(cartID int64) error {
	return r.db.Delete(&entity.Cart{}, cartID).Error
}

func (r *cartRepository) RemoveItem(customerID int64, cartID int64) error {
	type result struct {
		Quantity int64
		Subtotal float64
	}

	var res result

	// Retrieve quantity and subtotal
	if err := r.db.
		Model(&entity.Cart{}).
		Where("id = ? AND customer_id = ?", cartID, customerID).
		Select("quantity", "subtotal").
		Scan(&res).Error; err != nil {
		return err
	}

	// If only 1 item, delete the cart item
	if res.Quantity <= 1 {
		if err := r.db.
			Where("id = ? AND customer_id = ?", cartID, customerID).
			Delete(&entity.Cart{}).Error; err != nil {
			return err
		}
	} else {
		// Update: decrement quantity and subtotal
		newSubtotal := res.Subtotal / float64(res.Quantity)
		if err := r.db.
			Model(&entity.Cart{}).
			Where("id = ? AND customer_id = ?", cartID, customerID).
			Updates(map[string]interface{}{
				"quantity": gorm.Expr("quantity - ?", 1),
				"subtotal": gorm.Expr("subtotal - ?", newSubtotal),
			}).Error; err != nil {
			return err
		}
	}

	return nil
}

func (r *cartRepository) ClearCustomerCart(customerID int64) error {
	result := r.db.Where("customer_id = ?", customerID).Delete(&entity.Cart{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
