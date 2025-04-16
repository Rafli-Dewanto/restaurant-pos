package usecase

import (
	"bytes"
	"cakestore/internal/entity"
	"cakestore/internal/model"
	"cakestore/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/midtrans/midtrans-go"
)

type PaymentUseCase interface {
	CreatePaymentURL(order *entity.Order) (*model.PaymentResponse, error)
	GetOrderStatus(orderID string) (string, error)
}

type paymentUseCase struct {
	endpoint string
}

func NewPaymentUseCase(endpoint string) PaymentUseCase {
	return &paymentUseCase{
		endpoint: endpoint,
	}
}

func (uc *paymentUseCase) CreatePaymentURL(order *entity.Order) (*model.PaymentResponse, error) {
	var req model.CreatePaymentRequest

	req.TransactionDetails = midtrans.TransactionDetails{
		OrderID:  strconv.Itoa(order.ID),
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

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("failed to create payment URL, status code: %d", resp.StatusCode)
	}

	var paymentResponse model.PaymentResponse
	if err := json.NewDecoder(resp.Body).Decode(&paymentResponse); err != nil {
		return nil, err
	}

	return &paymentResponse, nil
}

func (uc *paymentUseCase) GetOrderStatus(orderID string) (string, error) {
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

	return orderStatus.TransactionStatus, nil
}
