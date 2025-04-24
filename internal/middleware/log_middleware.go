package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

func LogMiddleware(logger *logrus.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ip := c.IP()
		method := c.Method()
		path := c.Path()

		// Process request
		err := c.Next()

		// Log after response is generated
		status := c.Response().StatusCode()

		logger.WithFields(logrus.Fields{
			"ip":     ip,
			"method": method,
			"path":   path,
			"status": status,
			"error":  err,
		}).Info("HTTP Request")

		return err
	}
}
