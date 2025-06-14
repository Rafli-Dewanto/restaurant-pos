package controller

import (
	"cakestore/internal/constants"
	"cakestore/internal/usecase"
	"cakestore/utils"
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

type WishListController struct {
	wishListUseCase usecase.WishListUseCase
	logger          *logrus.Logger
}

func NewWishListController(
	wishListUseCase usecase.WishListUseCase,
	logger *logrus.Logger,
) *WishListController {
	return &WishListController{
		wishListUseCase: wishListUseCase,
		logger:          logger,
	}
}

func (h *WishListController) CreateWishList(ctx *fiber.Ctx) error {
	h.logger.Trace("Creating wishlist")
	customerID := ctx.Locals(constants.ClaimsKeyID).(int64)
	menuID, err := strconv.ParseInt(ctx.Params("menuId"), 10, 64)
	if err != nil {
		h.logger.Errorf("Error converting menu ID: %v", err)
		return utils.WriteErrorResponse(ctx, http.StatusBadRequest, "Invalid menu ID")
	}

	err = h.wishListUseCase.CreateWishList(customerID, menuID)
	if err != nil {
		if err == constants.ErrMenuAlreadyInWishlist {
			h.logger.Warnf("Wishlist already exists: %v", err)
			return utils.WriteErrorResponse(ctx, http.StatusBadRequest, "Wishlist already exists")
		}
		h.logger.Errorf("Error creating wishlist: %v", err)
		return utils.WriteErrorResponse(ctx, http.StatusInternalServerError, "Failed to create wishlist")
	}

	return utils.WriteResponse(ctx, http.StatusCreated, nil, "Wishlist created successfully", nil)
}

func (h *WishListController) GetWishListByCustomerID(ctx *fiber.Ctx) error {
	h.logger.Trace("Getting wishlist by customer ID")
	customerID := ctx.Locals(constants.ClaimsKeyID).(int64)

	paginationQuery := utils.GetPaginationFromRequest(ctx)

	wishlists, meta, err := h.wishListUseCase.GetWishList(customerID, paginationQuery)
	if err != nil {
		h.logger.Errorf("Error getting wishlists: %v", err)
		return utils.WriteErrorResponse(ctx, http.StatusInternalServerError, "Failed to get wishlists")
	}

	return utils.WriteResponse(ctx, http.StatusOK, wishlists, "Wishlists retrieved successfully", meta)
}

func (c *WishListController) DeleteWishList(ctx *fiber.Ctx) error {
	c.logger.Trace("Deleting wishlist")
	customerID := ctx.Locals(constants.ClaimsKeyID).(int64)
	menuID, err := strconv.ParseInt(ctx.Params("menuId"), 10, 64)
	if err != nil {
		c.logger.Errorf("Error converting menu ID: %v", err)
		return utils.WriteErrorResponse(ctx, http.StatusBadRequest, "Invalid menu ID")
	}

	err = c.wishListUseCase.DeleteWishList(customerID, menuID)
	if err != nil {
		c.logger.Errorf("Error deleting wishlist: %v", err)
		return utils.WriteErrorResponse(ctx, http.StatusInternalServerError, "Failed to delete wishlist")
	}

	return utils.WriteResponse(ctx, http.StatusOK, nil, "Wishlist deleted successfully", nil)
}
