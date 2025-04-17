package controller

import (
	"cakestore/internal/constants"
	"cakestore/internal/domain/entity"
	"cakestore/internal/domain/model"
	"cakestore/internal/usecase"
	"cakestore/utils"
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
	paymentUseCase    usecase.PaymentUseCase
}

func NewPaymentController(logger *logrus.Logger, midtransServerKey string, orderUseCase usecase.OrderUseCase, paymentUseCase usecase.PaymentUseCase) PaymentController {
	return &PaymentControllerImpl{
		logger:            logger,
		midtransServerKey: midtransServerKey,
		orderUseCase:      orderUseCase,
		paymentUseCase:    paymentUseCase,
	}
}

func (c *PaymentControllerImpl) GetTransactionStatus(ctx *fiber.Ctx) error {
	c.logger.Trace("UpdateOrderStatus called")
	var notif model.MidtransNotification
	if err := ctx.BodyParser(&notif); err != nil {
		return utils.WriteErrorResponse(ctx, fiber.StatusBadRequest, "Invalid request body")
	}

	rawSignature := notif.OrderID + notif.StatusCode + notif.GrossAmount + c.midtransServerKey

	hash := sha512.Sum512([]byte(rawSignature))
	computedSignature := hex.EncodeToString(hash[:])
	if computedSignature != notif.SignatureKey {
		c.logger.Errorf("Invalid signature key: %s", notif.SignatureKey)
		return utils.WriteErrorResponse(ctx, fiber.StatusUnauthorized, "Invalid signature key")
	}
	c.logger.Info("Webhook received")

	switch notif.TransactionStatus {
	case "capture", "settlement":
		if err := c.orderUseCase.UpdateOrderStatus(notif.OrderID, string(entity.OrderStatusPaid)); err != nil {
			c.logger.Errorf("Failed to update order status for settlement: %v", err)
			return utils.WriteErrorResponse(ctx, fiber.StatusInternalServerError, "Failed to update order status")
		}
		if err := c.paymentUseCase.UpdateOrderStatus(notif.OrderID, constants.PaymentStatusSuccess); err != nil {
			c.logger.Info("Failed to update payment status for settelement")
			return utils.WriteErrorResponse(ctx, fiber.StatusInternalServerError, "Failed to update order status")
		}
		return utils.WriteResponse(ctx, fiber.StatusOK, nil, "Transaction successful", nil)
	case "pending":
		if err := c.orderUseCase.UpdateOrderStatus(notif.OrderID, string(entity.OrderStatusPending)); err != nil {
			c.logger.Errorf("Failed to update order status for pending: %v", err)
			return utils.WriteErrorResponse(ctx, fiber.StatusInternalServerError, "Failed to update order status")
		}
		if err := c.paymentUseCase.UpdateOrderStatus(notif.OrderID, constants.PaymentStatusPending); err != nil {
			c.logger.Errorf("Failed to update payment status for pending: %v", err)
			return utils.WriteErrorResponse(ctx, fiber.StatusInternalServerError, "Failed to update order status")
		}
		return utils.WriteResponse(ctx, fiber.StatusOK, nil, "Transaction pending", nil)
	case "expire", "cancel":
		if err := c.orderUseCase.UpdateOrderStatus(notif.OrderID, string(entity.OrderStatusCancelled)); err != nil {
			c.logger.Errorf("Failed to update order status for expire: %v", err)
			return utils.WriteErrorResponse(ctx, fiber.StatusInternalServerError, "Failed to update order status")
		}
		if err := c.paymentUseCase.UpdateOrderStatus(notif.OrderID, constants.PaymentStatusCancelled); err != nil {
			c.logger.Errorf("Failed to update payment status for expire: %v", err)
			return utils.WriteErrorResponse(ctx, fiber.StatusInternalServerError, "Failed to update order status")
		}
		return utils.WriteResponse(ctx, fiber.StatusOK, nil, "Transaction cancelled", nil)
	}
	return ctx.SendStatus(fiber.StatusOK)
}
