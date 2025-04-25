package repository

import (
	"cakestore/internal/domain/entity"
	"cakestore/internal/domain/model"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type WishListRepository interface {
	Create(wishlist *entity.WishList) error
	GetByCustomerID(customerID int, params *model.PaginationQuery) ([]entity.Cake, *model.PaginatedMeta, error)
	Delete(customerID, cakeID int) error
	GetByCustomerIDAndCakeID(customerID, cakeID int) (*entity.WishList, error)
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

func (r *wishListRepository) GetByCustomerIDAndCakeID(customerID, cakeID int) (*entity.WishList, error) {
	var wishlist entity.WishList
	if err := r.db.Where("customer_id = ? AND cake_id = ?", customerID, cakeID).First(&wishlist).Error; err != nil {
		return nil, err
	}
	return &wishlist, nil
}

func (r *wishListRepository) Create(wishlist *entity.WishList) error {
	return r.db.Create(wishlist).Error
}

func (r *wishListRepository) GetByCustomerID(customerID int, params *model.PaginationQuery) ([]entity.Cake, *model.PaginatedMeta, error) {
	var cakes []entity.Cake
	var total int64

	if err := r.db.
		Table("cakes").
		Joins("JOIN wishlists ON wishlists.cake_id = cakes.id").
		Where("wishlists.customer_id = ? AND wishlists.deleted_at IS NULL", customerID).
		Count(&total).Error; err != nil {
		return nil, nil, err
	}

	offset := (params.Page - 1) * params.Limit

	if err := r.db.
		Table("cakes").
		Joins("JOIN wishlists ON wishlists.cake_id = cakes.id").
		Where("wishlists.customer_id = ? AND wishlists.deleted_at IS NULL", customerID).
		Limit(int(params.Limit)).
		Offset(int(offset)).
		Find(&cakes).Error; err != nil {
		return nil, nil, err
	}

	lastPage := int((int(params.Limit) + params.Limit - 1) / params.Limit)
	meta := &model.PaginatedMeta{
		CurrentPage: int64(params.Page),
		Total:       total,
		PerPage:     int64(params.Limit),
		LastPage:    lastPage,
		HasNextPage: params.Page < int(lastPage),
		HasPrevPage: params.Page > 1,
	}

	return cakes, meta, nil
}

func (r *wishListRepository) Delete(customerID, cakeID int) error {
	return r.db.
		Where("customer_id = ? AND cake_id = ?", customerID, cakeID).
		Delete(&entity.WishList{}).
		Error
}
