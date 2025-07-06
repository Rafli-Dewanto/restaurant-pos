// internal/bootstrap/bootstrap.go
package bootstrap

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
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Application struct {
	App    *fiber.App
	Config *configs.Config
	DB     *gorm.DB
	Logger *logrus.Logger
}

type Dependencies struct {
	// Repositories
	MenuRepository        repository.MenuRepository
	CustomerRepository    repository.CustomerRepository
	CartRepository        repository.CartRepository
	OrderRepository       repository.OrderRepository
	PaymentRepository     repository.PaymentRepository
	WishlistRepository    repository.WishListRepository
	ReservationRepository repository.ReservationRepository
	InventoryRepository   repository.InventoryRepository
	TableRepository       repository.TableRepository

	// Use Cases
	MenuUseCase        usecase.MenuUseCase
	CustomerUseCase    usecase.CustomerUseCase
	CartUseCase        usecase.CartUseCase
	OrderUseCase       usecase.OrderUseCase
	PaymentUseCase     usecase.PaymentUseCase
	WishlistUseCase    usecase.WishListUseCase
	ReservationUseCase usecase.ReservationUseCase
	InventoryUseCase   usecase.InventoryUseCase
	TableUseCase       usecase.TableUseCase

	// Controllers
	MenuController        *controller.MenuController
	CustomerController    *controller.CustomerController
	OrderController       *controller.OrderController
	CartController        *controller.CartController
	PaymentController     controller.PaymentController
	WishlistController    *controller.WishListController
	ReservationController *controller.ReservationController
	InventoryController   *controller.InventoryController
	TableController       *controller.TableController
}

func NewApplication() *Application {
	logger := utils.NewLogger()
	cfg := configs.LoadConfig()
	db := database.ConnectPostgres(cfg)

	// Run migrations
	if err := database.RunMigrations(db); err != nil {
		log.Fatalf("❌ Failed to run database migrations: %v", err)
	}

	app := fiber.New()

	return &Application{
		App:    app,
		Config: cfg,
		DB:     db,
		Logger: logger,
	}
}

func (a *Application) initializeRepositories() Dependencies {
	deps := Dependencies{}

	// Initialize repositories
	deps.MenuRepository = repository.NewMenuRepository(a.DB, a.Logger)
	deps.CustomerRepository = repository.NewCustomerRepository(a.DB, a.Logger)
	deps.CartRepository = repository.NewCartRepository(a.DB, a.Logger)
	deps.OrderRepository = repository.NewOrderRepository(a.DB, a.Logger)
	deps.PaymentRepository = repository.NewPaymentRepository(a.DB, a.Logger)
	deps.WishlistRepository = repository.NewWishListRepository(a.DB, a.Logger)
	deps.ReservationRepository = repository.NewReservationRepository(a.DB, a.Logger)
	deps.InventoryRepository = repository.NewInventoryRepository(a.DB, a.Logger)
	deps.TableRepository = repository.NewTableRepository(a.DB, a.Logger)

	return deps
}

func (a *Application) initializeUseCases(deps *Dependencies) {
	// Initialize use cases
	deps.MenuUseCase = usecase.NewMenuUseCase(deps.MenuRepository, a.Logger)
	deps.CustomerUseCase = usecase.NewCustomerUseCase(deps.CustomerRepository, a.Logger, a.Config.JWT_SECRET)
	deps.CartUseCase = usecase.NewCartUseCase(deps.CartRepository, deps.MenuRepository, a.Logger)
	deps.OrderUseCase = usecase.NewOrderUseCase(deps.OrderRepository, deps.MenuRepository, deps.CustomerRepository, a.Logger, a.Config.SERVER_ENV)
	deps.PaymentUseCase = usecase.NewPaymentUseCase(a.Config.MIDTRANS_ENDPOINT, deps.PaymentRepository, a.Logger, a.Config.SERVER_ENV)
	deps.WishlistUseCase = usecase.NewWishListUseCase(deps.WishlistRepository, deps.MenuRepository, a.Logger)
	deps.ReservationUseCase = usecase.NewReservationUseCase(deps.ReservationRepository, a.Logger, deps.TableRepository)
	deps.InventoryUseCase = usecase.NewInventoryUseCase(deps.InventoryRepository, a.Logger)
	deps.TableUseCase = usecase.NewTableUseCase(deps.TableRepository, a.Logger)
}

func (a *Application) initializeControllers(deps *Dependencies) {
	// Initialize controllers
	deps.MenuController = controller.NewMenuController(deps.MenuUseCase, a.Logger)
	deps.CustomerController = controller.NewCustomerController(deps.CustomerUseCase, a.Logger)
	deps.OrderController = controller.NewOrderController(deps.OrderUseCase, deps.PaymentUseCase, a.Logger)
	deps.CartController = controller.NewCartController(deps.CartUseCase, a.Logger)
	deps.PaymentController = controller.NewPaymentController(a.Logger, a.Config.MIDTRANS_SERVER_KEY, deps.OrderUseCase, deps.PaymentUseCase)
	deps.WishlistController = controller.NewWishListController(deps.WishlistUseCase, a.Logger)
	deps.ReservationController = controller.NewReservationController(deps.ReservationUseCase, a.Logger)
	deps.InventoryController = controller.NewInventoryController(deps.InventoryUseCase, a.Logger)
	deps.TableController = controller.NewTableController(deps.TableUseCase, a.Logger)
}

func (a *Application) seedDatabase(deps *Dependencies) {
	// Initialize and run seeder
	dbSeeder := seeder.NewSeeder(deps.CustomerRepository, deps.MenuRepository, a.Logger, deps.InventoryRepository, deps.TableRepository)
	if err := dbSeeder.SeedAll(); err != nil {
		log.Printf("⚠️ Warning: Failed to seed database: %v", err)
	}
}

func (a *Application) setupHealthCheck() {
	healthChecker := health.NewHealthChecker(a.DB)
	a.App.Get("/health", func(c *fiber.Ctx) error {
		health := healthChecker.Check()
		if health.Status != "healthy" {
			return c.Status(fiber.StatusServiceUnavailable).JSON(health)
		}
		return c.JSON(health)
	})
}

func (a *Application) setupRoutes(deps *Dependencies) {
	routeConfig := route.RouteConfig{
		App:                   a.App,
		MenuController:        deps.MenuController,
		CustomerController:    deps.CustomerController,
		CartController:        deps.CartController,
		OrderController:       deps.OrderController,
		PaymentController:     deps.PaymentController,
		WishlistController:    deps.WishlistController,
		ReservationController: deps.ReservationController,
		InventoryController:   deps.InventoryController,
		TableController:       deps.TableController,
		JWTSecret:             a.Config.JWT_SECRET,
		Log:                   a.Logger,
	}
	routeConfig.Setup()
}

func (a *Application) Bootstrap() {
	// Initialize all dependencies in order
	deps := a.initializeRepositories()
	a.initializeUseCases(&deps)
	a.initializeControllers(&deps)

	// Seed database
	a.seedDatabase(&deps)

	// Setup health check
	a.setupHealthCheck()

	// Setup routes
	a.setupRoutes(&deps)
}

func (a *Application) Start() {
	port := a.Config.SERVER_PORT
	if port == "" {
		port = "8080"
	}
	log.Printf("🚀 Server running on port %s", port)
	log.Fatal(a.App.Listen("0.0.0.0:" + port))
}
