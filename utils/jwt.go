package utils

import (
	"fmt"
	"time"

	configs "cakestore/internal/config"

	"github.com/golang-jwt/jwt/v5"
)

var cfg = configs.LoadConfig()

type Claims struct {
	Email      string `json:"email"`
	Name       string `json:"name"`
	CustomerID int    `json:"customer_id"`
	Role       string `json:"role"`
	jwt.RegisteredClaims
}

// GenerateToken creates a new JWT token with email and name claims
func GenerateToken(customerID int, email, name, role string) (string, error) {
	claims := &Claims{
		Email:      email,
		Name:       name,
		CustomerID: customerID,
		Role:       role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(cfg.JWT_SECRET))
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	return tokenString, nil
}

// VerifyToken validates the JWT token and returns the claims
func VerifyToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(cfg.JWT_SECRET), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

// ExtractTokenFromBearer extracts token from "Bearer <token>" format
func ExtractTokenFromBearer(bearerToken string) string {
	const prefix = "Bearer "
	if len(bearerToken) > len(prefix) && bearerToken[:len(prefix)] == prefix {
		return bearerToken[len(prefix):]
	}
	return ""
}
