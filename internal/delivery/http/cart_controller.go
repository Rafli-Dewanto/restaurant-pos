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
		utils.WriteErrorResponse(ctx, fiber.StatusBadRequest, err.Error())
	}

	err := c.cartUseCase.CreateCart(customerID, &req)
	if err != nil {
		c.logger.Errorf("❌ Failed to create cart: %v", err)
		utils.WriteErrorResponse(ctx, fiber.StatusInternalServerError, err.Error())
		return nil
	}

	utils.WriteResponse(ctx, fiber.StatusCreated, nil, "Cart created successfully", nil)
	return nil
}

func (c *CartController) GetCartByID(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	cartID, err := strconv.Atoi(id)
	if err != nil {
		c.logger.Errorf("❌ Failed to parse cart ID: %v", err)
		utils.WriteErrorResponse(ctx, fiber.StatusBadRequest, err.Error())
		return nil
	}

	cart, err := c.cartUseCase.GetCartByID(cartID)
	if err != nil {
		c.logger.Errorf("❌ Failed to fetch cart: %v", err)
		utils.WriteErrorResponse(ctx, fiber.StatusNotFound, err.Error())
		return nil
	}

	utils.WriteResponse(ctx, fiber.StatusOK, cart, "Cart fetched successfully", nil)
	return nil
}

func (c *CartController) GetCartByCustomerID(ctx *fiber.Ctx) error {
	customerID := ctx.Locals(constants.ClaimsKeyID).(int)

	params := new(model.PaginationQuery)
	if err := ctx.QueryParser(params); err != nil {
		c.logger.Errorf("❌ Failed to parse query params: %v", err)
		utils.WriteErrorResponse(ctx, fiber.StatusBadRequest, err.Error())
		return nil
	}

	data, err := c.cartUseCase.GetCartByCustomerID(customerID, params)
	if err != nil {
		c.logger.Errorf("❌ Failed to fetch carts: %v", err)
		utils.WriteErrorResponse(ctx, fiber.StatusInternalServerError, err.Error())
		return nil
	}
	metaData := model.ToPaginatedMeta(data)

	utils.WriteResponse(ctx, fiber.StatusOK, data.Data, "Carts fetched successfully", metaData)
	return nil
}
