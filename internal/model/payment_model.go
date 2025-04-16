package model

import (
	"time"

	"github.com/midtrans/midtrans-go"
)

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
