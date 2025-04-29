package main

import (
	configs "cakestore/internal/config"
	"cakestore/internal/database"
	controller "cakestore/internal/delivery/http"
	"cakestore/internal/delivery/http/route"
	"cakestore/internal/repository"
	"cakestore/internal/seeder"
	"cakestore/internal/usecase"
	"cakestore/utils"
	"log"

	"github.com/gofiber/fiber/v2"
)

func main() {
	logger := utils.NewLogger()
	cfg := configs.LoadConfig()
	db := database.ConnectMySQL(cfg)
	err := database.RunMigrations(db)
	if err != nil {
		log.Fatalf("❌ Failed to run database migrations: %v", err)
	}

	// Initialize repositories
	cakeRepository := repository.NewCakeRepository(db, logger)
	customerRepository := repository.NewCustomerRepository(db, logger)
	cartRepository := repository.NewCartRepository(db, logger)
	orderRepository := repository.NewOrderRepository(db, logger)
	paymentRepository := repository.NewPaymentRepository(db, logger)
	wishlistRepository := repository.NewWishListRepository(db, logger)

	// Initialize and run seeder
	dbSeeder := seeder.NewSeeder(customerRepository, cakeRepository, logger)
	if err := dbSeeder.SeedAll(); err != nil {
		log.Printf("⚠️ Warning: Failed to seed database: %v", err)
	}

	app := fiber.New()

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	// usecase
	cakeUseCase := usecase.NewCakeUseCase(cakeRepository, logger)
	customerUseCase := usecase.NewCustomerUseCase(customerRepository, logger, cfg.JWT_SECRET)
	cartUseCase := usecase.NewCartUseCase(cartRepository, cakeRepository, logger)
	orderUseCase := usecase.NewOrderUseCase(orderRepository, cakeRepository, customerRepository, logger, cfg.SERVER_ENV)
	paymentUsecase := usecase.NewPaymentUseCase(cfg.MIDTRANS_ENDPOINT, paymentRepository, logger, cfg.SERVER_ENV)
	wishlistUseCase := usecase.NewWishListUseCase(wishlistRepository, cakeRepository, logger)

	// controller
	cakeController := controller.NewCakeController(cakeUseCase, logger)
	customerController := controller.NewCustomerController(customerUseCase, logger)
	orderController := controller.NewOrderController(orderUseCase, paymentUsecase, logger)
	cartController := controller.NewCartController(cartUseCase, logger)
	paymentController := controller.NewPaymentController(logger, cfg.MIDTRANS_SERVER_KEY, orderUseCase, paymentUsecase)
	wishlistController := controller.NewWishListController(wishlistUseCase, logger)

	routeConfig := route.RouteConfig{
		App:                app,
		CakeController:     cakeController,
		CustomerController: customerController,
		CartController:     cartController,
		OrderController:    orderController,
		PaymentController:  paymentController,
		WishlistController: wishlistController,
		JWTSecret:          cfg.JWT_SECRET,
		Log:                logger,
	}
	routeConfig.Setup()

	port := cfg.SERVER_PORT
	if port == "" {
		port = "8080"
	}
	log.Printf("🚀 Server running on port %s", port)
	log.Fatal(app.Listen(":" + port))
}
