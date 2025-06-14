package usecase

import (
	"cakestore/internal/constants"
	"cakestore/internal/domain/model"
	"cakestore/internal/repository"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
)

type CartUseCase interface {
	CreateCart(customerID int64, req *model.AddCart) error
	GetCartByID(id int64) (*model.CartModel, error)
	GetCartByCustomerID(customerID int64, params *model.PaginationQuery) ([]model.UserCartResponse, *model.PaginatedMeta, error)
	RemoveCart(customerID int64, cartID int64) error
	ClearCart(cartID int64) error
	BulkDeleteCart(customerID int64, cartIDs []int64) error
}

type cartUseCase struct {
	cartRepo repository.CartRepository
	menuRepo repository.MenuRepository
	logger   *logrus.Logger
	validate *validator.Validate
}

func NewCartUseCase(
	cartRepo repository.CartRepository,
	menuRepo repository.MenuRepository,
	logger *logrus.Logger,
) CartUseCase {
	return &cartUseCase{
		cartRepo: cartRepo,
		menuRepo: menuRepo,
		logger:   logger,
		validate: validator.New(),
	}
}

func (uc *cartUseCase) CreateCart(customerID int64, req *model.AddCart) error {
	if err := uc.validate.Struct(req); err != nil {
		uc.logger.Errorf("Validation failed for request: %v", err)
		return err
	}

	menu, err := uc.menuRepo.GetByID(req.MenuID)
	if err != nil {
		uc.logger.Errorf("Error getting menu with ID %d: %v", req.MenuID, err)
		return err
	}

	// check if customer already have the same menu added, if so update the quantity
	cart, err := uc.cartRepo.GetByCustomerIDAndMenuID(customerID, req.MenuID)
	// if not, create a new cart
	if err != nil {
		cartModel := &model.CartModel{
			CustomerID: customerID,
			MenuID:     req.MenuID,
			Quantity:   req.Quantity,
			Price:      menu.Price,
			Subtotal:   menu.Price * float64(req.Quantity),
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		cartEntity := model.ToCartEntity(cartModel)

		if err := uc.cartRepo.Create(cartEntity); err != nil {
			uc.logger.Errorf("Error creating cart: %v", err)
			return err
		}
		uc.logger.Infof("Successfully created cart for customer ID %d", customerID)
		return nil
	}
	if cart != nil {
		cart.Quantity += req.Quantity
		cart.Subtotal = menu.Price * float64(cart.Quantity)
		if err := uc.cartRepo.Update(cart); err != nil {
			uc.logger.Errorf("Error updating cart with customer ID %d and menu ID %d: %v", customerID, req.MenuID, err)
			return err
		}
		uc.logger.Infof("Successfully updated cart with customer ID %d and menu ID %d", customerID, req.MenuID)
		return nil
	}
	return nil
}

func (uc *cartUseCase) GetCartByID(id int64) (*model.CartModel, error) {
	cart, err := uc.cartRepo.GetByID(id)
	if err != nil {
		uc.logger.Errorf("Error fetching cart by ID %d: %v", id, err)
		return nil, err
	}
	return model.ToCartModel(cart), nil
}

func (uc *cartUseCase) GetCartByCustomerID(customerID int64, params *model.PaginationQuery) ([]model.UserCartResponse, *model.PaginatedMeta, error) {
	data, err := uc.cartRepo.GetByCustomerID(customerID, params)
	if err != nil {
		uc.logger.Errorf("Error fetching carts for customer ID %d: %v", customerID, err)
		return nil, nil, err
	}

	return data.Data, model.ToPaginatedMeta(data), nil
}

func (uc *cartUseCase) RemoveCart(customerID int64, cartID int64) error {
	// Verify the cart exists and belongs to the customer
	cart, err := uc.cartRepo.GetByID(cartID)
	if err != nil {
		uc.logger.Errorf("Error fetching cart with ID %d: %v", cartID, err)
		return err
	}

	if cart.CustomerID != customerID {
		uc.logger.Errorf("Customer %d attempted to remove cart item %d belonging to customer %d", customerID, cartID, cart.CustomerID)
		return constants.ErrUnauthorized
	}

	if err := uc.cartRepo.RemoveItem(customerID, cartID); err != nil {
		uc.logger.Errorf("Error removing cart item %d for customer %d: %v", cartID, customerID, err)
		return err
	}

	uc.logger.Infof("Successfully removed cart item %d for customer %d", cartID, customerID)
	return nil
}

func (uc *cartUseCase) ClearCart(customerID int64) error {
	// Clear all cart items for the customer
	if err := uc.cartRepo.ClearCustomerCart(customerID); err != nil {
		uc.logger.Errorf("Error clearing cart for customer %d: %v", customerID, err)
		return err
	}

	uc.logger.Infof("Successfully cleared cart for customer %d", customerID)
	return nil
}

func (uc *cartUseCase) BulkDeleteCart(customerID int64, cartIDs []int64) error {
	if err := uc.cartRepo.BulkDelete(customerID, cartIDs); err != nil {
		uc.logger.Errorf("Error deleting carts for customer %d: %v", customerID, err)
		return err
	}

	uc.logger.Infof("Successfully deleted carts for customer %d", customerID)
	return nil
}
