package controller

import (
	"cakestore/internal/domain/model"
	"cakestore/internal/usecase"
	"cakestore/utils"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

type OrderController struct {
	orderUseCase   usecase.OrderUseCase
	paymentUseCase usecase.PaymentUseCase
	logger         *logrus.Logger
	validator      *validator.Validate
}

func NewOrderController(orderUseCase usecase.OrderUseCase, paymentUseCase usecase.PaymentUseCase, logger *logrus.Logger) *OrderController {
	return &OrderController{
		orderUseCase:   orderUseCase,
		logger:         logger,
		validator:      validator.New(),
		paymentUseCase: paymentUseCase,
	}
}

func (c *OrderController) CreateOrder(ctx *fiber.Ctx) error {
	// Get customer ID from JWT token
	customerID := ctx.Locals("customer_id").(int)

	var request model.CreateOrderRequest
	if err := ctx.BodyParser(&request); err != nil {
		c.logger.Error("Failed to parse body: ", err)
		return utils.WriteErrorResponse(ctx, fiber.StatusBadRequest, "Invalid request body")
	}

	if err := c.validator.Struct(request); err != nil {
		c.logger.Error("Validation failed: ", err)
		return utils.WriteErrorResponse(ctx, fiber.StatusBadRequest, err.Error())
	}

	order, err := c.orderUseCase.CreateOrder(customerID, &request)
	if err != nil {
		c.logger.Error("Failed to create order: ", err)
		return utils.WriteErrorResponse(ctx, fiber.StatusInternalServerError, "Failed to create order")
	}

	_, err = c.orderUseCase.GetOrderByID(order.ID)
	if err != nil {
		c.logger.Error("Failed to get order details: ", err)
		return utils.WriteErrorResponse(ctx, fiber.StatusInternalServerError, "Failed to get order details")
	}

	// make payment link from midtrans
	paymentURL, err := c.paymentUseCase.CreatePaymentURL(order)
	if err != nil {
		c.logger.Error("Failed to create payment URL: ", err.Error())
		return utils.WriteErrorResponse(ctx, fiber.StatusInternalServerError, "Failed to create payment URL")
	}

	utils.WriteResponse(ctx, fiber.StatusCreated, paymentURL, "Order created successfully", nil)
	return nil
}

func (c *OrderController) GetOrderByID(ctx *fiber.Ctx) error {
	orderID, err := strconv.Atoi(ctx.Params("id"))
	if err != nil {
		c.logger.Error("Failed to parse order ID: ", err)
		utils.WriteErrorResponse(ctx, fiber.StatusBadRequest, "Invalid order ID")
	}

	order, err := c.orderUseCase.GetOrderByID(orderID)
	if err != nil {
		c.logger.Error("Failed to get order: ", err)
		utils.WriteErrorResponse(ctx, fiber.StatusInternalServerError, "Failed to get order")
	}

	utils.WriteResponse(ctx, fiber.StatusOK, order, "Order fetched successfully", nil)
	return nil
}

func (c *OrderController) GetCustomerOrders(ctx *fiber.Ctx) error {
	// Get customer ID from JWT token
	customerID := ctx.Locals("customer_id").(int)

	orders, err := c.orderUseCase.GetCustomerOrders(customerID)
	if err != nil {
		c.logger.Error("Failed to get customer orders: ", err)
		utils.WriteErrorResponse(ctx, fiber.StatusInternalServerError, "Failed to get customer orders")
	}

	utils.WriteResponse(ctx, fiber.StatusOK, orders, "Customer orders fetched successfully", nil)
	return nil
}
