package usecase

import (
	"cakestore/internal/database"
	"cakestore/internal/domain/entity"
	"cakestore/internal/domain/model"
	"cakestore/internal/repository"
	"context"
	"errors"
	"fmt"
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
	UpdateFoodStatus(orderID int64, foodStatus entity.FoodStatus) error
}

type orderUseCaseImpl struct {
	orderRepo    repository.OrderRepository
	menuRepo     repository.MenuRepository
	customerRepo repository.CustomerRepository
	logger       *logrus.Logger
	env          string
	cache        database.RedisCache
}

func NewOrderUseCase(
	orderRepo repository.OrderRepository,
	menuRepo repository.MenuRepository,
	customerRepo repository.CustomerRepository,
	logger *logrus.Logger,
	env string,
	cache database.RedisCache,
) OrderUseCase {
	return &orderUseCaseImpl{
		orderRepo:    orderRepo,
		menuRepo:     menuRepo,
		customerRepo: customerRepo,
		logger:       logger,
		env:          env,
		cache:        cache,
	}
}

func (uc *orderUseCaseImpl) UpdateFoodStatus(orderID int64, foodStatus entity.FoodStatus) error {
	if err := uc.orderRepo.UpdateFoodStatus(orderID, foodStatus); err != nil {
		return err
	}

	// Invalidate cache
	cacheKey := fmt.Sprintf("order:%d", orderID)
	if err := uc.cache.Delete(context.Background(), cacheKey); err != nil {
		uc.logger.Errorf("Error deleting cache for order ID %d: %v", orderID, err)
	}
	if err := uc.cache.Delete(context.Background(), "orders:all:*"); err != nil {
		uc.logger.Errorf("Error deleting cache for all orders: %v", err)
	}

	return nil
}

