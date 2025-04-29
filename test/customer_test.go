package test

import (
	"bytes"
	configs "cakestore/internal/config"
	"cakestore/internal/database"
	controller "cakestore/internal/delivery/http"
	"cakestore/internal/domain/entity"
	"cakestore/internal/domain/model"
	"cakestore/internal/middleware"
	"cakestore/internal/repository"
	"cakestore/internal/usecase"
	"cakestore/utils"
	"encoding/json"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type AuthTestSuite struct {
	suite.Suite
	app     *fiber.App
	db      *gorm.DB
	handler *controller.CustomerController
	useCase usecase.CustomerUseCase
	repo    repository.CustomerRepository
	logger  *logrus.Logger
	token   string
}

func (suite *AuthTestSuite) SetupTest() {
	cfg := configs.LoadConfig()
	db := database.ConnectPostgres(cfg)
	// Run migrations
	err := db.AutoMigrate(&entity.Customer{})
	assert.NoError(suite.T(), err)

	suite.db = db
	suite.logger = utils.NewLogger()
	suite.repo = repository.NewCustomerRepository(db, suite.logger)
	suite.useCase = usecase.NewCustomerUseCase(suite.repo, suite.logger, cfg.JWT_SECRET)
	suite.handler = controller.NewCustomerController(suite.useCase, suite.logger)

	suite.app = fiber.New()

	// Generate a test token
	token, err := utils.GenerateToken(123, "test@example.com", "Test User", "customer")
	suite.Require().NoError(err)
	suite.token = token

	// Setup routes
	suite.app.Post("/register", suite.handler.Register)
	suite.app.Post("/login", suite.handler.Login)
	suite.app.Get("/authorize", middleware.AuthMiddleware(cfg.JWT_SECRET), suite.handler.Authorize)
	suite.app.Get("/customers/me", middleware.AuthMiddleware(cfg.JWT_SECRET), suite.handler.GetCustomerByID)
	suite.app.Put("/customers/:id", middleware.AuthMiddleware(cfg.JWT_SECRET), suite.handler.UpdateProfile)
}

func (suite *AuthTestSuite) TestRegister() {
	payload := model.CreateCustomerRequest{
		Email:    "newuser@example.com",
		Password: "password123",
		Name:     "New User",
		Address:  "123 Main St",
	}

	jsonPayload, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(jsonPayload))
	req.Header.Set("Content-Type", "application/json")
	resp, err := suite.app.Test(req)
	suite.Require().NoError(err)
	suite.Equal(201, resp.StatusCode)
}

func (suite *AuthTestSuite) TestLogin() {
	payload := `{"email": "rafli@email.com", "password": "password123"}`
	req := httptest.NewRequest("POST", "/login", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	resp, err := suite.app.Test(req)
	suite.Require().NoError(err)
	suite.Equal(200, resp.StatusCode)
}

func (suite *AuthTestSuite) TestAuthorize() {
	req := httptest.NewRequest("GET", "/authorize", nil)
	req.Header.Set("Authorization", "Bearer "+suite.token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := suite.app.Test(req)
	suite.Require().NoError(err)
	suite.Equal(200, resp.StatusCode)
}

func (suite *AuthTestSuite) TestGetCustomerByID() {
	// FIXME: this test is failing
	payload := model.CreateCustomerRequest{
		Email:    "existinguser9@example.com",
		Password: "password123",
		Name:     "Existing User",
		Address:  "456 Another St",
	}

	jsonPayload, _ := json.Marshal(payload)

	createReq := httptest.NewRequest("POST", "/register", bytes.NewBuffer(jsonPayload))
	createReq.Header.Set("Content-Type", "application/json")
	createResp, err := suite.app.Test(createReq)
	suite.Require().NoError(err)
	suite.Equal(201, createResp.StatusCode)

	respBody, _ := io.ReadAll(createResp.Body)
	var response utils.Response
	_ = json.Unmarshal(respBody, &response)

	// parse token
	token, ok := response.Data.(string)
	suite.Require().True(ok)

	// Fetch the customer by assumed ID
	req := httptest.NewRequest("GET", "/customers/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := suite.app.Test(req)
	suite.Require().NoError(err)
	suite.Equal(200, resp.StatusCode)
}

func (suite *AuthTestSuite) TestUpdateProfile() {
	// FIXME: this test is failing
	payload := `{"name": "Updated User"}`
	req := httptest.NewRequest("PUT", "/customers/123", strings.NewReader(payload))
	req.Header.Set("Authorization", "Bearer "+suite.token)
	resp, err := suite.app.Test(req)
	suite.Require().NoError(err)
	suite.Equal(200, resp.StatusCode)
}

func TestAuthTestSuite(t *testing.T) {
	suite.Run(t, new(AuthTestSuite))
}
