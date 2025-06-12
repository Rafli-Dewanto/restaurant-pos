package controller

import (
	"cakestore/internal/domain/model"
	"cakestore/internal/usecase"
	"cakestore/utils"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

type InventoryController struct {
	useCase usecase.InventoryUseCase
	logger  *logrus.Logger
}

func NewInventoryController(useCase usecase.InventoryUseCase, logger *logrus.Logger) *InventoryController {
	return &InventoryController{
		useCase: useCase,
		logger:  logger,
	}
}

func (c *InventoryController) CreateInventory(ctx *fiber.Ctx) error {
	var request model.CreateInventoryRequest
	if err := ctx.BodyParser(&request); err != nil {
		c.logger.Errorf("Error parsing request body: %v", err)
		return utils.WriteErrorResponse(ctx, fiber.StatusBadRequest, "Invalid request body")
	}

	ingredient, err := c.useCase.Create(&request)
	if err != nil {
		c.logger.Errorf("Error creating ingredient: %v", err)
		return utils.WriteErrorResponse(ctx, fiber.StatusInternalServerError, "Failed to create ingredient")
	}

	return ctx.Status(fiber.StatusCreated).JSON(utils.Response{
		Message: "Ingredient created successfully",
		Data:    ingredient,
	})
}

func (c *InventoryController) GetInventoryByID(ctx *fiber.Ctx) error {
	id, err := strconv.ParseUint(ctx.Params("id"), 10, 32)
	if err != nil {
		c.logger.Errorf("Error parsing ingredient ID: %v", err)
		return utils.WriteErrorResponse(ctx, fiber.StatusBadRequest, "Invalid ingredient ID")
	}

	ingredient, err := c.useCase.GetByID(uint(id))
	if err != nil {
		c.logger.Errorf("Error getting ingredient: %v", err)
		return utils.WriteErrorResponse(ctx, fiber.StatusNotFound, "Ingredient not found")
	}

	return utils.WriteResponse(ctx, fiber.StatusOK, ingredient, "Ingredient retrieved successfully", nil)
}

func (c *InventoryController) GetAllInventories(ctx *fiber.Ctx) error {
	params := new(model.InventoryQueryParams)

	page, _ := strconv.Atoi(ctx.Query("page", "1"))
	perPage, _ := strconv.Atoi(ctx.Query("per_page", "10"))
	params.Page = int64(page)
	params.Limit = int64(perPage)
	params.Search = ctx.Query("search")

	ingredients, err := c.useCase.GetAll(params)
	if err != nil {
		c.logger.Errorf("Error getting ingredients: %v", err)
		return utils.WriteErrorResponse(ctx, fiber.StatusInternalServerError, "Failed to get ingredients")
	}

	return utils.WriteResponse(ctx, fiber.StatusOK, ingredients.Data, "Ingredients retrieved successfully", model.ToPaginatedMeta(ingredients))
}

func (c *InventoryController) UpdateInventory(ctx *fiber.Ctx) error {
	id, err := strconv.ParseUint(ctx.Params("id"), 10, 32)
	if err != nil {
		c.logger.Errorf("Error parsing ingredient ID: %v", err)
		return utils.WriteErrorResponse(ctx, fiber.StatusBadRequest, "Invalid ingredient ID")
	}

	var request model.UpdateInventoryRequest
	if err := ctx.BodyParser(&request); err != nil {
		c.logger.Errorf("Error parsing request body: %v", err)
		return utils.WriteErrorResponse(ctx, fiber.StatusBadRequest, "Invalid request body")
	}

	ingredient, err := c.useCase.Update(uint(id), &request)
	if err != nil {
		c.logger.Errorf("Error updating ingredient: %v", err)
		return utils.WriteErrorResponse(ctx, fiber.StatusInternalServerError, "Failed to update ingredient")
	}

	return utils.WriteResponse(ctx, fiber.StatusOK, ingredient, "Ingredient updated successfully", nil)
}

func (c *InventoryController) DeleteInventory(ctx *fiber.Ctx) error {
	id, err := strconv.ParseUint(ctx.Params("id"), 10, 32)
	if err != nil {
		c.logger.Errorf("Error parsing ingredient ID: %v", err)
		return utils.WriteErrorResponse(ctx, fiber.StatusBadRequest, "Invalid ingredient ID")
	}

	if err := c.useCase.Delete(uint(id)); err != nil {
		c.logger.Errorf("Error deleting ingredient: %v", err)
		return utils.WriteErrorResponse(ctx, fiber.StatusInternalServerError, "Failed to delete ingredient")
	}

	return utils.WriteResponse(ctx, fiber.StatusOK, nil, "Ingredient deleted successfully", nil)
}

func (c *InventoryController) UpdateInventoryStock(ctx *fiber.Ctx) error {
	id, err := strconv.ParseUint(ctx.Params("id"), 10, 32)
	if err != nil {
		c.logger.Errorf("Error parsing ingredient ID: %v", err)
		return utils.WriteErrorResponse(ctx, fiber.StatusBadRequest, "Invalid ingredient ID")
	}

	var request struct {
		Quantity float64 `json:"quantity" validate:"required"`
	}
	if err := ctx.BodyParser(&request); err != nil {
		c.logger.Errorf("Error parsing request body: %v", err)
		return utils.WriteErrorResponse(ctx, fiber.StatusBadRequest, "Invalid request body")
	}

	if err := c.useCase.UpdateStock(uint(id), request.Quantity); err != nil {
		c.logger.Errorf("Error updating ingredient stock: %v", err)
		return utils.WriteErrorResponse(ctx, fiber.StatusInternalServerError, "Failed to update ingredient stock")
	}

	return utils.WriteResponse(ctx, fiber.StatusOK, nil, "Ingredient stock updated successfully", nil)
}

func (c *InventoryController) GetLowStockInventories(ctx *fiber.Ctx) error {
	c.logger.Info("HIT")
	ingredients, err := c.useCase.GetLowStockIngredients()
	c.logger.Info(ingredients)
	if err != nil {
		c.logger.Errorf("Error getting low stock ingredients: %v", err)
		return utils.WriteErrorResponse(ctx, fiber.StatusInternalServerError, "Failed to get low stock ingredients")
	}

	return utils.WriteResponse(ctx, fiber.StatusOK, ingredients, "Low stock ingredients retrieved successfully", nil)
}
