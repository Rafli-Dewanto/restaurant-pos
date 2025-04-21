package utils

import (
	configs "cakestore/internal/config"
)

func GenerateRequestHeader() map[string]string {
	cfg := configs.LoadConfig()
	SERVER_KEY := cfg.MIDTRANS_SERVER_KEY
	base64ServerKey := EncodeToBase64(SERVER_KEY)

	return map[string]string{
		"Authorization": "Basic " + base64ServerKey,
		"Content-Type":  "application/json",
		"Accept":        "application/json",
	}
}
