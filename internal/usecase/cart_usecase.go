package usecase

import (
	"cakestore/internal/constants"
	"cakestore/internal/domain/entity"
	"cakestore/internal/domain/model"
	"cakestore/internal/repository"
	"errors"

	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
)

type CartUseCase interface {
	CreateCart(customerID int) (*entity.Cart, error)
	GetCartByID(id int) (*entity.Cart, error)
	GetCartByCustomerID(customerID int, params *model.PaginationQuery) (*model.PaginationResponse, error)
	AddItem(cartID int, cakeID int, quantity int) error
	UpdateItemQuantity(cartID int, itemID int, quantity int) error
	RemoveItem(cartID int, itemID int) error
	ClearCart(cartID int) error
}

type cartUseCase struct {
	cartRepo repository.CartRepository
	cakeRepo repository.CakeRepository
	logger   *logrus.Logger
	validate *validator.Validate
}

func NewCartUseCase(
	cartRepo repository.CartRepository,
	cakeRepo repository.CakeRepository,
	logger *logrus.Logger,
) CartUseCase {
	return &cartUseCase{
		cartRepo: cartRepo,
		cakeRepo: cakeRepo,
		logger:   logger,
		validate: validator.New(),
	}
}

func (uc *cartUseCase) CreateCart(customerID int) (*entity.Cart, error) {
	cart := &entity.Cart{
		CustomerID: customerID,
		Total:      0,
	}

	if err := uc.cartRepo.Create(cart); err != nil {
		uc.logger.Errorf("Error creating cart for customer %d: %v", customerID, err)
		return nil, err
	}

	return cart, nil
}

func (uc *cartUseCase) GetCartByID(id int) (*entity.Cart, error) {
	cart, err := uc.cartRepo.GetByID(id)
	if err != nil {
		uc.logger.Errorf("Error getting cart by ID %d: %v", id, err)
		return nil, err
	}
	return cart, nil
}

func (uc *cartUseCase) GetCartByCustomerID(customerID int, params *model.PaginationQuery) (*model.PaginationResponse, error) {
	if params == nil {
		params = &model.PaginationQuery{}
	}
	cart, err := uc.cartRepo.GetByCustomerID(customerID, params)
	if err != nil {
		// If cart doesn't exist, create a new one
		if errors.Is(err, constants.ErrNotFound) {
			newCart, createErr := uc.CreateCart(customerID)
			if createErr != nil {
				uc.logger.Errorf("Error creating cart for customer %d: %v", customerID, createErr)
				return nil, createErr
			}
			// Convert single cart to pagination response
			return &model.PaginationResponse{
				Data:       []interface{}{newCart},
				Total:      1,
				Page:       1,
				TotalPages: 1,
			}, nil
		}
		uc.logger.Errorf("Error getting cart for customer %d: %v", customerID, err)
		return nil, err
	}
	return cart, nil
}

func (uc *cartUseCase) AddItem(cartID int, cakeID int, quantity int) error {
	if quantity <= 0 {
		return errors.New("quantity must be greater than 0")
	}

	// Get cake details
	cake, err := uc.cakeRepo.GetByID(cakeID)
	if err != nil {
		uc.logger.Errorf("Error getting cake details: %v", err)
		return err
	}

	// Create cart item
	cartItem := &entity.CartItem{
		CartID:   cartID,
		CakeID:   cakeID,
		Quantity: quantity,
		Price:    cake.Price,
		Subtotal: cake.Price * float64(quantity),
	}

	if err := uc.cartRepo.AddItem(cartItem); err != nil {
		uc.logger.Errorf("Error adding item to cart: %v", err)
		return err
	}

	return nil
}

func (uc *cartUseCase) UpdateItemQuantity(cartID int, itemID int, quantity int) error {
	if quantity <= 0 {
		return errors.New("quantity must be greater than 0")
	}

	if err := uc.cartRepo.UpdateItemQuantity(itemID, quantity); err != nil {
		uc.logger.Errorf("Error updating item quantity: %v", err)
		return err
	}

	return nil
}

func (uc *cartUseCase) RemoveItem(cartID int, itemID int) error {
	if err := uc.cartRepo.RemoveItem(itemID); err != nil {
		uc.logger.Errorf("Error removing item from cart: %v", err)
		return err
	}

	return nil
}

func (uc *cartUseCase) ClearCart(cartID int) error {
	if err := uc.cartRepo.ClearCart(cartID); err != nil {
		uc.logger.Errorf("Error clearing cart: %v", err)
		return err
	}

	return nil
}
