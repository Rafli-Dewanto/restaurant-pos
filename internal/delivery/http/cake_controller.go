package controller

import (
	"cakestore/internal/entity"
	"cakestore/internal/model"
	"cakestore/internal/usecase"
	"database/sql"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

type CakeController struct {
	cakeUseCase usecase.CakeUseCase
	logger      *logrus.Logger
	validator   *validator.Validate
}

func NewCakeController(cakeUseCase usecase.CakeUseCase, logger *logrus.Logger) *CakeController {
	return &CakeController{
		cakeUseCase: cakeUseCase,
		logger:      logger,
		validator:   validator.New(),
	}
}

func (c *CakeController) GetAllCakes(ctx *fiber.Ctx) error {
	cakes, err := c.cakeUseCase.GetAllCakes()
	if err != nil {
		c.logger.Errorf("Failed to fetch cakes: %v", err)
		model.WriteErrorResponse(ctx, fiber.StatusInternalServerError, "Failed to fetch cakes")
	}

	cakeResponses := make([]*model.CakeModel, len(cakes))
	for i, cake := range cakes {
		cakeResponses[i] = model.CakeToResponse(&cake)
	}

	model.WriteResponse(ctx, fiber.StatusOK, cakeResponses, "Success", nil)
	return nil
}

func (c *CakeController) GetCakeByID(ctx *fiber.Ctx) error {
	id, err := strconv.Atoi(ctx.Params("id"))
	if err != nil {
		c.logger.Error("Failed to parse cake ID: ", err)
		model.WriteErrorResponse(ctx, fiber.StatusBadRequest, "Invalid cake ID")
	}

	cake, err := c.cakeUseCase.GetCakeByID(id)
	if err != nil {
		c.logger.Errorf("Failed to get cake: %v", err)
		model.WriteErrorResponse(ctx, fiber.StatusInternalServerError, "Failed to get cake")
	}

	model.WriteResponse(ctx, fiber.StatusOK, model.CakeToResponse(cake), "Success", nil)
	return nil
}

func (c *CakeController) CreateCake(ctx *fiber.Ctx) error {
	var request model.CreateUpdateCakeRequest
	if err := ctx.BodyParser(&request); err != nil {
		c.logger.Error("Failed to parse body: ", err)
		model.WriteErrorResponse(ctx, fiber.StatusBadRequest, "Invalid request body")
	}

	if err := c.validatePayload(request); err != nil {
		c.logger.Error("Validation failed: ", err)
		model.WriteErrorResponse(ctx, fiber.StatusBadRequest, err.Error())
	}

	cake := &entity.Cake{
		Title:       request.Title,
		Description: request.Description,
		Rating:      float64(request.Rating),
		Image:       request.ImageURL,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		DeletedAt:   sql.NullTime{},
	}

	if err := c.cakeUseCase.CreateCake(cake); err != nil {
		c.logger.Error("Failed to create cake: ", err)
		model.WriteErrorResponse(ctx, fiber.StatusInternalServerError, "Failed to create cake")
	}

	model.WriteResponse(ctx, fiber.StatusCreated, model.CakeToResponse(cake), "Success", nil)
	return nil
}

func (c *CakeController) UpdateCake(ctx *fiber.Ctx) error {

	var request model.CreateUpdateCakeRequest
	if err := ctx.BodyParser(&request); err != nil {
		model.WriteErrorResponse(ctx, fiber.StatusBadRequest, "Invalid request body")
	}

	if err := c.validatePayload(request); err != nil {
		model.WriteErrorResponse(ctx, fiber.StatusBadRequest, err.Error())
	}

	cakeID, err := strconv.Atoi(ctx.Params("id"))
	if err != nil {
		model.WriteErrorResponse(ctx, fiber.StatusBadRequest, "Invalid cake ID")
	}

	cake := &entity.Cake{
		ID:          cakeID,
		Title:       request.Title,
		Description: request.Description,
		Rating:      float64(request.Rating),
		Image:       request.ImageURL,
		UpdatedAt:   time.Now(),
	}

	if err := c.cakeUseCase.UpdateCake(cake); err != nil {
		c.logger.Error("Failed to update cake: ", err)
		model.WriteErrorResponse(ctx, fiber.StatusInternalServerError, "Failed to update cake")
	}

	resp := model.CakeToResponse(cake)
	model.WriteResponse(ctx, fiber.StatusOK, resp, "Success", nil)
	return nil
}

func (c *CakeController) DeleteCake(ctx *fiber.Ctx) error {
	id, err := strconv.Atoi(ctx.Params("id"))
	if err != nil {
		c.logger.Error("Failed to parse cake ID: ", err)
		model.WriteErrorResponse(ctx, fiber.StatusBadRequest, "Invalid cake ID")
	}

	err = c.cakeUseCase.SoftDeleteCake(id)
	if err != nil {
		c.logger.Error("Failed to delete cake: ", err)
		model.WriteErrorResponse(ctx, fiber.StatusInternalServerError, "Failed to delete cake")
	}

	model.WriteResponse(ctx, fiber.StatusOK, nil, "Success", nil)
	return nil
}

func (c *CakeController) validatePayload(request model.CreateUpdateCakeRequest) error {
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
