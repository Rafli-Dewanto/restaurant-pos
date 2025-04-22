package controller

import (
	"cakestore/internal/constants"
	"cakestore/internal/domain/model"
	"cakestore/internal/usecase"
	"cakestore/utils"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

type CartController struct {
	cartUseCase usecase.CartUseCase
	logger      *logrus.Logger
}

func NewCartController(cartUseCase usecase.CartUseCase, logger *logrus.Logger) *CartController {
	return &CartController{
		cartUseCase: cartUseCase,
		logger:      logger,
	}
}

func (c *CartController) AddCart(ctx *fiber.Ctx) error {
	customerID := ctx.Locals(constants.ClaimsKeyID).(int)
	var req model.AddCart

	if err := ctx.BodyParser(&req); err != nil {
		c.logger.Errorf("❌ Failed to parse request body: %v", err)
		return utils.WriteErrorResponse(ctx, fiber.StatusBadRequest, err.Error())
	}

	err := c.cartUseCase.CreateCart(customerID, &req)
	if err != nil {
		c.logger.Errorf("❌ Failed to create cart: %v", err)
		return utils.WriteErrorResponse(ctx, fiber.StatusInternalServerError, err.Error())
	}

	return utils.WriteResponse(ctx, fiber.StatusCreated, nil, "Cart created successfully", nil)
}

func (c *CartController) GetCartByID(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	cartID, err := strconv.Atoi(id)
	if err != nil {
		c.logger.Errorf("❌ Failed to parse cart ID: %v", err)
		return utils.WriteErrorResponse(ctx, fiber.StatusBadRequest, err.Error())
	}

	cart, err := c.cartUseCase.GetCartByID(cartID)
	if err != nil {
		c.logger.Errorf("❌ Failed to fetch cart: %v", err)
		return utils.WriteErrorResponse(ctx, fiber.StatusNotFound, err.Error())
	}

	return utils.WriteResponse(ctx, fiber.StatusOK, cart, "Cart fetched successfully", nil)
}

func (c *CartController) GetCartByCustomerID(ctx *fiber.Ctx) error {
	customerID := ctx.Locals(constants.ClaimsKeyID).(int)

	params := new(model.PaginationQuery)
	if err := ctx.QueryParser(params); err != nil {
		c.logger.Errorf("❌ Failed to parse query params: %v", err)
		return utils.WriteErrorResponse(ctx, fiber.StatusBadRequest, err.Error())
	}

	data, err := c.cartUseCase.GetCartByCustomerID(customerID, params)
	if err != nil {
		c.logger.Errorf("❌ Failed to fetch carts: %v", err)
		return utils.WriteErrorResponse(ctx, fiber.StatusInternalServerError, err.Error())
	}
	metaData := model.ToPaginatedMeta(data)

	return utils.WriteResponse(ctx, fiber.StatusOK, data.Data, "Carts fetched successfully", metaData)
}

func (c *CartController) RemoveCart(ctx *fiber.Ctx) error {
	customerID := ctx.Locals(constants.ClaimsKeyID).(int)
	cartID, err := strconv.Atoi(ctx.Params("id"))
	if err != nil {
		c.logger.Errorf("❌ Failed to parse cart ID: %v", err)
		return utils.WriteErrorResponse(ctx, fiber.StatusBadRequest, err.Error())
	}

	err = c.cartUseCase.RemoveCart(customerID, cartID)
	if err != nil {
		c.logger.Errorf("❌ Failed to remove cart: %v", err)
		return utils.WriteErrorResponse(ctx, fiber.StatusInternalServerError, err.Error())
	}

	return utils.WriteResponse(ctx, fiber.StatusOK, nil, "Cart removed successfully", nil)
}

func (c *CartController) ClearCart(ctx *fiber.Ctx) error {
	customerID := ctx.Locals(constants.ClaimsKeyID).(int)

	err := c.cartUseCase.ClearCart(customerID)
	if err != nil {
		c.logger.Errorf("❌ Failed to clear cart: %v", err)
		return utils.WriteErrorResponse(ctx, fiber.StatusInternalServerError, err.Error())
	}

	return utils.WriteResponse(ctx, fiber.StatusOK, nil, "Cart cleared successfully", nil)
}
