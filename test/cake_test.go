package test

import (
	"bytes"
	configs "cakestore/config"
	controller "cakestore/internal/delivery/http"
	"cakestore/internal/domain/entity"
	"cakestore/internal/domain/model"
	"cakestore/internal/repository"
	"cakestore/internal/usecase"
	"cakestore/utils"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type CakeHandlerTestSuite struct {
	suite.Suite
	app     *fiber.App
	db      *gorm.DB
	handler *controller.CakeController
	useCase usecase.CakeUseCase
	repo    repository.CakeRepository
	logger  *logrus.Logger
	token   string
}

func (suite *CakeHandlerTestSuite) SetupTest() {
	cfg := configs.LoadConfig()

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Println(dsn)
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}

	// Run migrations
	err = db.AutoMigrate(&entity.Cake{})
	assert.NoError(suite.T(), err)

	suite.db = db
	suite.logger = utils.NewLogger()
	suite.repo = repository.NewCakeRepository(db, suite.logger)
	suite.useCase = usecase.NewCakeUseCase(suite.repo, suite.logger)
	suite.handler = controller.NewCakeController(suite.useCase, suite.logger)

	// Initialize Fiber app
	suite.app = fiber.New()

	// Generate a test token
	token, err := utils.GenerateToken(123, "test@example.com", "Test User")
	suite.Require().NoError(err)
	suite.token = token

	// Setup routes
	suite.app.Post("/cakes", suite.handler.CreateCake)
	suite.app.Get("/cakes/:id", suite.handler.GetCakeByID)
	suite.app.Get("/cakes", suite.handler.GetAllCakes)
	suite.app.Put("/cakes/:id", suite.handler.UpdateCake)
	suite.app.Delete("/cakes/:id", suite.handler.DeleteCake)
}

func TestSanity(t *testing.T) {
	t.Log("Sanity test runs")
}

func TestCakeHandlerSuite(t *testing.T) {
	suite.Run(t, new(CakeHandlerTestSuite))
}

func (suite *CakeHandlerTestSuite) TestCreate() {
	request := model.CreateUpdateCakeRequest{
		Title:       "Test Cake",
		Description: "Test Description",
		Rating:      4.5,
		ImageURL:    "http://example.com/test.jpg",
	}

	jsonValue, _ := json.Marshal(request)
	req := httptest.NewRequest("POST", "/cakes", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.token)

	resp, err := suite.app.Test(req)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusCreated, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var response utils.Response
	_ = json.Unmarshal(body, &response)

	jsonCake, _ := json.Marshal(response.Data)
	var cake entity.Cake
	_ = json.Unmarshal(jsonCake, &cake)

	assert.NotZero(suite.T(), cake.ID)
	assert.Equal(suite.T(), request.Title, cake.Title)

	assert.NoError(suite.T(), err)
	assert.NotZero(suite.T(), cake.ID)
	assert.Equal(suite.T(), request.Title, cake.Title)
}

func (suite *CakeHandlerTestSuite) TestGetByID() {
	cake := &entity.Cake{
		Title:       "Test Cake",
		Description: "Test Description",
		Rating:      4.5,
		Image:       "test.jpg",
	}

	jsonValue, _ := json.Marshal(cake)
	req := httptest.NewRequest("POST", "/cakes", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.token)

	resp, _ := suite.app.Test(req)

	body, _ := io.ReadAll(resp.Body)
	var respCreate utils.Response
	json.Unmarshal(body, &respCreate)

	// Test getting the cake
	req = httptest.NewRequest("GET", "/cakes/"+strconv.Itoa(int(respCreate.Data.(*entity.Cake).ID)), nil)
	resp, err := suite.app.Test(req)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusOK, resp.StatusCode)

	body, _ = io.ReadAll(resp.Body)
	var response utils.Response
	err = json.Unmarshal(body, &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), cake.Title, response.Data.(*entity.Cake).Title)

	// Test getting non-existent cake
	req = httptest.NewRequest("GET", "/cakes/9999", nil)
	resp, _ = suite.app.Test(req)
	assert.Equal(suite.T(), fiber.StatusNotFound, resp.StatusCode)
}

func (suite *CakeHandlerTestSuite) TestGetAll() {
	cakes := []entity.Cake{
		{Title: "Cake A", Rating: 4.5},
		{Title: "Cake B", Rating: 5.0},
		{Title: "Cake C", Rating: 4.0},
	}

	// Create test cakes
	for _, cake := range cakes {
		jsonValue, _ := json.Marshal(cake)
		req := httptest.NewRequest("POST", "/cakes", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")
		suite.app.Test(req)
	}

	// Test getting all cakes
	req := httptest.NewRequest("GET", "/cakes", nil)
	resp, err := suite.app.Test(req)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var response []entity.Cake
	err = json.Unmarshal(body, &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), len(cakes), len(response))
	assert.Greater(suite.T(), response[0].Rating, response[1].Rating)
	assert.Greater(suite.T(), response[1].Rating, response[2].Rating)
}

func (suite *CakeHandlerTestSuite) TestUpdate() {
	// Create a test cake first
	cake := &entity.Cake{
		Title:       "Original Cake",
		Description: "Original Description",
		Rating:      4.0,
		Image:       "original.jpg",
	}
	jsonValue, _ := json.Marshal(cake)
	req := httptest.NewRequest("POST", "/cakes", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := suite.app.Test(req)

	body, _ := io.ReadAll(resp.Body)
	var createdCake entity.Cake
	json.Unmarshal(body, &createdCake)

	// Update the cake
	updatedCake := createdCake
	updatedCake.Title = "Updated Cake"
	updatedCake.Rating = 4.5

	jsonValue, _ = json.Marshal(updatedCake)
	req = httptest.NewRequest("PUT", "/cakes/"+strconv.Itoa(int(createdCake.ID)), bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	resp, err := suite.app.Test(req)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusOK, resp.StatusCode)

	body, _ = io.ReadAll(resp.Body)
	var response entity.Cake
	err = json.Unmarshal(body, &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Updated Cake", response.Title)
	assert.Equal(suite.T(), 4.5, response.Rating)

	// Test updating non-existent cake
	req = httptest.NewRequest("PUT", "/cakes/9999", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	resp, _ = suite.app.Test(req)
	assert.Equal(suite.T(), fiber.StatusNotFound, resp.StatusCode)
}

func (suite *CakeHandlerTestSuite) TestDelete() {
	// Create a test cake first
	cake := &entity.Cake{
		Title:       "Test Cake",
		Description: "Test Description",
	}
	jsonValue, _ := json.Marshal(cake)
	req := httptest.NewRequest("POST", "/cakes", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := suite.app.Test(req)

	body, _ := io.ReadAll(resp.Body)
	var createdCake entity.Cake
	json.Unmarshal(body, &createdCake)

	// Delete the cake
	req = httptest.NewRequest("DELETE", "/cakes/"+strconv.Itoa(int(createdCake.ID)), nil)
	resp, _ = suite.app.Test(req)
	assert.Equal(suite.T(), fiber.StatusOK, resp.StatusCode)

	// Verify cake is deleted
	req = httptest.NewRequest("GET", "/cakes/"+strconv.Itoa(int(createdCake.ID)), nil)
	resp, _ = suite.app.Test(req)
	assert.Equal(suite.T(), fiber.StatusNotFound, resp.StatusCode)
}
