package controller

import (
	"cakestore/internal/domain/entity"
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
	customerID := ctx.Locals("customer_id").(int64)

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
		// if error delete previous order
		if err := c.orderUseCase.DeleteOrder(order.ID); err != nil {
			c.logger.Error("Failed to delete order: ", err)
			return utils.WriteErrorResponse(ctx, fiber.StatusInternalServerError, "Failed to delete order")
		}
		return utils.WriteErrorResponse(ctx, fiber.StatusInternalServerError, "Failed to create payment URL")
	}

	return utils.WriteResponse(ctx, fiber.StatusCreated, paymentURL, "Order created successfully", nil)
}

func (c *OrderController) GetOrderByID(ctx *fiber.Ctx) error {
	orderID, err := strconv.ParseInt(ctx.Params("id"), 10, 64)
	if err != nil {
		c.logger.Error("Failed to parse order ID: ", err)
		return utils.WriteErrorResponse(ctx, fiber.StatusBadRequest, "Invalid order ID")
	}

	order, err := c.orderUseCase.GetOrderByID(orderID)
	if err != nil {
		c.logger.Error("Failed to get order: ", err)
		return utils.WriteErrorResponse(ctx, fiber.StatusInternalServerError, "Failed to get order")
	}

	return utils.WriteResponse(ctx, fiber.StatusOK, order, "Order fetched successfully", nil)
}

func (c *OrderController) GetCustomerOrders(ctx *fiber.Ctx) error {
	// Get customer ID from JWT token
	customerID := ctx.Locals("customer_id").(int64)

	orders, err := c.orderUseCase.GetCustomerOrders(customerID)
	if err != nil {
		c.logger.Error("Failed to get customer orders: ", err)
		return utils.WriteErrorResponse(ctx, fiber.StatusInternalServerError, "Failed to get customer orders")
	}

	return utils.WriteResponse(ctx, fiber.StatusOK, orders, "Customer orders fetched successfully", nil)
}

func (c *OrderController) GetAllOrders(ctx *fiber.Ctx) error {
	c.logger.Tracef("GetAllOrders controller")

	var params model.PaginationQuery
	if err := ctx.QueryParser(&params); err != nil {
		c.logger.Error("Failed to parse pagination query: ", err)
		return utils.WriteErrorResponse(ctx, fiber.StatusBadRequest, "Invalid pagination query")
	}

	orders, meta, err := c.orderUseCase.GetAllOrders(&params)
	if err != nil {
		c.logger.Error("Failed to get all orders: ", err)
		return utils.WriteErrorResponse(ctx, fiber.StatusInternalServerError, "Failed to get all orders")
	}
	return utils.WriteResponse(ctx, fiber.StatusOK, orders, "All orders fetched successfully", meta)
}

func (c *OrderController) UpdateFoodStatus(ctx *fiber.Ctx) error {
	orderID, err := strconv.ParseInt(ctx.Params("id"), 10, 64)
	if err != nil {
		c.logger.Error("Failed to parse order ID: ", err)
		return utils.WriteErrorResponse(ctx, fiber.StatusBadRequest, "Invalid order ID")
	}

	var req model.UpdateFoodStatusRequest
	err = ctx.BodyParser(&req)
	if err != nil {
		c.logger.Error("Failed to parse body: ", err)
		return utils.WriteErrorResponse(ctx, fiber.StatusBadRequest, "Invalid request body")
	}

	if err := c.validator.Struct(req); err != nil {
		c.logger.Error("Validation failed: ", err)
		return utils.WriteErrorResponse(ctx, fiber.StatusBadRequest, err.Error())
	}

	if err := c.orderUseCase.UpdateFoodStatus(orderID, entity.FoodStatus(req.FoodStatus)); err != nil {
		c.logger.Error("Failed to update food status: ", err)
		return utils.WriteErrorResponse(ctx, fiber.StatusInternalServerError, "Failed to update food status")
	}

	return utils.WriteResponse(ctx, fiber.StatusOK, nil, "Food status updated successfully", nil)
}
