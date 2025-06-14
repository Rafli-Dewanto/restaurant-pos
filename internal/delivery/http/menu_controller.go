package controller

import (
	"cakestore/internal/constants"
	"cakestore/internal/domain/entity"
	"cakestore/internal/domain/model"
	"cakestore/internal/usecase"
	"cakestore/utils"
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

type MenuController struct {
	menuUseCase usecase.MenuUseCase
	logger      *logrus.Logger
	validator   *validator.Validate
}

func NewMenuController(menuUseCase usecase.MenuUseCase, logger *logrus.Logger) *MenuController {
	return &MenuController{
		menuUseCase: menuUseCase,
		logger:      logger,
		validator:   validator.New(),
	}
}

func (c *MenuController) GetAllMenus(ctx *fiber.Ctx) error {
	var params model.MenuQueryParams
	if err := ctx.QueryParser(&params); err != nil {
		c.logger.Error("Failed to parse query params: ", err)
		return utils.WriteErrorResponse(ctx, fiber.StatusBadRequest, "Invalid query params")
	}

	menus, err := c.menuUseCase.GetAllMenus(&params)
	if err != nil {
		c.logger.Errorf("Failed to fetch menus: %v", err)
		return utils.WriteErrorResponse(ctx, fiber.StatusInternalServerError, "Failed to fetch menus")
	}

	metaData := &model.PaginatedMeta{
		CurrentPage: int64(menus.Page),
		Total:       menus.Total,
		PerPage:     int64(menus.PageSize),
		LastPage:    menus.TotalPages,
		HasNextPage: menus.Page < menus.TotalPages,
		HasPrevPage: menus.Page > 1,
	}

	menuResponses := make([]*model.MenuModel, len(menus.Data))
	for i, menu := range menus.Data {
		menuResponses[i] = model.ToMenuResponse(&menu)
	}

	return utils.WriteResponse(ctx, fiber.StatusOK, menuResponses, "Success", metaData)
}

func (c *MenuController) GetMenuByID(ctx *fiber.Ctx) error {
	id, err := strconv.ParseInt(ctx.Params("id"), 10, 64)
	if err != nil {
		c.logger.Error("Failed to parse menu ID: ", err)
		return utils.WriteErrorResponse(ctx, fiber.StatusBadRequest, "Invalid menu ID")
	}

	menu, err := c.menuUseCase.GetMenuByID(id)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return utils.WriteErrorResponse(ctx, fiber.StatusNotFound, "Menu not found")
		}
		c.logger.Errorf("Failed to get menu: %v", err)
		return utils.WriteErrorResponse(ctx, fiber.StatusInternalServerError, "Failed to get menu")
	}

	return utils.WriteResponse(ctx, fiber.StatusOK, model.ToMenuResponse(menu), "Success", nil)
}

func (c *MenuController) CreateMenu(ctx *fiber.Ctx) error {
	var request model.CreateUpdateMenuRequest
	if err := ctx.BodyParser(&request); err != nil {
		c.logger.Error("Failed to parse body: ", err)
		return utils.WriteErrorResponse(ctx, fiber.StatusBadRequest, "Invalid request body")
	}

	if err := c.validatePayload(request); err != nil {
		c.logger.Error("Validation failed: ", err)
		return utils.WriteErrorResponse(ctx, fiber.StatusBadRequest, err.Error())
	}

	menu := &entity.Menu{
		Title:       request.Title,
		Description: request.Description,
		Rating:      float64(request.Rating),
		Image:       request.ImageURL,
		Price:       request.Price,
		Quantity:    request.Quantity,
		Category:    request.Category,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		DeletedAt:   sql.NullTime{},
	}

	if err := c.menuUseCase.CreateMenu(menu); err != nil {
		c.logger.Error("Failed to create menu: ", err)
		return utils.WriteErrorResponse(ctx, fiber.StatusInternalServerError, "Failed to create menu")
	}

	return utils.WriteResponse(ctx, fiber.StatusCreated, model.ToMenuResponse(menu), "Success", nil)
}

func (c *MenuController) UpdateMenu(ctx *fiber.Ctx) error {
	var request model.CreateUpdateMenuRequest
	if err := ctx.BodyParser(&request); err != nil {
		c.logger.Error("Failed to parse body: ", err)
		return utils.WriteErrorResponse(ctx, fiber.StatusBadRequest, "Invalid request body")
	}

	if err := c.validatePayload(request); err != nil {
		c.logger.Error("Validation failed: ", err)
		return utils.WriteErrorResponse(ctx, fiber.StatusBadRequest, err.Error())
	}

	menuID, err := strconv.ParseInt(ctx.Params("id"), 10, 64)
	if err != nil {
		c.logger.Errorf("❌ Failed to parse menu ID: %v", err)
		return utils.WriteErrorResponse(ctx, fiber.StatusBadRequest, "Invalid menu ID")
	}

	menu := &entity.Menu{
		ID:          menuID,
		Title:       request.Title,
		Description: request.Description,
		Price:       request.Price,
		Quantity:    request.Quantity,
		Category:    request.Category,
		Rating:      float64(request.Rating),
		Image:       request.ImageURL,
		UpdatedAt:   time.Now(),
	}

	if err := c.menuUseCase.UpdateMenu(menu); err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return utils.WriteErrorResponse(ctx, fiber.StatusNotFound, "Menu not found")
		}
		c.logger.Error("Failed to update menu: ", err)
		return utils.WriteErrorResponse(ctx, fiber.StatusInternalServerError, "Failed to update menu: "+err.Error())
	}

	resp := model.ToMenuResponse(menu)
	return utils.WriteResponse(ctx, fiber.StatusOK, resp, "Success", nil)
}

func (c *MenuController) DeleteMenu(ctx *fiber.Ctx) error {
	id, err := strconv.ParseInt(ctx.Params("id"), 10, 64)
	if err != nil {
		c.logger.Error("Failed to parse menu ID: ", err)
		return utils.WriteErrorResponse(ctx, fiber.StatusBadRequest, "Invalid menu ID")
	}

	err = c.menuUseCase.SoftDeleteMenu(id)
	if err != nil {
		c.logger.Error("Failed to delete menu: ", err)
		return utils.WriteErrorResponse(ctx, fiber.StatusInternalServerError, "Failed to delete menu")
	}

	return utils.WriteResponse(ctx, fiber.StatusOK, nil, "Success", nil)
}

func (c *MenuController) validatePayload(request model.CreateUpdateMenuRequest) error {
	if err := c.validator.Struct(request); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		errMessages := make([]string, len(validationErrors))
		for i, e := range validationErrors {
			errMessages[i] = "Field '" + e.Field() + "' failed on '" + e.Tag() + "' rule"
		}
		return fiber.NewError(http.StatusBadRequest, "Validation failed: "+strings.Join(errMessages, ", "))
	}
	return nil
}
