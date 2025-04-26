package usecase

import (
	"cakestore/internal/constants"
	"cakestore/internal/domain/entity"
	"cakestore/internal/domain/model"
	"cakestore/internal/repository"

	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
)

type WishListUseCase interface {
	CreateWishList(customerID, cakeID int64) error
	GetWishList(customerID int64, params *model.PaginationQuery) ([]model.CakeModel, *model.PaginatedMeta, error)
	DeleteWishList(customerID, cakeID int64) error
}

type wishListUseCase struct {
	wishListRepo repository.WishListRepository
	cakeRepo     repository.CakeRepository
	logger       *logrus.Logger
	validate     *validator.Validate
}

func NewWishListUseCase(
	wishListRepo repository.WishListRepository,
	cakeRepo repository.CakeRepository,
	logger *logrus.Logger,
) WishListUseCase {
	return &wishListUseCase{
		wishListRepo: wishListRepo,
		cakeRepo:     cakeRepo,
		logger:       logger,
		validate:     validator.New(),
	}
}

func (uc *wishListUseCase) CreateWishList(customerID, cakeID int64) error {
	wishlist := &entity.WishList{
		CustomerID: customerID,
		CakeID:     cakeID,
	}
	// if the cake is already in the wishlist, return an error
	_, err := uc.wishListRepo.GetByCustomerIDAndCakeID(customerID, cakeID)
	if err == nil {
		return constants.ErrCakeAlreadyInWishlist
	}

	return uc.wishListRepo.Create(wishlist)
}

func (u *wishListUseCase) GetWishList(customerID int64, params *model.PaginationQuery) ([]model.CakeModel, *model.PaginatedMeta, error) {
	cakes, meta, err := u.wishListRepo.GetByCustomerID(customerID, params)
	if err != nil {
		return nil, nil, err
	}

	var cakeResponses []model.CakeModel
	for _, cake := range cakes {
		cakeResponses = append(cakeResponses, model.CakeModel{
			ID:          cake.ID,
			Title:       cake.Title,
			Description: cake.Description,
			Price:       cake.Price,
			ImageURL:    cake.Image,
			Rating:      cake.Rating,
			Category:    cake.Category,
		})
	}

	return cakeResponses, meta, nil
}

func (uc *wishListUseCase) DeleteWishList(customerID, cakeID int64) error {
	return uc.wishListRepo.Delete(customerID, cakeID)
}
