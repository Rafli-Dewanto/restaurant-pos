package usecase

import (
	"cakestore/internal/entity"
	"cakestore/internal/model"
	"cakestore/internal/repository"
	"errors"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
)

type OrderUseCase interface {
	CreateOrder(customerID int, request *model.CreateOrderRequest) (*entity.Order, error)
	GetOrderByID(id int) (*model.OrderResponse, error)
	GetCustomerOrders(customerID int) ([]model.OrderResponse, error)
	UpdateOrderStatus(id string, status string) error
	DeleteOrder(id int) error
}

type orderUseCase struct {
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
	return &orderUseCase{
		orderRepo:    orderRepo,
		cakeRepo:     cakeRepo,
		customerRepo: customerRepo,
		logger:       logger,
	}
}

func (uc *orderUseCase) CreateOrder(customerID int, request *model.CreateOrderRequest) (*entity.Order, error) {
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
		Address:    request.Address,
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

func (uc *orderUseCase) GetOrderByID(id int) (*model.OrderResponse, error) {
	order, err := uc.orderRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	return uc.orderToResponse(order), nil
}

func (uc *orderUseCase) GetCustomerOrders(customerID int) ([]model.OrderResponse, error) {
	orders, err := uc.orderRepo.GetByCustomerID(customerID)
	if err != nil {
		return nil, err
	}

	responses := make([]model.OrderResponse, len(orders))
	for i, order := range orders {
		responses[i] = *uc.orderToResponse(&order)
	}

	return responses, nil
}

func (uc *orderUseCase) UpdateOrderStatus(id string, status string) error {
	orderStatus := entity.OrderStatus(status)

	if uc.env == "development" {
		if err := uc.orderRepo.UpdateStatus(9, orderStatus); err != nil {
			uc.logger.Errorf("Error updating order status: %v", err)
			return err
		}
		return nil
	}

	orderID, err := strconv.Atoi(id)
	if err != nil {
		return err
	}

	if err := uc.orderRepo.UpdateStatus(orderID, orderStatus); err != nil {
		uc.logger.Errorf("Error updating order status: %v", err)
		return err
	}
	return nil
}

func (uc *orderUseCase) DeleteOrder(id int) error {
	if err := uc.orderRepo.Delete(id); err != nil {
		uc.logger.Errorf("Error deleting order: %v", err)
		return err
	}
	return nil
}

func (uc *orderUseCase) orderToResponse(order *entity.Order) *model.OrderResponse {
	itemResponses := make([]model.OrderItemResponse, len(order.Items))
	for i, item := range order.Items {
		itemResponses[i] = model.OrderItemResponse{
			ID:       item.ID,
			Cake:     *model.CakeToResponse(&item.Cake),
			Quantity: item.Quantity,
			Price:    item.Price,
		}
	}

	return &model.OrderResponse{
		ID: order.ID,
		Customer: model.CustomerResponse{
			ID:      order.Customer.ID,
			Name:    order.Customer.Name,
			Email:   order.Customer.Email,
			Address: order.Customer.Address,
		},
		Status:     string(order.Status),
		TotalPrice: order.TotalPrice,
		Address:    order.Address,
		Items:      itemResponses,
		CreatedAt:  order.CreatedAt.Format(time.RFC3339),
		UpdatedAt:  order.UpdatedAt.Format(time.RFC3339),
	}
}
