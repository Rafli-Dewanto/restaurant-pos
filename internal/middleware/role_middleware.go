package middleware

import (
	"github.com/gofiber/fiber/v2"
)

func RoleMiddleware(requiredRoles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userRole := c.Locals("role")

		for _, role := range requiredRoles {
			if userRole == role {
				return c.Next()
			}
		}

		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "You don't have permission to access this resource",
		})
	}
}
