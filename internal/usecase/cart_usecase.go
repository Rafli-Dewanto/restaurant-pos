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

type CartUseCase interface {
	CreateCart(customerID int, req *model.AddCart) error
	GetCartByID(id int) (*model.CartModel, error)
	GetCartByCustomerID(customerID int, params *model.PaginationQuery) (*model.PaginationResponse, error)
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
	if err != nil {
		uc.logger.Errorf("Error getting cart with customer ID %d and cake ID %d: %v", customerID, req.CakeID, err)
		return err
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

func (uc *cartUseCase) GetCartByID(id int) (*model.CartModel, error) {
	cart, err := uc.cartRepo.GetByID(id)
	if err != nil {
		uc.logger.Errorf("Error fetching cart by ID %d: %v", id, err)
		return nil, err
	}
	return model.ToCartModel(cart), nil
}

func (uc *cartUseCase) GetCartByCustomerID(customerID int, params *model.PaginationQuery) (*model.PaginationResponse, error) {
	data, err := uc.cartRepo.GetByCustomerID(customerID, params)
	if err != nil {
		uc.logger.Errorf("Error fetching carts for customer ID %d: %v", customerID, err)
		return nil, err
	}

	carts, ok := data.Data.([]*entity.Cart)
	if !ok {
		uc.logger.Error("Invalid data type for carts.Data")
		return nil, constants.ErrInvalidInterfaceConversion
	}

	var cartModels []*model.CartModel

	// convert cart entity to cart model
	for _, cart := range carts {
		cartModels = append(cartModels, model.ToCartModel(cart))
	} 
	data.Data = cartModels
	return data, nil
}

func (uc *cartUseCase) ClearCart(cartID int) error {
	// ...logic to clear cart...
	return nil

}
