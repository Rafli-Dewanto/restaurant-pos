package usecase

import (
	"cakestore/internal/domain/entity"
	"cakestore/internal/domain/model"
	"cakestore/internal/repository"
	"errors"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
)

type OrderUseCase interface {
	CreateOrder(customerID int64, request *model.CreateOrderRequest) (*entity.Order, error)
	GetOrderByID(id int64) (*model.OrderResponse, error)
	GetPendingOrder(customerID int64, orderID int64) (*model.OrderResponse, error)
	GetAllOrders(params *model.PaginationQuery) (*[]model.OrderResponse, *model.PaginatedMeta, error)
	GetCustomerOrders(customerID int64) ([]model.OrderResponse, error)
	UpdateOrderStatus(id string, status string) error
	DeleteOrder(id int64) error
}

type orderUseCaseImpl struct {
	orderRepo    repository.OrderRepository
	cakeRepo     repository.CakeRepository
	customerRepo repository.CustomerRepository
	logger       *logrus.Logger
	env          string
}

func NewOrderUseCase(
	orderRepo repository.OrderRepository,
	cakeRepo repository.CakeRepository,
	customerRepo repository.CustomerRepository,
	logger *logrus.Logger,
	env string,
) OrderUseCase {
	return &orderUseCaseImpl{
		orderRepo:    orderRepo,
		cakeRepo:     cakeRepo,
		customerRepo: customerRepo,
		logger:       logger,
		env:          env,
	}
}

func (uc *orderUseCaseImpl) GetPendingOrder(customerID int64, orderID int64) (*model.OrderResponse, error) {
	order, err := uc.orderRepo.GetPendingPaymentByOrderID(customerID, orderID)
	if err != nil {
		return nil, err
	}
	response := model.OrderToResponse(&order)
	return response, nil
}

func (uc *orderUseCaseImpl) CreateOrder(customerID int64, request *model.CreateOrderRequest) (*entity.Order, error) {
	customer, err := uc.customerRepo.GetByID(customerID)
	if err != nil {
		return nil, errors.New("customer not found")
	}

	// Create order items and calculate total price
	var orderItems []entity.OrderItem
	var totalPrice float64

	for _, item := range request.Items {
		// Validate cake exists
		_, err := uc.cakeRepo.GetByID(item.CakeID)
		if err != nil {
			return nil, errors.New("cake not found")
		}

		orderItem := entity.OrderItem{
			CakeID:   item.CakeID,
			Quantity: item.Quantity,
			Price:    item.Price,
		}
		orderItems = append(orderItems, orderItem)
		totalPrice += item.Price * float64(item.Quantity)
	}

	order := &entity.Order{
		CustomerID: customerID,
		Customer:   *customer,
		Status:     entity.OrderStatusPending,
		TotalPrice: totalPrice,
		Address:    customer.Address,
		Items:      orderItems,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := uc.orderRepo.Create(order); err != nil {
		uc.logger.Errorf("Error creating order: %v", err)
		return nil, err
	}

	return order, nil
}

func (uc *orderUseCaseImpl) GetOrderByID(id int64) (*model.OrderResponse, error) {
	order, err := uc.orderRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	return model.OrderToResponse(order), nil
}

func (uc *orderUseCaseImpl) GetCustomerOrders(customerID int64) ([]model.OrderResponse, error) {
	orders, err := uc.orderRepo.GetByCustomerID(customerID)
	if err != nil {
		return nil, err
	}

	responses := make([]model.OrderResponse, len(orders))
	for i, order := range orders {
		responses[i] = *model.OrderToResponse(&order)
	}

	return responses, nil
}

func (uc *orderUseCaseImpl) UpdateOrderStatus(id string, status string) error {
	orderStatus := entity.OrderStatus(status)
	uc.logger.Tracef("UpdateOrderStatus usecase ~ in %s", uc.env)

	if uc.env == "development" {
		orderID, err := uc.orderRepo.GetPendingOrder()
		if err != nil {
			uc.logger.Errorf("UpdateOrderStatus usecase ~Error getting pending order: %v", err)
			return err
		}
		if err := uc.orderRepo.UpdateStatus(orderID, orderStatus); err != nil {
			uc.logger.Errorf("UpdateOrderStatus usecase ~Error updating order status: %v", err)
			return err
		}
		return nil
	}

	orderID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return err
	}

	if err := uc.orderRepo.UpdateStatus(orderID, orderStatus); err != nil {
		uc.logger.Errorf("Error updating order status: %v", err)
		return err
	}
	return nil
}

func (uc *orderUseCaseImpl) DeleteOrder(id int64) error {
	if err := uc.orderRepo.Delete(id); err != nil {
		uc.logger.Errorf("Error deleting order: %v", err)
		return err
	}
	return nil
}

func (uc *orderUseCaseImpl) GetAllOrders(params *model.PaginationQuery) (*[]model.OrderResponse, *model.PaginatedMeta, error) {
	uc.logger.Trace("GetAllOrders usecase ~ in ", uc.env)

	orders, meta, err := uc.orderRepo.GetAll(params)
	if err != nil {
		uc.logger.Errorf("Error getting all orders: %v", err)
		return nil, nil, err
	}

	responses := make([]model.OrderResponse, len(orders))
	for i, order := range orders {
		responses[i] = *model.OrderToResponse(&order)
	}

	return &responses, meta, nil
}
