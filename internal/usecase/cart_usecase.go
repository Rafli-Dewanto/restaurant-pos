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
	CreateCart(customerID int, req *model.AddCart) error
	GetCartByID(id int) (*model.CartModel, error)
	GetCartByCustomerID(customerID int, params *model.PaginationQuery) ([]*model.CartModel, *model.PaginatedMeta, error)
	RemoveCart(customerID int, cartID int) error
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

func (uc *cartUseCase) CreateCart(customerID int, req *model.AddCart) error {
	if err := uc.validate.Struct(req); err != nil {
		uc.logger.Errorf("Validation failed for request: %v", err)
		return err
	}

	cake, err := uc.cakeRepo.GetByID(req.CakeID)
	if err != nil {
		uc.logger.Errorf("Error getting cake with ID %d: %v", req.CakeID, err)
		return err
	}

	// check if customer already have the same cake added, if so update the quantity
	cart, err := uc.cartRepo.GetByCustomerIDAndCakeID(customerID, req.CakeID)
	// if not, create a new cart
	if err != nil {
		cartModel := &model.CartModel{
			CustomerID: customerID,
			CakeID:     req.CakeID,
			Quantity:   req.Quantity,
			Price:      cake.Price,
			Subtotal:   cake.Price * float64(req.Quantity),
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
		cart.Subtotal = cake.Price * float64(cart.Quantity)
		if err := uc.cartRepo.Update(cart); err != nil {
			uc.logger.Errorf("Error updating cart with customer ID %d and cake ID %d: %v", customerID, req.CakeID, err)
			return err
		}
		uc.logger.Infof("Successfully updated cart with customer ID %d and cake ID %d", customerID, req.CakeID)
		return nil
	}
	return nil
}

func (uc *cartUseCase) GetCartByID(id int) (*model.CartModel, error) {
	cart, err := uc.cartRepo.GetByID(id)
	if err != nil {
		uc.logger.Errorf("Error fetching cart by ID %d: %v", id, err)
		return nil, err
	}
	return model.ToCartModel(cart), nil
}

func (uc *cartUseCase) GetCartByCustomerID(customerID int, params *model.PaginationQuery) ([]*model.CartModel, *model.PaginatedMeta, error) {
	data, err := uc.cartRepo.GetByCustomerID(customerID, params)
	if err != nil {
		uc.logger.Errorf("Error fetching carts for customer ID %d: %v", customerID, err)
		return nil, nil, err
	}

	var cartModels []*model.CartModel

	// convert cart entity to cart model
	for _, cart := range data.Data {
		cartModels = append(cartModels, model.ToCartModel(cart))
	}

	return cartModels, model.ToPaginatedMeta(data), nil
}

func (uc *cartUseCase) RemoveCart(customerID int, cartID int) error {
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

func (uc *cartUseCase) ClearCart(customerID int) error {
	// Clear all cart items for the customer
	if err := uc.cartRepo.ClearCustomerCart(customerID); err != nil {
		uc.logger.Errorf("Error clearing cart for customer %d: %v", customerID, err)
		return err
	}

	uc.logger.Infof("Successfully cleared cart for customer %d", customerID)
	return nil
}
