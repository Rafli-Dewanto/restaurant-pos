package controller

import (
	"cakestore/internal/domain/entity"
	"cakestore/internal/domain/model"
	"cakestore/internal/usecase"
	"cakestore/utils"
	"net/http"
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

func (h *CartController) GetCart(c *fiber.Ctx) error {
	customerID := c.Locals("customerID")

	var params model.PaginationQuery
	if err := c.QueryParser(&params); err != nil {
		h.logger.Error("Failed to parse query params: ", err)
		return utils.WriteErrorResponse(c, fiber.StatusBadRequest, "Invalid query params")
	}

	// Set default values if not provided
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.Limit <= 0 {
		params.Limit = 10
	}

	if customerID == 0 {
		return utils.WriteErrorResponse(c, http.StatusUnauthorized, "unauthorized")
	}

	cart, err := h.cartUseCase.GetCartByCustomerID(customerID.(int), &params)
	if err != nil {
		h.logger.Errorf("Error getting cart: %v", err)
		return utils.WriteErrorResponse(c, http.StatusInternalServerError, err.Error())
	}

	return utils.WriteResponse(c, http.StatusOK, cart, "Success", nil)
}

func (h *CartController) AddItem(c *fiber.Ctx) error {
	h.logger.Info("AddItem")
	customerID := c.Locals("customer_id").(int)

	if customerID == 0 {
		return utils.WriteErrorResponse(c, http.StatusUnauthorized, "unauthorized")
	}
	var req model.AddCart

	if err := c.BodyParser(&req); err != nil {
		h.logger.Error("Failed to parse body: ", err)
		return utils.WriteErrorResponse(c, http.StatusBadRequest, "Invalid request body")
	}

	cart, err := h.cartUseCase.GetCartByCustomerID(customerID, nil)
	if err != nil {
		h.logger.Errorf("Error getting/creating cart: %v", err)
		return utils.WriteErrorResponse(c, http.StatusInternalServerError, err.Error())
	}
	cartData, ok := cart.Data.(entity.Cart)
	if !ok {
		return utils.WriteErrorResponse(c, http.StatusInternalServerError, "Invalid cart data")
	}

	if err := h.cartUseCase.AddItem(cartData.ID, req.CakeID, req.Quantity); err != nil {
		h.logger.Errorf("Error adding item to cart: %v", err)
		return utils.WriteErrorResponse(c, http.StatusInternalServerError, err.Error())
	}

	return utils.WriteResponse(c, http.StatusOK, nil, "Item added to cart successfully", nil)
}

func (h *CartController) UpdateItemQuantity(c *fiber.Ctx) error {
	customerID := c.Locals("customerID")
	if customerID == 0 {
		return utils.WriteErrorResponse(c, http.StatusUnauthorized, "unauthorized")
	}

	itemID, err := strconv.Atoi(c.Params("itemId"))
	if err != nil {
		return utils.WriteErrorResponse(c, http.StatusBadRequest, "invalid item ID")
	}

	var req struct {
		Quantity int `json:"quantity" validate:"required,min=1"`
	}

	if err := c.BodyParser(&req); err != nil {
		h.logger.Error("Failed to parse body: ", err)
		return utils.WriteErrorResponse(c, http.StatusBadRequest, "Invalid request body")
	}

	// Get cart for customer
	cart, err := h.cartUseCase.GetCartByCustomerID(customerID.(int), nil)
	if err != nil {
		h.logger.Errorf("Error getting cart: %v", err)
		return utils.WriteErrorResponse(c, http.StatusInternalServerError, err.Error())
	}
	cartData, ok := cart.Data.(entity.Cart)
	if !ok {
		return utils.WriteErrorResponse(c, http.StatusInternalServerError, "Invalid cart data")
	}

	if err := h.cartUseCase.UpdateItemQuantity(cartData.ID, itemID, req.Quantity); err != nil {
		h.logger.Errorf("Error updating item quantity: %v", err)
		return utils.WriteErrorResponse(c, http.StatusInternalServerError, err.Error())
	}

	return utils.WriteResponse(c, http.StatusOK, nil, "Item quantity updated successfully", nil)
}

func (h *CartController) RemoveItem(c *fiber.Ctx) error {
	customerID := c.Locals("customerID")
	if customerID == 0 {
		return utils.WriteErrorResponse(c, http.StatusUnauthorized, "unauthorized")
	}

	itemID, err := strconv.Atoi(c.Params("itemId"))
	if err != nil {
		return utils.WriteErrorResponse(c, http.StatusBadRequest, "invalid item ID")
	}

	// Get cart for customer
	cart, err := h.cartUseCase.GetCartByCustomerID(customerID.(int), nil)
	if err != nil {
		h.logger.Errorf("Error getting cart: %v", err)
		return utils.WriteErrorResponse(c, http.StatusInternalServerError, err.Error())
	}
	cartData, ok := cart.Data.(entity.Cart)
	if !ok {
		return utils.WriteErrorResponse(c, http.StatusInternalServerError, "Invalid cart data")
	}

	// Remove item from cart
	if err := h.cartUseCase.RemoveItem(cartData.ID, itemID); err != nil {
		h.logger.Errorf("Error removing item from cart: %v", err)
		return utils.WriteErrorResponse(c, http.StatusInternalServerError, err.Error())
	}

	return utils.WriteResponse(c, http.StatusOK, nil, "Item removed from cart successfully", nil)
}

func (h *CartController) ClearCart(c *fiber.Ctx) error {
	customerID := c.Locals("customerID")
	if customerID == 0 {
		return utils.WriteErrorResponse(c, http.StatusUnauthorized, "unauthorized")
	}

	// Get cart for customer
	cart, err := h.cartUseCase.GetCartByCustomerID(customerID.(int), nil)
	if err != nil {
		h.logger.Errorf("Error getting cart: %v", err)
		return utils.WriteErrorResponse(c, http.StatusInternalServerError, err.Error())
	}

	cartData, ok := cart.Data.(entity.Cart)
	if !ok {
		return utils.WriteErrorResponse(c, http.StatusInternalServerError, "Invalid cart data")
	}

	// Clear cart
	if err := h.cartUseCase.ClearCart(cartData.ID); err != nil {
		h.logger.Errorf("Error clearing cart: %v", err)
		return utils.WriteErrorResponse(c, http.StatusInternalServerError, err.Error())
	}

	return utils.WriteResponse(c, http.StatusOK, nil, "Cart cleared successfully", nil)
}