func (uc *orderUseCaseImpl) GetPendingOrder(customerID int64, orderID int64) (*model.OrderResponse, error) {
	start := time.Now()
	defer func() {
		uc.logger.Infof("GetPendingOrder took %v", time.Since(start))
	}()

	// Try to get the order from the cache first
	cacheKey := fmt.Sprintf("order:pending:%d:%d", customerID, orderID)
	var order model.OrderResponse
	if err := uc.cache.Get(context.Background(), cacheKey, &order); err == nil {
		uc.logger.Info("Pending order fetched from cache")
		return &order, nil
	}

	// If not in cache, get from the database
	orderEntity, err := uc.orderRepo.GetPendingPaymentByOrderID(customerID, orderID)
	if err != nil {
		return nil, err
	}
	response := model.ToOrderResponse(&orderEntity)

	// Store the order in the cache for future requests
	if err := uc.cache.Set(context.Background(), cacheKey, response, 5*time.Minute); err != nil {
		uc.logger.Errorf("Error setting cache for pending order: %v", err)
	}

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
		// Validate menu exists
		_, err := uc.menuRepo.GetByID(item.MenuID)
		if err != nil {
			return nil, errors.New("menu not found")
		}

		orderItem := entity.OrderItem{
			MenuID:   item.MenuID,
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
		FoodStatus: entity.FoodStatusPending,
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
	start := time.Now()
	defer func() {
		uc.logger.Infof("GetOrderByID took %v", time.Since(start))
	}()

	// Try to get the order from the cache first
	cacheKey := fmt.Sprintf("order:%d", id)
	var order model.OrderResponse
	if err := uc.cache.Get(context.Background(), cacheKey, &order); err == nil {
		uc.logger.Info("Order fetched from cache")
		return &order, nil
	}

	// If not in cache, get from the database
	orderEntity, err := uc.orderRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	response := model.ToOrderResponse(orderEntity)

	// Store the order in the cache for future requests
	if err := uc.cache.Set(context.Background(), cacheKey, response, 5*time.Minute); err != nil {
		uc.logger.Errorf("Error setting cache for order ID %d: %v", id, err)
	}

	return response, nil
}

func (uc *orderUseCaseImpl) GetCustomerOrders(customerID int64) ([]model.OrderResponse, error) {
	start := time.Now()
	defer func() {
		uc.logger.Infof("GetCustomerOrders took %v", time.Since(start))
	}()

	// Try to get the orders from the cache first
	cacheKey := fmt.Sprintf("orders:customer:%d", customerID)
	var orders []model.OrderResponse
	if err := uc.cache.Get(context.Background(), cacheKey, &orders); err == nil {
		uc.logger.Info("Customer orders fetched from cache")
		return orders, nil
	}

	// If not in cache, get from the database
	orderEntities, err := uc.orderRepo.GetByCustomerID(customerID)
	if err != nil {
		return nil, err
	}

	responses := make([]model.OrderResponse, len(orderEntities))
	for i, order := range orderEntities {
		responses[i] = *model.ToOrderResponse(&order)
	}

	// Store the orders in the cache for future requests
	if err := uc.cache.Set(context.Background(), cacheKey, responses, 5*time.Minute); err != nil {
		uc.logger.Errorf("Error setting cache for customer orders: %v", err)
	}

	return responses, nil
}

func (uc *orderUseCaseImpl) UpdateOrderStatus(id string, status string) error {
	orderStatus := entity.OrderStatus(status)
	uc.logger.Tracef("UpdateOrderStatus usecase ~ in %s", uc.env)

	var orderID int64
	var err error

	if uc.env == "development" {
		orderID, err = uc.orderRepo.GetPendingOrder()
		if err != nil {
			uc.logger.Errorf("UpdateOrderStatus usecase ~Error getting pending order: %v", err)
			return err
		}
	} else {
		orderID, err = strconv.ParseInt(id, 10, 64)
		if err != nil {
			return err
		}
	}

	if err := uc.orderRepo.UpdateStatus(orderID, orderStatus); err != nil {
		uc.logger.Errorf("Error updating order status: %v", err)
		return err
	}

	// Invalidate cache
	cacheKey := fmt.Sprintf("order:%d", orderID)
	if err := uc.cache.Delete(context.Background(), cacheKey); err != nil {
		uc.logger.Errorf("Error deleting cache for order ID %d: %v", orderID, err)
	}
	if err := uc.cache.Delete(context.Background(), "orders:all:*"); err != nil {
		uc.logger.Errorf("Error deleting cache for all orders: %v", err)
	}

	return nil
}

func (uc *orderUseCaseImpl) DeleteOrder(id int64) error {
	if err := uc.orderRepo.Delete(id); err != nil {
		uc.logger.Errorf("Error deleting order: %v", err)
		return err
	}

	// Invalidate cache
	cacheKey := fmt.Sprintf("order:%d", id)
	if err := uc.cache.Delete(context.Background(), cacheKey); err != nil {
		uc.logger.Errorf("Error deleting cache for order ID %d: %v", id, err)
	}
	if err := uc.cache.Delete(context.Background(), "orders:all:*"); err != nil {
		uc.logger.Errorf("Error deleting cache for all orders: %v", err)
	}

	return nil
}

func (uc *orderUseCaseImpl) GetAllOrders(params *model.PaginationQuery) (*[]model.OrderResponse, *model.PaginatedMeta, error) {
	start := time.Now()
	defer func() {
		uc.logger.Infof("GetAllOrders took %v", time.Since(start))
	}()

	uc.logger.Trace("GetAllOrders usecase ~ in ", uc.env)

	// Try to get the orders from the cache first
	cacheKey := fmt.Sprintf("orders:all:page:%d:limit:%d", params.Page, params.Limit)
	var cachedData struct {
		Data []model.OrderResponse
		Meta *model.PaginatedMeta
	}
	if err := uc.cache.Get(context.Background(), cacheKey, &cachedData); err == nil {
		uc.logger.Info("All orders fetched from cache")
		return &cachedData.Data, cachedData.Meta, nil
	}

	// If not in cache, get from the database
	orders, meta, err := uc.orderRepo.GetAll(params)
	if err != nil {
		uc.logger.Errorf("Error getting all orders: %v", err)
		return nil, nil, err
	}

	responses := make([]model.OrderResponse, len(orders))
	for i, order := range orders {
		responses[i] = *model.ToOrderResponse(&order)
	}

	// Store the orders in the cache for future requests
	if err := uc.cache.Set(context.Background(), cacheKey, struct {
		Data []model.OrderResponse
		Meta *model.PaginatedMeta
	}{Data: responses, Meta: meta}, 5*time.Minute); err != nil {
		uc.logger.Errorf("Error setting cache for all orders: %v", err)
	}

	return &responses, meta, nil
}
