package controller

import (
	"cakestore/internal/entity"
	"cakestore/internal/model"
	"cakestore/internal/usecase"
	"crypto/sha512"
	"encoding/hex"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

type PaymentController interface {
	GetTransactionStatus(ctx *fiber.Ctx) error
}

type PaymentControllerImpl struct {
	logger            *logrus.Logger
	midtransServerKey string
	orderUseCase      usecase.OrderUseCase
}

func NewPaymentController(logger *logrus.Logger, midtransServerKey string, orderUseCase usecase.OrderUseCase) PaymentController {
	return &PaymentControllerImpl{
		logger:            logger,
		midtransServerKey: midtransServerKey,
		orderUseCase:      orderUseCase,
	}
}

func (c *PaymentControllerImpl) GetTransactionStatus(ctx *fiber.Ctx) error {
	c.logger.Trace("UpdateOrderStatus called")
	var notif model.MidtransNotification
	if err := ctx.BodyParser(&notif); err != nil {
		return model.WriteErrorResponse(ctx, fiber.StatusBadRequest, "Invalid request body")
	}

	rawSignature := notif.OrderID + notif.StatusCode + notif.GrossAmount + c.midtransServerKey

	hash := sha512.Sum512([]byte(rawSignature))
	computedSignature := hex.EncodeToString(hash[:])
	if computedSignature != notif.SignatureKey {
		c.logger.Errorf("Invalid signature key: %s", notif.SignatureKey)
		return model.WriteErrorResponse(ctx, fiber.StatusUnauthorized, "Invalid signature key")
	}
	c.logger.Info("Webhook received")

	switch notif.TransactionStatus {
	case "settlement":
		if err := c.orderUseCase.UpdateOrderStatus(notif.OrderID, string(entity.OrderStatusPaid)); err != nil {
			c.logger.Errorf("Failed to update order status: %v", err)
			return model.WriteErrorResponse(ctx, fiber.StatusInternalServerError, "Failed to update order status")
		}
		return model.WriteResponse(ctx, fiber.StatusOK, nil, "Transaction successful", nil)
	case "pending":
		if err := c.orderUseCase.UpdateOrderStatus(notif.OrderID, string(entity.OrderStatusPending)); err != nil {
			c.logger.Errorf("Failed to update order status: %v", err)
			return model.WriteErrorResponse(ctx, fiber.StatusInternalServerError, "Failed to update order status")
		}
		return model.WriteResponse(ctx, fiber.StatusOK, nil, "Transaction pending", nil)
	case "expire", "cancel":
		if err := c.orderUseCase.UpdateOrderStatus(notif.OrderID, string(entity.OrderStatusCancelled)); err != nil {
			c.logger.Errorf("Failed to update order status: %v", err)
			return model.WriteErrorResponse(ctx, fiber.StatusInternalServerError, "Failed to update order status")
		}
		return model.WriteResponse(ctx, fiber.StatusOK, nil, "Transaction cancelled", nil)
	}
	return ctx.SendStatus(fiber.StatusOK)
}
