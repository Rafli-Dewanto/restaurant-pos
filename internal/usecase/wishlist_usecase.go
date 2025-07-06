package usecase

import (
	"cakestore/internal/constants"
	"cakestore/internal/domain/entity"
	"cakestore/internal/domain/model"
	"cakestore/internal/repository"
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
}

func NewWishListUseCase(
	wishListRepo repository.WishListRepository,
	menuRepo repository.MenuRepository,
	logger *logrus.Logger,
) WishListUseCase {
	return &wishListUseCase{
		wishListRepo: wishListRepo,
		menuRepo:     menuRepo,
		logger:       logger,
		validate:     validator.New(),
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

	return uc.wishListRepo.Create(wishlist)
}

func (uc *wishListUseCase) GetWishList(customerID int64, params *model.PaginationQuery) ([]model.MenuModel, *model.PaginatedMeta, error) {
	start := time.Now()
	defer func() {
		uc.logger.Infof("GetWishList took %v", time.Since(start))
	}()

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

	return menuResponses, meta, nil
}

func (uc *wishListUseCase) DeleteWishList(customerID, menuID int64) error {
	return uc.wishListRepo.Delete(customerID, menuID)
}
