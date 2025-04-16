package model

import "cakestore/internal/domain/entity"

type CustomerResponse struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Address string `json:"address"`
}

type CreateCustomerRequest struct {
	Name     string `json:"name" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
	Address  string `json:"address" validate:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

func CustomerToResponse(customer *entity.Customer) *CustomerResponse {
	return &CustomerResponse{
		ID:      customer.ID,
		Name:    customer.Name,
		Email:   customer.Email,
		Address: customer.Address,
	}
}

func ToLoginResponse(token string) *LoginResponse {
	return &LoginResponse{
		Token: token,
	}
}
