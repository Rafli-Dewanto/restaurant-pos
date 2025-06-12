package controller

import (
	"cakestore/internal/constants"
	"cakestore/internal/domain/model"
	"cakestore/internal/usecase"
	"cakestore/utils"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

type ReservationController struct {
	useCase usecase.ReservationUseCase
	logger  *logrus.Logger
}

func NewReservationController(useCase usecase.ReservationUseCase, logger *logrus.Logger) *ReservationController {
	return &ReservationController{
		useCase: useCase,
		logger:  logger,
	}
}

func (c *ReservationController) CreateReservation(ctx *fiber.Ctx) error {
	customerID := ctx.Locals(constants.ClaimsKeyID).(int64)

	var request model.CreateReservationRequest
	if err := ctx.BodyParser(&request); err != nil {
		c.logger.Errorf("Error parsing request body: %v", err)
		return utils.WriteErrorResponse(ctx, fiber.StatusBadRequest, "Invalid request body")
	}

	reservation, err := c.useCase.Create(uint(customerID), &request)
	if err != nil {
		c.logger.Errorf("Error creating reservation: %v", err)
		return utils.WriteErrorResponse(ctx, fiber.StatusInternalServerError, "Failed to create reservation")
	}

	return ctx.Status(fiber.StatusCreated).JSON(utils.Response{
		Message: "Reservation created successfully",
		Data:    reservation,
	})
}

func (c *ReservationController) GetReservationByID(ctx *fiber.Ctx) error {
	id, err := strconv.ParseUint(ctx.Params("id"), 10, 32)
	if err != nil {
		c.logger.Errorf("Error parsing reservation ID: %v", err)
		return utils.WriteErrorResponse(ctx, fiber.StatusBadRequest, "Invalid reservation ID")
	}

	reservation, err := c.useCase.GetByID(uint(id))
	if err != nil {
		c.logger.Errorf("Error getting reservation: %v", err)
		return utils.WriteErrorResponse(ctx, fiber.StatusNotFound, "Reservation not found")
	}

	return ctx.JSON(utils.Response{
		Message: "Reservation retrieved successfully",
		Data:    reservation,
	})
}

func (c *ReservationController) GetAllReservations(ctx *fiber.Ctx) error {
	params := new(model.ReservationQueryParams)

	// Parse pagination parameters
	page, _ := strconv.Atoi(ctx.Query("page", "1"))
	perPage, _ := strconv.Atoi(ctx.Query("per_page", "10"))
	params.Page = int64(page)
	params.Limit = int64(perPage)

	// Parse other query parameters
	if customerID := ctx.Query("customer_id"); customerID != "" {
		id, err := strconv.ParseUint(customerID, 10, 32)
		if err == nil {
			params.CustomerID = uint(id)
		}
	}

	if status := ctx.Query("status"); status != "" {
		params.Status = status
	}

	if reserveDate := ctx.Query("reserve_date"); reserveDate != "" {
		date, err := time.Parse(time.RFC3339, reserveDate)
		if err == nil {
			params.ReserveDate = date
		}
	}

	if tableNumber := ctx.Query("table_number"); tableNumber != "" {
		num, err := strconv.Atoi(tableNumber)
		if err == nil {
			params.TableNumber = num
		}
	}

	reservations, err := c.useCase.GetAll(params)
	if err != nil {
		c.logger.Errorf("Error getting reservations: %v", err)
		return utils.WriteErrorResponse(ctx, fiber.StatusInternalServerError, "Failed to get reservations")
	}

	return utils.WriteResponse(ctx, fiber.StatusOK, reservations.Data, "Reservations retrieved successfully", model.ToPaginatedMeta(reservations))
}

func (c *ReservationController) UpdateReservation(ctx *fiber.Ctx) error {
	id, err := strconv.ParseUint(ctx.Params("id"), 10, 32)
	if err != nil {
		c.logger.Errorf("Error parsing reservation ID: %v", err)
		return utils.WriteErrorResponse(ctx, fiber.StatusBadRequest, "Invalid reservation ID")
	}

	var request model.UpdateReservationRequest
	if err := ctx.BodyParser(&request); err != nil {
		c.logger.Errorf("Error parsing request body: %v", err)
		return utils.WriteErrorResponse(ctx, fiber.StatusBadRequest, "Invalid request body")
	}

	reservation, err := c.useCase.Update(uint(id), &request)
	if err != nil {
		c.logger.Errorf("Error updating reservation: %v", err)
		return utils.WriteErrorResponse(ctx, fiber.StatusInternalServerError, "Failed to update reservation")
	}

	return ctx.JSON(utils.Response{
		Message: "Reservation updated successfully",
		Data:    reservation,
	})
}

func (c *ReservationController) DeleteReservation(ctx *fiber.Ctx) error {
	id, err := strconv.ParseUint(ctx.Params("id"), 10, 32)
	if err != nil {
		c.logger.Errorf("Error parsing reservation ID: %v", err)
		return utils.WriteErrorResponse(ctx, fiber.StatusBadRequest, "Invalid reservation ID")
	}

	if err := c.useCase.Delete(uint(id)); err != nil {
		c.logger.Errorf("Error deleting reservation: %v", err)
		return utils.WriteErrorResponse(ctx, fiber.StatusInternalServerError, "Failed to delete reservation")
	}

	return ctx.JSON(utils.Response{
		Message: "Reservation deleted successfully",
	})
}
