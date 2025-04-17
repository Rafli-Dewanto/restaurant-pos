package model

import (
	"cakestore/internal/constants"
	"cakestore/internal/domain/entity"
	"time"

	"github.com/midtrans/midtrans-go"
)

type PaymentModel struct {
	ID           int                     `json:"id"`
	OrderID      int                     `json:"order_id"`
	Amount       float64                 `json:"amount"`
	Status       constants.PaymentStatus `json:"status"`
	PaymentToken string                  `json:"payment_token"`
	PaymentURL   string                  `json:"payment_url"`
}

type CreatePaymentRequest struct {
	TransactionDetails midtrans.TransactionDetails `json:"transaction_details" validate:"required"`
	ItemDetails        []midtrans.ItemDetails      `json:"item_details"`
	CustomerDetails    midtrans.CustomerDetails    `json:"customer_details"`
}

type PaymentResponse struct {
	Token       string `json:"token"`
	RedirectURL string `json:"redirect_url"`
}

type GetOrderStatusResponse struct {
	StatusCode               string    `json:"status_code"`
	StatusMessage            string    `json:"status_message"`
	TransactionID            string    `json:"transaction_id"`
	MaskedCard               string    `json:"masked_card"`
	OrderID                  string    `json:"order_id"`
	PaymentType              string    `json:"payment_type"`
	TransactionTime          time.Time `json:"transaction_time"`
	TransactionStatus        string    `json:"transaction_status"`
	FraudStatus              string    `json:"fraud_status"`
	ApprovalCode             string    `json:"approval_code"`
	SignatureKey             string    `json:"signature_key"`
	Bank                     string    `json:"bank"`
	GrossAmount              string    `json:"gross_amount"`
	ChannelResponseCode      string    `json:"channel_response_code"`
	ChannelResponseMessage   string    `json:"channel_response_message"`
	CardType                 string    `json:"card_type"`
	PaymentOptionType        string    `json:"payment_option_type"`
	ShopeepayReferenceNumber string    `json:"shopeepay_reference_number"`
	ReferenceID              string    `json:"reference_id"`
}

type MidtransNotification struct {
	TransactionTime   string `json:"transaction_time"`
	TransactionStatus string `json:"transaction_status"`
	TransactionID     string `json:"transaction_id"`
	StatusMessage     string `json:"status_message"`
	StatusCode        string `json:"status_code"`
	SignatureKey      string `json:"signature_key"`
	SettlementTime    string `json:"settlement_time"`
	PaymentType       string `json:"payment_type"`
	OrderID           string `json:"order_id"`
	MerchantID        string `json:"merchant_id"`
	GrossAmount       string `json:"gross_amount"`
	FraudStatus       string `json:"fraud_status"`
	Currency          string `json:"currency"`
}

func ToPaymentEntity(paymentModel *PaymentModel) *entity.Payment {
	return &entity.Payment{
		ID:           paymentModel.ID,
		OrderID:      paymentModel.OrderID,
		Amount:       paymentModel.Amount,
		Status:       paymentModel.Status,
		PaymentToken: paymentModel.PaymentToken,
		PaymentURL:   paymentModel.PaymentURL,
		UpdatedAt:    time.Now(),
	}
}
