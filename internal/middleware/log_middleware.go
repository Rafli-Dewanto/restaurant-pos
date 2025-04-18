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
		logger.Infof("IP: %s, Method: %s, Path: %s", ip, method, path)
		c.Next()
		return nil
	}
}
