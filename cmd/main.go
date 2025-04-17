package main

import (
	configs "cakestore/config"
	controller "cakestore/internal/delivery/http"
	"cakestore/internal/delivery/http/route"
	"cakestore/internal/domain/entity"
	"cakestore/internal/repository"
	"cakestore/internal/seeder"
	"cakestore/internal/usecase"
	"cakestore/utils"
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	logger := utils.NewLogger()
	cfg := configs.LoadConfig()

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Println(dsn)
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}

	log.Println("🔄 Running database migrations...")
	err = db.AutoMigrate(
		&entity.Cake{},
		&entity.Customer{},
		&entity.Order{},
		&entity.OrderItem{},
		&entity.Payment{},
	)
	if err != nil {
		log.Fatalf("❌ Failed to run database migrations: %v", err)
	}
	log.Println("✅ Database migrations completed successfully")

	// Initialize repositories
	cakeRepository := repository.NewCakeRepository(db, logger)
	customerRepository := repository.NewCustomerRepository(db, logger)
	orderRepository := repository.NewOrderRepository(db, logger)
	paymentRepository := repository.NewPaymentRepository(db, logger)

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
	orderUseCase := usecase.NewOrderUseCase(orderRepository, cakeRepository, customerRepository, logger, cfg.SERVER_ENV)
	paymentUsecase := usecase.NewPaymentUseCase(cfg.MIDTRANS_ENDPOINT, paymentRepository, logger, cfg.SERVER_ENV)

	// controller
	cakeController := controller.NewCakeController(cakeUseCase, logger)
	customerController := controller.NewCustomerController(customerUseCase, logger)
	orderController := controller.NewOrderController(orderUseCase, paymentUsecase, logger)
	paymentController := controller.NewPaymentController(logger, cfg.MIDTRANS_SERVER_KEY, orderUseCase, paymentUsecase)

	routeConfig := route.RouteConfig{
		App:                app,
		CakeController:     cakeController,
		CustomerController: customerController,
		OrderController:    orderController,
		PaymentController:  paymentController,
		JWTSecret:          cfg.JWT_SECRET,
	}
	routeConfig.Setup()

	port := cfg.SERVER_PORT
	if port == "" {
		port = "8080"
	}
	log.Printf("🚀 Server running on port %s", port)
	log.Fatal(app.Listen(":" + port))
}
