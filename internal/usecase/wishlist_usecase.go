package usecase

import (
	"cakestore/internal/constants"
	"cakestore/internal/database"
	"cakestore/internal/domain/entity"
	"cakestore/internal/domain/model"
	"cakestore/internal/repository"
	"context"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
)

type WishListUseCase interface {
	CreateWishList(customerID, menuID int64) error
	GetWishList(customerID int64, params *model.PaginationQuery) ([]model.MenuModel, *model.PaginatedMeta, error)
	DeleteWishList(customerID, menuID int64) error
}

type wishListUseCase struct {
	wishListRepo repository.WishListRepository
	menuRepo     repository.MenuRepository
	logger       *logrus.Logger
	validate     *validator.Validate
	cache        database.RedisCache
}

func NewWishListUseCase(
	wishListRepo repository.WishListRepository,
	menuRepo repository.MenuRepository,
	logger *logrus.Logger,
	cache database.RedisCache,
) WishListUseCase {
	return &wishListUseCase{
		wishListRepo: wishListRepo,
		menuRepo:     menuRepo,
		logger:       logger,
		validate:     validator.New(),
		cache:        cache,
	}
}

func (uc *wishListUseCase) CreateWishList(customerID, menuID int64) error {
	wishlist := &entity.WishList{
		CustomerID: customerID,
		MenuID:     menuID,
	}
	// if the menu is already in the wishlist, return an error
	_, err := uc.wishListRepo.GetByCustomerIDAndMenuID(customerID, menuID)
	if err == nil {
		return constants.ErrMenuAlreadyInWishlist
	}

	if err := uc.wishListRepo.Create(wishlist); err != nil {
		return err
	}

	// Invalidate cache
	cacheKey := fmt.Sprintf("wishlist:%d", customerID)
	if err := uc.cache.Delete(context.Background(), cacheKey); err != nil {
		uc.logger.Errorf("Error deleting cache for wishlist of customer ID %d: %v", customerID, err)
	}

	return nil
}

func (uc *wishListUseCase) GetWishList(customerID int64, params *model.PaginationQuery) ([]model.MenuModel, *model.PaginatedMeta, error) {
	start := time.Now()
	defer func() {
		uc.logger.Infof("GetWishList took %v", time.Since(start))
	}()

	// Try to get the wishlist from the cache first
	cacheKey := fmt.Sprintf("wishlist:%d:page:%d:limit:%d", customerID, params.Page, params.Limit)
	var cachedData struct {
		Data []model.MenuModel
		Meta *model.PaginatedMeta
	}
	if err := uc.cache.Get(context.Background(), cacheKey, &cachedData); err == nil {
		uc.logger.Info("Wishlist fetched from cache")
		return cachedData.Data, cachedData.Meta, nil
	}

	// If not in cache, get from the database
	menus, meta, err := uc.wishListRepo.GetByCustomerID(customerID, params)
	if err != nil {
		return nil, nil, err
	}

	var menuResponses []model.MenuModel
	for _, m := range menus {
		menuResponses = append(menuResponses, model.MenuModel{
			ID:          m.ID,
			Title:       m.Title,
			Description: m.Description,
			Price:       m.Price,
			ImageURL:    m.Image,
			Rating:      m.Rating,
			Category:    m.Category,
		})
	}

	// Store the wishlist in the cache for future requests
	if err := uc.cache.Set(context.Background(), cacheKey, struct {
		Data []model.MenuModel
		Meta *model.PaginatedMeta
	}{Data: menuResponses, Meta: meta}, 5*time.Minute); err != nil {
		uc.logger.Errorf("Error setting cache for wishlist of customer ID %d: %v", customerID, err)
	}

	return menuResponses, meta, nil
}

func (uc *wishListUseCase) DeleteWishList(customerID, menuID int64) error {
	if err := uc.wishListRepo.Delete(customerID, menuID); err != nil {
		return err
	}

	// Invalidate cache
	cacheKey := fmt.Sprintf("wishlist:%d", customerID)
	if err := uc.cache.Delete(context.Background(), cacheKey); err != nil {
		uc.logger.Errorf("Error deleting cache for wishlist of customer ID %d: %v", customerID, err)
	}

	return nil
}
