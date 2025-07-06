package usecase

import (
	"bytes"
	"cakestore/internal/constants"
	"cakestore/internal/database"
	"cakestore/internal/domain/entity"
	"cakestore/internal/domain/model"
	"cakestore/internal/repository"
	"cakestore/utils"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/midtrans/midtrans-go"
	"github.com/sirupsen/logrus"
)

type PaymentUseCase interface {
	CreatePaymentURL(order *entity.Order) (*model.PaymentResponse, error)
	GetOrderStatus(orderID string) (string, error)
	UpdateOrderStatus(id string, status constants.PaymentStatus) error
	GetPaymentByOrderID(order *entity.Order) (*entity.Payment, error)
}

type paymentUseCase struct {
	paymentRepository repository.PaymentRepository
	endpoint          string
	log               *logrus.Logger
	env               string
	cache             database.RedisCache
}

func NewPaymentUseCase(
	endpoint string,
	paymentRepository repository.PaymentRepository,
	log *logrus.Logger,
	env string,
	cache database.RedisCache,
) PaymentUseCase {
	return &paymentUseCase{
		endpoint:          endpoint,
		paymentRepository: paymentRepository,
		log:               log,
		env:               env,
		cache:             cache,
	}
}

func (uc *paymentUseCase) GetPaymentByOrderID(order *entity.Order) (*entity.Payment, error) {
	start := time.Now()
	defer func() {
		uc.log.Infof("GetPaymentByOrderID took %v", time.Since(start))
	}()

	// Try to get the payment from the cache first
	cacheKey := fmt.Sprintf("payment:order:%d", order.ID)
	var payment entity.Payment
	if err := uc.cache.Get(context.Background(), cacheKey, &payment); err == nil {
		uc.log.Info("Payment fetched from cache")
		return &payment, nil
	}

	// If not in cache, get from the database
	paymentEntity, err := uc.paymentRepository.GetPaymentByOrderID(order.ID)
	if err != nil {
		return nil, err
	}

	// Store the payment in the cache for future requests
	if err := uc.cache.Set(context.Background(), cacheKey, paymentEntity, 5*time.Minute); err != nil {
		uc.log.Errorf("Error setting cache for payment by order ID %d: %v", order.ID, err)
	}

	return paymentEntity, nil
}

func (uc *paymentUseCase) CreatePaymentURL(order *entity.Order) (*model.PaymentResponse, error) {
	var req model.CreatePaymentRequest

	req.TransactionDetails = midtrans.TransactionDetails{
		OrderID:  "ORDER-" + strconv.Itoa(int(order.ID)) + "-" + uuid.New().String(),
		GrossAmt: int64(order.TotalPrice),
	}

	headers := utils.GenerateRequestHeader()

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequest("POST", uc.endpoint+"/snap/v1/transactions", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	for key, value := range headers {
		httpReq.Header.Set(key, value)
	}

	// optional: additional notification URLs
	// httpReq.Header.Set("X-Append-Notification", "https://5a48-2a09-bac5-3a09-25d7-00-3c5-35.ngrok-free.app/payment/notification/")

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// read response body
	// Read and cache the body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Restore the body so it can be decoded again
	resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("failed to create payment URL, status code: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var paymentResponse model.PaymentResponse
	if err := json.NewDecoder(resp.Body).Decode(&paymentResponse); err != nil {
		return nil, err
	}

	// insert payment to db
	payment := &entity.Payment{
		OrderID: order.ID,
		Amount:  order.TotalPrice,
		// FIXME workaround for midtrans webhook delay or error, force update to success
		Status:       constants.PaymentStatusSuccess,
		PaymentToken: paymentResponse.Token,
		PaymentURL:   paymentResponse.RedirectURL,
	}
	if err := uc.paymentRepository.CreatePayment(payment); err != nil {
		return nil, err
	}

	return &paymentResponse, nil
}

func (uc *paymentUseCase) GetOrderStatus(orderID string) (string, error) {
	start := time.Now()
	defer func() {
		uc.log.Infof("GetOrderStatus took %v", time.Since(start))
	}()

	// Try to get the order status from the cache first
	cacheKey := fmt.Sprintf("order_status:%s", orderID)
	var status string
	if err := uc.cache.Get(context.Background(), cacheKey, &status); err == nil {
		uc.log.Info("Order status fetched from cache")
		return status, nil
	}

	// If not in cache, get from the external service
	endpoint := fmt.Sprintf("%s/v2/%s/status", uc.endpoint, orderID)
	headers := utils.GenerateRequestHeader()

	httpReq, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return "", err
	}

	for key, value := range headers {
		httpReq.Header.Set(key, value)
	}

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get order status, status code: %d", resp.StatusCode)
	}

	var orderStatus model.GetOrderStatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&orderStatus); err != nil {
		return "", err
	}

	if orderStatus.StatusCode != "200" {
		return "", fmt.Errorf("failed to get order status, status code: %s", orderStatus.StatusCode)
	}

	// Store the order status in the cache for future requests
	if err := uc.cache.Set(context.Background(), cacheKey, orderStatus.TransactionStatus, 5*time.Minute); err != nil {
		uc.log.Errorf("Error setting cache for order status %s: %v", orderID, err)
	}

	return orderStatus.TransactionStatus, nil
}

func (uc *paymentUseCase) UpdateOrderStatus(id string, status constants.PaymentStatus) error {
	var orderID int64
	var err error

	if uc.env == "development" {
		orderID, err = uc.paymentRepository.GetPendingPayment()
		if err != nil {
			uc.log.Errorf("Error getting pending payment: %v", err)
			return err
		}
	} else {
		uc.log.Info("Running in production mode, updating payment status")
		orderID, err = strconv.ParseInt(id, 10, 64)
		if err != nil {
			return err
		}
	}

	payment := model.ToPaymentEntity(&model.PaymentModel{
		OrderID: orderID,
		Status:  status,
	})

	if err := uc.paymentRepository.UpdatePayment(payment); err != nil {
		uc.log.Errorf("Error updating order status: %v", err)
		return err
	}

	// Invalidate cache
	cacheKey := fmt.Sprintf("payment:order:%d", orderID)
	if err := uc.cache.Delete(context.Background(), cacheKey); err != nil {
		uc.log.Errorf("Error deleting cache for payment by order ID %d: %v", orderID, err)
	}
	orderStatusCacheKey := fmt.Sprintf("order_status:%s", id)
	if err := uc.cache.Delete(context.Background(), orderStatusCacheKey); err != nil {
		uc.log.Errorf("Error deleting cache for order status %s: %v", id, err)
	}

	return nil
}
