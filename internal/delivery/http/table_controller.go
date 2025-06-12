package controller

import (
	"cakestore/internal/domain/model"
	"cakestore/internal/usecase"
	"cakestore/utils"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

type TableController struct {
	useCase usecase.TableUseCase
	logger  *logrus.Logger
}

func NewTableController(useCase usecase.TableUseCase, logger *logrus.Logger) *TableController {
	return &TableController{
		useCase: useCase,
		logger:  logger,
	}
}

func (c *TableController) CreateTable(ctx *fiber.Ctx) error {
	var request model.CreateTableRequest
	if err := ctx.BodyParser(&request); err != nil {
		c.logger.Errorf("Error parsing request body: %v", err)
		return utils.WriteErrorResponse(ctx, fiber.StatusBadRequest, "Invalid request body")
	}

	table, err := c.useCase.Create(&request)
	if err != nil {
		c.logger.Errorf("Error creating table: %v", err)
		return utils.WriteErrorResponse(ctx, fiber.StatusInternalServerError, "Failed to create table")
	}

	return ctx.Status(fiber.StatusCreated).JSON(utils.Response{
		Message: "Table created successfully",
		Data:    table,
	})
}

func (c *TableController) GetTableByID(ctx *fiber.Ctx) error {
	id, err := strconv.ParseUint(ctx.Params("id"), 10, 32)
	if err != nil {
		c.logger.Errorf("Error parsing table ID: %v", err)
		return utils.WriteErrorResponse(ctx, fiber.StatusBadRequest, "Invalid table ID")
	}

	table, err := c.useCase.GetByID(uint(id))
	if err != nil {
		c.logger.Errorf("Error getting table: %v", err)
		return utils.WriteErrorResponse(ctx, fiber.StatusNotFound, "Table not found")
	}

	return ctx.JSON(utils.Response{
		Message: "Table retrieved successfully",
		Data:    table,
	})
}

func (c *TableController) GetAllTables(ctx *fiber.Ctx) error {
	params := new(model.TableQueryParams)

	// Parse pagination parameters
	page, _ := strconv.Atoi(ctx.Query("page", "1"))
	perPage, _ := strconv.Atoi(ctx.Query("per_page", "10"))
	params.Page = int64(page)
	params.Limit = int64(perPage)

	// Parse other query parameters
	if capacity := ctx.Query("capacity"); capacity != "" {
		cap, err := strconv.Atoi(capacity)
		if err == nil {
			params.Capacity = cap
		}
	}

	if isAvailable := ctx.Query("is_available"); isAvailable != "" {
		available, err := strconv.ParseBool(isAvailable)
		if err == nil {
			params.IsAvailable = &available
		}
	}

	tables, err := c.useCase.GetAll(params)
	if err != nil {
		c.logger.Errorf("Error getting tables: %v", err)
		return utils.WriteErrorResponse(ctx, fiber.StatusInternalServerError, "Failed to get tables")
	}

	return utils.WriteResponse(ctx, fiber.StatusOK, tables.Data, "Tables retrieved successfully", model.ToPaginatedMeta(tables))
}

func (c *TableController) UpdateTable(ctx *fiber.Ctx) error {
	id, err := strconv.ParseUint(ctx.Params("id"), 10, 32)
	if err != nil {
		c.logger.Errorf("Error parsing table ID: %v", err)
		return utils.WriteErrorResponse(ctx, fiber.StatusBadRequest, "Invalid table ID")
	}

	var request model.UpdateTableRequest
	if err := ctx.BodyParser(&request); err != nil {
		c.logger.Errorf("Error parsing request body: %v", err)
		return utils.WriteErrorResponse(ctx, fiber.StatusBadRequest, "Invalid request body")
	}

	table, err := c.useCase.Update(uint(id), &request)
	if err != nil {
		c.logger.Errorf("Error updating table: %v", err)
		return utils.WriteErrorResponse(ctx, fiber.StatusInternalServerError, "Failed to update table")
	}

	return ctx.JSON(utils.Response{
		Message: "Table updated successfully",
		Data:    table,
	})
}

func (c *TableController) DeleteTable(ctx *fiber.Ctx) error {
	id, err := strconv.ParseUint(ctx.Params("id"), 10, 32)
	if err != nil {
		c.logger.Errorf("Error parsing table ID: %v", err)
		return utils.WriteErrorResponse(ctx, fiber.StatusBadRequest, "Invalid table ID")
	}

	if err := c.useCase.Delete(uint(id)); err != nil {
		c.logger.Errorf("Error deleting table: %v", err)
		return utils.WriteErrorResponse(ctx, fiber.StatusInternalServerError, "Failed to delete table")
	}

	return ctx.JSON(utils.Response{
		Message: "Table deleted successfully",
	})
}

func (c *TableController) GetAvailableTables(ctx *fiber.Ctx) error {
	// Parse query parameters
	reserveTimeStr := ctx.Query("reserve_time")
	durationStr := ctx.Query("duration")

	if reserveTimeStr == "" || durationStr == "" {
		return utils.WriteErrorResponse(ctx, fiber.StatusBadRequest, "Reserve time and duration are required")
	}

	reserveTime, err := time.Parse(time.RFC3339, reserveTimeStr)
	if err != nil {
		c.logger.Errorf("Error parsing reserve time: %v", err)
		return utils.WriteErrorResponse(ctx, fiber.StatusBadRequest, "Invalid reserve time format")
	}

	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		c.logger.Errorf("Error parsing duration: %v", err)
		return utils.WriteErrorResponse(ctx, fiber.StatusBadRequest, "Invalid duration format")
	}

	tables, err := c.useCase.GetAvailableTables(reserveTime, duration)
	if err != nil {
		c.logger.Errorf("Error getting available tables: %v", err)
		return utils.WriteErrorResponse(ctx, fiber.StatusInternalServerError, "Failed to get available tables")
	}

	return ctx.JSON(utils.Response{
		Message: "Available tables retrieved successfully",
		Data:    tables,
	})
}

func (c *TableController) UpdateTableAvailability(ctx *fiber.Ctx) error {
	id, err := strconv.ParseUint(ctx.Params("id"), 10, 32)
	if err != nil {
		c.logger.Errorf("Error parsing table ID: %v", err)
		return utils.WriteErrorResponse(ctx, fiber.StatusBadRequest, "Invalid table ID")
	}

	var request struct {
		IsAvailable bool `json:"is_available"`
	}
	if err := ctx.BodyParser(&request); err != nil {
		c.logger.Errorf("Error parsing request body: %v", err)
		return utils.WriteErrorResponse(ctx, fiber.StatusBadRequest, "Invalid request body")
	}

	if err := c.useCase.UpdateAvailability(uint(id), request.IsAvailable); err != nil {
		c.logger.Errorf("Error updating table availability: %v", err)
		return utils.WriteErrorResponse(ctx, fiber.StatusInternalServerError, "Failed to update table availability")
	}

	return ctx.JSON(utils.Response{
		Message: "Table availability updated successfully",
	})
}
