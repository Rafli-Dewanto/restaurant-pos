package constants

import "errors"

var (
	ErrInternalServerError        = errors.New("internal server error")
	ErrInvalidRequest             = errors.New("invalid request")
	ErrInvalidRequestBody         = errors.New("invalid request body")
	ErrInvalidRequestParam        = errors.New("invalid request param")
	ErrUnauthorized               = errors.New("unauthorized")
	ErrInvalidEmail               = errors.New("invalid email")
	ErrInvalidPassword            = errors.New("invalid password")
	ErrInvalidName                = errors.New("invalid name")
	ErrInvalidCustomerID          = errors.New("invalid customer ID")
	ErrInvalidQuantity            = errors.New("invalid quantity")
	ErrInvalidItemID              = errors.New("invalid item ID")
	ErrInvalidOrderID             = errors.New("invalid order ID")
	ErrInvalidPaymentID           = errors.New("invalid payment ID")
	ErrInvalidPaymentStatus       = errors.New("invalid payment status")
	ErrInvalidPaymentResponse     = errors.New("invalid payment response")
	ErrInvalidCartID              = errors.New("invalid cart ID")
	ErrInvalidCartItemID          = errors.New("invalid cart item ID")
	ErrNotFound                   = errors.New("not found")
	ErrInvalidInterfaceConversion = errors.New("invalid data type for interface conversion")
	ErrMenuAlreadyInWishlist      = errors.New("menu already in wishlist")
)
