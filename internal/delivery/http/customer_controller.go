package controller

import (
	"cakestore/internal/model"
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

func (c *CustomerController) Register(ctx *fiber.Ctx) error {
	var request model.CreateCustomerRequest
	if err := ctx.BodyParser(&request); err != nil {
		c.logger.Error("Failed to parse body: ", err)
		return model.WriteErrorResponse(ctx, fiber.StatusBadRequest, "Invalid request body")
	}

	c.logger.Debug("Request: ", request)
	if err := c.validator.Struct(request); err != nil {
		c.logger.Error("Validation failed: ", err)
		return model.WriteErrorResponse(ctx, fiber.StatusBadRequest, err.Error())
	}

	customer, err := c.customerUseCase.Register(&request)
	if err != nil {
		c.logger.Error("Failed to register customer: ", err)
		return model.WriteErrorResponse(ctx, fiber.StatusInternalServerError, err.Error())
	}

	// Generate JWT token
	token, err := utils.GenerateToken(customer.ID, customer.Email, customer.Name)
	if err != nil {
		c.logger.Error("Failed to generate token: ", err)
		return model.WriteErrorResponse(ctx, fiber.StatusInternalServerError, "Failed to generate token")
	}

	model.WriteResponse(ctx, fiber.StatusCreated, token, "Customer registered successfully", nil)
	return nil
}

func (c *CustomerController) Login(ctx *fiber.Ctx) error {
	var request model.LoginRequest
	if err := ctx.BodyParser(&request); err != nil {
		c.logger.Error("Failed to parse body: ", err.Error())
		return model.WriteErrorResponse(ctx, fiber.StatusBadRequest, "Invalid request body")
	}

	if err := c.validator.Struct(request); err != nil {
		c.logger.Error("Validation failed: ", err)
		model.WriteErrorResponse(ctx, fiber.StatusBadRequest, "Invalid request body")
		return nil
	}

	token, err := c.customerUseCase.Login(&request)
	if err != nil {
		c.logger.Error("Failed to login: ", err)
		return model.WriteErrorResponse(ctx, fiber.StatusInternalServerError, "Failed to login")
	}

	model.WriteResponse(ctx, fiber.StatusOK, token, "Login successful", nil)
	return nil
}

func (c *CustomerController) UpdateProfile(ctx *fiber.Ctx) error {
	customerID, err := strconv.Atoi(ctx.Params("id"))
	if err != nil {
		c.logger.Error("Failed to parse customer ID: ", err)
		model.WriteErrorResponse(ctx, fiber.StatusBadRequest, "Invalid customer ID")
	}

	var request model.CreateCustomerRequest
	if err := ctx.BodyParser(&request); err != nil {
		c.logger.Error("Failed to parse body: ", err)
		model.WriteErrorResponse(ctx, fiber.StatusBadRequest, "Invalid request body")
	}

	if err := c.validator.Struct(request); err != nil {
		c.logger.Error("Validation failed: ", err)
		model.WriteErrorResponse(ctx, fiber.StatusBadRequest, err.Error())
	}

	if err := c.customerUseCase.UpdateCustomer(customerID, &request); err != nil {
		c.logger.Error("Failed to update profile: ", err)
		model.WriteErrorResponse(ctx, fiber.StatusInternalServerError, "Failed to update profile")
	}

	model.WriteResponse(ctx, fiber.StatusOK, nil, "Profile updated successfully", nil)
	return nil
}
