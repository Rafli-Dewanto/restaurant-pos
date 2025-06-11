package route

import (
	"cakestore/internal/constants"
	http "cakestore/internal/delivery/http"
	"cakestore/internal/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/sirupsen/logrus"
)

type RouteConfig struct {
	App                *fiber.App
	CakeController     *http.CakeController
	CustomerController *http.CustomerController
	CartController     *http.CartController
	OrderController    *http.OrderController
	WishlistController *http.WishListController
	PaymentController  http.PaymentController
	JWTSecret          string
	Log                *logrus.Logger
}

func (c *RouteConfig) Setup() {
	c.SetupRoute()
}

func (c *RouteConfig) SetupRoute() {
	c.App.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PATCH,PUT,DELETE",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization, X-App-Role",
	}))
	c.App.Use(middleware.LogMiddleware(c.Log))

	// Public routes
	c.App.Post("/register", c.CustomerController.Register)
	c.App.Post("/login", c.CustomerController.Login)
	// Midtrans notification webhook
	c.App.Post("/payment/notification", c.PaymentController.GetTransactionStatus)
	// cakes
	c.App.Get("/cakes", c.CakeController.GetAllCakes)
	c.App.Get("/cakes/:id", c.CakeController.GetCakeByID)

	// Protected routes
	protectedRoutes := c.App.Group("/", middleware.AuthMiddleware(c.JWTSecret))

	// Customer routes
	protectedRoutes.Get("/authorize", c.CustomerController.Authorize)
	protectedRoutes.Get("/customers/me", c.CustomerController.GetCustomerByID)
	protectedRoutes.Put("/customers/:id", c.CustomerController.UpdateProfile)

	// employee routes
	employeeRoutes := protectedRoutes.Group("/employees")
	employeeRoutes.Get("/", c.CustomerController.GetEmployees)
	employeeRoutes.Get("/:id", c.CustomerController.GetEmployeeByID)
	employeeRoutes.Put("/:id", middleware.RoleMiddleware(constants.RoleAdmin), c.CustomerController.UpdateEmployee)
	employeeRoutes.Delete("/:id", middleware.RoleMiddleware(constants.RoleAdmin), c.CustomerController.DeleteEmployee)

	// Cake routes
	cakes := protectedRoutes.Group("/cakes")
	cakes.Post("/", middleware.RoleMiddleware(constants.RoleAdmin, constants.RoleCashier, constants.RoleKitchen, constants.RoleWaitress), c.CakeController.CreateCake)
	cakes.Put("/:id", middleware.RoleMiddleware(constants.RoleAdmin, constants.RoleCashier, constants.RoleKitchen, constants.RoleWaitress), c.CakeController.UpdateCake)
	cakes.Delete("/:id", middleware.RoleMiddleware(constants.RoleAdmin, constants.RoleCashier, constants.RoleKitchen, constants.RoleWaitress), c.CakeController.DeleteCake)

	// Cart routes
	carts := protectedRoutes.Group("/carts")
	carts.Post("/", c.CartController.AddCart)
	carts.Get("/customer", c.CartController.GetCartByCustomerID)
	carts.Get("/:id", c.CartController.GetCartByID)
	carts.Delete("/:id", c.CartController.RemoveCart)
	carts.Delete("/", c.CartController.ClearCart)

	// Order routes
	orders := protectedRoutes.Group("/orders")
	orders.Get("/customers", c.OrderController.GetAllOrders)
	orders.Post("/", c.OrderController.CreateOrder)
	orders.Get("/", c.OrderController.GetCustomerOrders)
	orders.Get("/:id", c.OrderController.GetOrderByID)

	wishlist := protectedRoutes.Group("/wishlists")
	wishlist.Get("/", c.WishlistController.GetWishListByCustomerID)
	wishlist.Post("/:cakeId", c.WishlistController.CreateWishList)
	wishlist.Delete("/:cakeId", c.WishlistController.DeleteWishList)
}
