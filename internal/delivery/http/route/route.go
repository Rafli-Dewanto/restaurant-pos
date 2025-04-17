package route

import (
	"cakestore/internal/constants"
	http "cakestore/internal/delivery/http"
	"cakestore/internal/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

type RouteConfig struct {
	App                *fiber.App
	CakeController     *http.CakeController
	CustomerController *http.CustomerController
	OrderController    *http.OrderController
	PaymentController  http.PaymentController
	JWTSecret          string
}

func (c *RouteConfig) Setup() {
	c.SetupRoute()
}

func (c *RouteConfig) SetupRoute() {
	c.App.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PATCH,PUT,DELETE",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))

	// Public routes
	c.App.Post("/register", c.CustomerController.Register)
	c.App.Post("/login", c.CustomerController.Login)
	// Midtrans notification webhook
	c.App.Post("/payment/notification", c.PaymentController.GetTransactionStatus)

	// Protected routes
	protectedRoutes := c.App.Group("/", middleware.AuthMiddleware(c.JWTSecret))

	// Customer routes
	protectedRoutes.Get("/customers/me", c.CustomerController.GetCustomerByID)
	protectedRoutes.Put("/customers/:id", c.CustomerController.UpdateProfile)

	// Cake routes
	cakes := protectedRoutes.Group("/cakes")
	cakes.Get("/", c.CakeController.GetAllCakes)
	cakes.Get("/:id", c.CakeController.GetCakeByID)
	cakes.Post("/", middleware.RoleMiddleware(constants.RoleAdmin), c.CakeController.CreateCake)
	cakes.Put("/:id", middleware.RoleMiddleware(constants.RoleAdmin), c.CakeController.UpdateCake)
	cakes.Delete("/:id", middleware.RoleMiddleware(constants.RoleAdmin), c.CakeController.DeleteCake)

	// Order routes
	orders := protectedRoutes.Group("/orders")
	orders.Post("/", c.OrderController.CreateOrder)
	orders.Get("/", c.OrderController.GetCustomerOrders)
	orders.Get("/:id", c.OrderController.GetOrderByID)
}
