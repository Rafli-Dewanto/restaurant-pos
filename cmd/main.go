package main

import (
	configs "cakestore/internal/config"
	"cakestore/internal/database"
	controller "cakestore/internal/delivery/http"
	"cakestore/internal/delivery/http/route"
	"cakestore/internal/health"
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
	db := database.ConnectPostgres(cfg)
	err := database.RunMigrations(db)
	if err != nil {
		log.Fatalf("❌ Failed to run database migrations: %v", err)
	}

	// Initialize repositories
	menuRepository := repository.NewMenuRepository(db, logger)
	customerRepository := repository.NewCustomerRepository(db, logger)
	cartRepository := repository.NewCartRepository(db, logger)
	orderRepository := repository.NewOrderRepository(db, logger)
	paymentRepository := repository.NewPaymentRepository(db, logger)
	wishlistRepository := repository.NewWishListRepository(db, logger)
	reservationRepository := repository.NewReservationRepository(db, logger)
	inventoryRepository := repository.NewInventoryRepository(db, logger)
	tableRepository := repository.NewTableRepository(db, logger)

	// Initialize and run seeder
	dbSeeder := seeder.NewSeeder(customerRepository, menuRepository, logger, inventoryRepository, tableRepository)
	if err := dbSeeder.SeedAll(); err != nil {
		log.Printf("⚠️ Warning: Failed to seed database: %v", err)
	}

	app := fiber.New()

	healthChecker := health.NewHealthChecker(db)
	app.Get("/health", func(c *fiber.Ctx) error {
		health := healthChecker.Check()
		if health.Status != "healthy" {
			return c.Status(fiber.StatusServiceUnavailable).JSON(health)
		}
		return c.JSON(health)
	})

	// usecase
	menuUseCase := usecase.NewMenuUseCase(menuRepository, logger)
	customerUseCase := usecase.NewCustomerUseCase(customerRepository, logger, cfg.JWT_SECRET)
	cartUseCase := usecase.NewCartUseCase(cartRepository, menuRepository, logger)
	orderUseCase := usecase.NewOrderUseCase(orderRepository, menuRepository, customerRepository, logger, cfg.SERVER_ENV)
	paymentUsecase := usecase.NewPaymentUseCase(cfg.MIDTRANS_ENDPOINT, paymentRepository, logger, cfg.SERVER_ENV)
	wishlistUseCase := usecase.NewWishListUseCase(wishlistRepository, menuRepository, logger)
	reservationUseCase := usecase.NewReservationUseCase(reservationRepository, logger, tableRepository)
	inventoryUseCase := usecase.NewInventoryUseCase(inventoryRepository, logger)
	tableUseCase := usecase.NewTableUseCase(tableRepository, logger)

	// controller
	menuController := controller.NewMenuController(menuUseCase, logger)
	customerController := controller.NewCustomerController(customerUseCase, logger)
	orderController := controller.NewOrderController(orderUseCase, paymentUsecase, logger)
	cartController := controller.NewCartController(cartUseCase, logger)
	paymentController := controller.NewPaymentController(logger, cfg.MIDTRANS_SERVER_KEY, orderUseCase, paymentUsecase)
	wishlistController := controller.NewWishListController(wishlistUseCase, logger)
	reservationController := controller.NewReservationController(reservationUseCase, logger)
	ingredientController := controller.NewInventoryController(inventoryUseCase, logger)
	tableController := controller.NewTableController(tableUseCase, logger)

	routeConfig := route.RouteConfig{
		App:                   app,
		MenuController:        menuController,
		CustomerController:    customerController,
		CartController:        cartController,
		OrderController:       orderController,
		PaymentController:     paymentController,
		WishlistController:    wishlistController,
		ReservationController: reservationController,
		InventoryController:   ingredientController,
		TableController:       tableController,
		JWTSecret:             cfg.JWT_SECRET,
		Log:                   logger,
	}
	routeConfig.Setup()

	port := cfg.SERVER_PORT
	if port == "" {
		port = "8080"
	}
	log.Printf("🚀 Server running on port %s", port)
	log.Fatal(app.Listen("0.0.0.0:" + port))
}
