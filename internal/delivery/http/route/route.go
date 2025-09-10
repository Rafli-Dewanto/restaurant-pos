package route

import (
	"cakestore/internal/constants"
	http "cakestore/internal/delivery/http"
	"cakestore/internal/middleware"

	"github.com/gofiber/contrib/swagger"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	"github.com/sirupsen/logrus"
)

type RouteConfig struct {
	App                   *fiber.App
	MenuController        *http.MenuController
	CustomerController    *http.CustomerController
	CartController        *http.CartController
	OrderController       *http.OrderController
	WishlistController    *http.WishListController
	PaymentController     http.PaymentController
	ReservationController *http.ReservationController
	InventoryController   *http.InventoryController
	TableController       *http.TableController
	JWTSecret             string
	Log                   *logrus.Logger
}

func (c *RouteConfig) Setup() {
	c.SetupRoute()
}

func (c *RouteConfig) SetupRoute() {
	// CORS configuration
	c.App.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PATCH,PUT,DELETE",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization, X-App-Role, User-Agent",
	}))

	// Performance profiling
	c.App.Use(pprof.New())

	// Logging middleware
	c.App.Use(middleware.LogMiddleware(c.Log))

	// Global rate limiting - Apply to all routes
	c.App.Use(middleware.BasicRateLimit(c.Log))

	// Static files
	c.App.Static("/docs", "./docs")

	// Swagger configuration
	cfg := swagger.Config{
		FilePath: "./docs/swagger.json",
		Path:     "docs",
		Title:    "Swagger API Docs",
		BasePath: "/api/v1/",
	}
	c.App.Use(swagger.New(cfg))

	// Public routes with specific rate limiting

	// Authentication routes - strict rate limiting
	authGroup := c.App.Group("/", middleware.AuthRateLimit(c.Log))
	authGroup.Post("/register", c.CustomerController.Register)
	authGroup.Post("/login", c.CustomerController.Login)

	// Payment webhook - separate rate limiting
	paymentWebhook := c.App.Group("/payment", middleware.PaymentRateLimit(c.Log))
	paymentWebhook.Post("/notification/", c.PaymentController.GetTransactionStatus)

	// Menu routes - moderate rate limiting for public access
	menuPublic := c.App.Group("/menus", middleware.IPBasedRateLimit(50, 15*60, c.Log)) // 50 requests per 15 minutes
	menuPublic.Get("/", c.MenuController.GetAllMenus)
	menuPublic.Get("/:id", c.MenuController.GetMenuByID)

	// Protected routes
	protectedRoutes := c.App.Group("/api/v1",
		middleware.AuthMiddleware(c.JWTSecret),
		middleware.UserBasedRateLimit(200, 60*60, c.Log), // 200 requests per hour per user
	)

	// Customer routes
	protectedRoutes.Get("/authorize", c.CustomerController.Authorize)
	protectedRoutes.Get("/customers/me", c.CustomerController.GetCustomerByID)
	protectedRoutes.Put("/customers/:id", c.CustomerController.UpdateProfile)

	// Employee routes - Admin/Manager level rate limiting
	employeeRoutes := protectedRoutes.Group("/employees")
	employeeRoutes.Get("/", c.CustomerController.GetEmployees)
	employeeRoutes.Get("/:id", c.CustomerController.GetEmployeeByID)
	employeeRoutes.Put("/:id", middleware.RoleMiddleware(constants.RoleAdmin), c.CustomerController.UpdateEmployee)
	employeeRoutes.Delete("/:id", middleware.RoleMiddleware(constants.RoleAdmin), c.CustomerController.DeleteEmployee)

	// Menu routes - Staff operations
	menus := protectedRoutes.Group("/menus")
	menus.Post("/", middleware.RoleMiddleware(constants.RoleAdmin, constants.RoleCashier, constants.RoleKitchen, constants.RoleWaitress), c.MenuController.CreateMenu)
	menus.Put("/:id", middleware.RoleMiddleware(constants.RoleAdmin, constants.RoleCashier, constants.RoleKitchen, constants.RoleWaitress), c.MenuController.UpdateMenu)
	menus.Delete("/:id", middleware.RoleMiddleware(constants.RoleAdmin, constants.RoleCashier, constants.RoleKitchen, constants.RoleWaitress), c.MenuController.DeleteMenu)

	// Cart routes - Higher rate limiting for frequent operations
	carts := protectedRoutes.Group("/carts", middleware.UserBasedRateLimit(100, 15*60, c.Log)) // 100 requests per 15 minutes
	carts.Post("/", c.CartController.AddCart)
	carts.Get("/customer", c.CartController.GetCartByCustomerID)
	carts.Get("/:id", c.CartController.GetCartByID)
	carts.Delete("/:id", c.CartController.RemoveCart)
	carts.Delete("/", c.CartController.ClearCart)
	carts.Post("/bulk", c.CartController.BulkDeleteCart)

	// Order routes - Moderate rate limiting
	orders := protectedRoutes.Group("/orders", middleware.UserBasedRateLimit(50, 60*60, c.Log)) // 50 orders per hour
	orders.Get("/customers", c.OrderController.GetAllOrders)
	orders.Post("/", c.OrderController.CreateOrder)
	orders.Get("/", c.OrderController.GetCustomerOrders)
	orders.Get("/:id", c.OrderController.GetOrderByID)
	orders.Patch("/:id/food-status", middleware.RoleMiddleware(constants.RoleAdmin, constants.RoleKitchen), c.OrderController.UpdateFoodStatus)

	// Payment routes - Strict rate limiting for security
	payment := protectedRoutes.Group("/payments", middleware.UserBasedRateLimit(10, 60*60, c.Log)) // 10 payment requests per hour
	payment.Get("/:id", c.PaymentController.GetPaymentURL)

	// Wishlist routes - Moderate rate limiting
	wishlist := protectedRoutes.Group("/wishlists", middleware.UserBasedRateLimit(30, 15*60, c.Log)) // 30 requests per 15 minutes
	wishlist.Get("/", c.WishlistController.GetWishListByCustomerID)
	wishlist.Post("/:menuId", c.WishlistController.CreateWishList)
	wishlist.Delete("/:menuId", c.WishlistController.DeleteWishList)

	// Reservation routes - Moderate rate limiting
	reservation := protectedRoutes.Group("/reservations", middleware.UserBasedRateLimit(20, 60*60, c.Log)) // 20 reservations per hour
	reservation.Post("/", c.ReservationController.CreateReservation)
	reservation.Get("/", c.ReservationController.GetAllReservations)
	reservation.Get("/admin", middleware.RoleMiddleware(constants.RoleAdmin), c.ReservationController.AdminGetAllCustomerReservations)
	reservation.Get("/:id", c.ReservationController.GetReservationByID)
	reservation.Put("/:id", middleware.RoleMiddleware(constants.RoleAdmin, constants.RoleWaitress), c.ReservationController.UpdateReservation)
	reservation.Delete("/:id", middleware.RoleMiddleware(constants.RoleAdmin, constants.RoleWaitress), c.ReservationController.DeleteReservation)

	// Inventory routes - Staff only, moderate rate limiting
	inventory := protectedRoutes.Group("/inventories", middleware.UserBasedRateLimit(100, 60*60, c.Log)) // 100 requests per hour for staff
	inventory.Get("/", c.InventoryController.GetAllInventories)
	inventory.Get("/low-stock", middleware.RoleMiddleware(constants.RoleAdmin, constants.RoleKitchen), c.InventoryController.GetLowStockInventories)
	// temporary fix for conflicting route (/low-stock)
	inventory.Get("/by-id/:id", middleware.RoleMiddleware(constants.RoleAdmin, constants.RoleKitchen), c.InventoryController.GetInventoryByID)
	inventory.Post("/", middleware.RoleMiddleware(constants.RoleAdmin, constants.RoleKitchen), c.InventoryController.CreateInventory)
	inventory.Put("/:id", middleware.RoleMiddleware(constants.RoleAdmin, constants.RoleKitchen), c.InventoryController.UpdateInventory)
	inventory.Delete("/:id", middleware.RoleMiddleware(constants.RoleAdmin, constants.RoleKitchen), c.InventoryController.DeleteInventory)
	inventory.Put("/:id/stock", middleware.RoleMiddleware(constants.RoleAdmin, constants.RoleKitchen), c.InventoryController.UpdateInventoryStock)

	// Table routes - Staff operations, moderate rate limiting
	tables := protectedRoutes.Group("/tables", middleware.UserBasedRateLimit(80, 60*60, c.Log)) // 80 requests per hour
	tables.Get("/", c.TableController.GetAllTables)
	tables.Get("/:id", c.TableController.GetTableByID)
	tables.Post("/", middleware.RoleMiddleware(constants.RoleAdmin, constants.RoleCashier), c.TableController.CreateTable)
	tables.Put("/:id", middleware.RoleMiddleware(constants.RoleAdmin, constants.RoleCashier), c.TableController.UpdateTable)
	tables.Patch("/:id/availability", middleware.RoleMiddleware(constants.RoleAdmin, constants.RoleCashier), c.TableController.UpdateTableAvailability)
	tables.Delete("/:id", middleware.RoleMiddleware(constants.RoleAdmin, constants.RoleCashier), c.TableController.DeleteTable)
}
