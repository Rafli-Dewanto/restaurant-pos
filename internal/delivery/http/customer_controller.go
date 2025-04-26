package controller

import (
	"cakestore/internal/constants"
	"cakestore/internal/domain/model"
	"cakestore/internal/usecase"
	"cakestore/utils"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

type CustomerController struct {
	customerUseCase usecase.CustomerUseCase
	logger          *logrus.Logger
	validator       *validator.Validate
}

func NewCustomerController(customerUseCase usecase.CustomerUseCase, logger *logrus.Logger) *CustomerController {
	return &CustomerController{
		customerUseCase: customerUseCase,
		logger:          logger,
		validator:       validator.New(),
	}
}

func (c *CustomerController) Authorize(ctx *fiber.Ctx) error {
	c.logger.Tracef("Authorize controller")
	return utils.WriteResponse(ctx, fiber.StatusOK, nil, "Authorized successfully", nil)
}

func (c *CustomerController) Register(ctx *fiber.Ctx) error {
	var request model.CreateCustomerRequest
	if err := ctx.BodyParser(&request); err != nil {
		c.logger.Error("Failed to parse body: ", err)
		return utils.WriteErrorResponse(ctx, fiber.StatusBadRequest, "Invalid request body")
	}

	if err := c.validator.Struct(request); err != nil {
		c.logger.Error("Validation failed: ", err)
		return utils.WriteErrorResponse(ctx, fiber.StatusBadRequest, err.Error())
	}

	customer, err := c.customerUseCase.Register(&request)
	if err != nil {
		c.logger.Error("Failed to register customer: ", err)
		return utils.WriteErrorResponse(ctx, fiber.StatusInternalServerError, err.Error())
	}

	// Generate JWT token
	token, err := utils.GenerateToken(customer.ID, customer.Email, customer.Name, customer.Role)
	if err != nil {
		c.logger.Error("Failed to generate token: ", err)
		return utils.WriteErrorResponse(ctx, fiber.StatusInternalServerError, "Failed to generate token")
	}

	return utils.WriteResponse(ctx, fiber.StatusCreated, token, "Customer registered successfully", nil)
}

func (c *CustomerController) Login(ctx *fiber.Ctx) error {
	var request model.LoginRequest
	if err := ctx.BodyParser(&request); err != nil {
		c.logger.Error("Failed to parse body: ", err.Error())
		return utils.WriteErrorResponse(ctx, fiber.StatusBadRequest, "Invalid request body")
	}

	if err := c.validator.Struct(request); err != nil {
		c.logger.Error("Validation failed: ", err)
		return utils.WriteErrorResponse(ctx, fiber.StatusBadRequest, "Invalid request body")
	}

	token, err := c.customerUseCase.Login(&request)
	if err != nil {
		c.logger.Error("Failed to login: ", err)
		return utils.WriteErrorResponse(ctx, fiber.StatusInternalServerError, "Failed to login")
	}

	return utils.WriteResponse(ctx, fiber.StatusOK, token, "Login successful", nil)
}

func (c *CustomerController) UpdateProfile(ctx *fiber.Ctx) error {
	customerID, err := strconv.ParseInt(ctx.Params("id"), 10, 64)
	if err != nil {
		c.logger.Error("Failed to parse customer ID: ", err)
		return utils.WriteErrorResponse(ctx, fiber.StatusBadRequest, "Invalid customer ID")
	}

	var request model.CreateCustomerRequest
	if err := ctx.BodyParser(&request); err != nil {
		c.logger.Error("Failed to parse body: ", err)
		return utils.WriteErrorResponse(ctx, fiber.StatusBadRequest, "Invalid request body")
	}

	if err := c.validator.Struct(request); err != nil {
		c.logger.Error("Validation failed: ", err)
		return utils.WriteErrorResponse(ctx, fiber.StatusBadRequest, err.Error())
	}

	if err := c.customerUseCase.UpdateCustomer(customerID, &request); err != nil {
		c.logger.Error("Failed to update profile: ", err)
		return utils.WriteErrorResponse(ctx, fiber.StatusInternalServerError, "Failed to update profile")
	}

	return utils.WriteResponse(ctx, fiber.StatusOK, nil, "Profile updated successfully", nil)
}

func (c *CustomerController) GetCustomerByID(ctx *fiber.Ctx) error {
	customerIDStr := ctx.Locals(constants.ClaimsKeyID)

	customerID, ok := customerIDStr.(int64)
	if !ok {
		c.logger.Infof("customer_id: %+v, type: %T", customerIDStr, customerIDStr)
		c.logger.Error("Failed to parse customer ID: ")
		return utils.WriteErrorResponse(ctx, fiber.StatusBadRequest, "Invalid customer ID")
	}

	customer, err := c.customerUseCase.GetCustomerByID(customerID)
	if err != nil {
		c.logger.Error("Failed to get customer: ", err)
		return utils.WriteErrorResponse(ctx, fiber.StatusInternalServerError, "Failed to get customer")
	}

	return utils.WriteResponse(ctx, fiber.StatusOK, model.CustomerToResponse(customer), "Customer fetched successfully", nil)
}
