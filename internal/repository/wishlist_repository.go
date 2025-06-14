package repository

import (
	"cakestore/internal/domain/entity"
	"cakestore/internal/domain/model"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type WishListRepository interface {
	Create(wishlist *entity.WishList) error
	GetByCustomerID(customerID int64, params *model.PaginationQuery) ([]entity.Menu, *model.PaginatedMeta, error)
	Delete(customerID, menuID int64) error
	GetByCustomerIDAndMenuID(customerID, menuID int64) (*entity.WishList, error)
}

type wishListRepository struct {
	db     *gorm.DB
	logger *logrus.Logger
}

func NewWishListRepository(db *gorm.DB, logger *logrus.Logger) WishListRepository {
	return &wishListRepository{
		db:     db,
		logger: logger,
	}
}

func (r *wishListRepository) GetByCustomerIDAndMenuID(customerID, menuID int64) (*entity.WishList, error) {
	var wishlist entity.WishList
	if err := r.db.Where("customer_id = ? AND menu_id = ? AND deleted_at IS NULL", customerID, menuID).First(&wishlist).Error; err != nil {
		return nil, err
	}
	return &wishlist, nil
}

func (r *wishListRepository) Create(wishlist *entity.WishList) error {
	return r.db.Create(wishlist).Error
}

func (r *wishListRepository) GetByCustomerID(customerID int64, params *model.PaginationQuery) ([]entity.Menu, *model.PaginatedMeta, error) {
	var menus []entity.Menu
	var total int64

	if err := r.db.
		Table("menus").
		Joins("JOIN wishlists ON wishlists.menu_id = menus.id").
		Where("wishlists.customer_id = ? AND wishlists.deleted_at IS NULL", customerID).
		Count(&total).Error; err != nil {
		return nil, nil, err
	}

	offset := (params.Page - 1) * params.Limit

	if err := r.db.
		Table("menus").
		Joins("JOIN wishlists ON wishlists.menu_id = menus.id").
		Where("wishlists.customer_id = ? AND wishlists.deleted_at IS NULL", customerID).
		Limit(int(params.Limit)).
		Offset(int(offset)).
		Find(&menus).Error; err != nil {
		return nil, nil, err
	}

	lastPage := int64((int64(params.Limit) + params.Limit - 1) / params.Limit)
	meta := &model.PaginatedMeta{
		CurrentPage: int64(params.Page),
		Total:       total,
		PerPage:     int64(params.Limit),
		LastPage:    lastPage,
		HasNextPage: params.Page < int64(lastPage),
		HasPrevPage: params.Page > 1,
	}

	return menus, meta, nil
}

func (r *wishListRepository) Delete(customerID, menuID int64) error {
	return r.db.Model(&entity.WishList{}).
		Where("customer_id = ? AND menu_id = ?", customerID, menuID).
		Updates(map[string]interface{}{
			"deleted_at": gorm.Expr("NOW()"),
		}).Error
}
