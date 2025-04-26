package test

import (
	"bytes"
	configs "cakestore/internal/config"
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
	token, err := utils.GenerateToken(123, "test@example.com", "Test User", "customer")
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
	suite.T().Run("Test successful cake creation", func(t *testing.T) {
		request := model.CreateUpdateCakeRequest{
			Title:       "Test Cake",
			Description: "Test Description",
			Price:       90000,
			Category:    "birthday_cake",
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
	})

	suite.T().Run("Test invalid request body", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/cakes", bytes.NewBuffer([]byte(`{invalid json}`)))
		req.Header.Set("Content-Type", "application/json")
		resp, err := suite.app.Test(req)
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), fiber.StatusBadRequest, resp.StatusCode)
	})
}

func (suite *CakeHandlerTestSuite) TestGetByID() {
	suite.T().Run("Test successful cake retrieval", func(t *testing.T) {
		cake := &entity.Cake{
			Title:       "Test Cake",
			Description: "Test Description",
			Rating:      4.5,
			Price:       90000,
			Category:    "birthday_cake",
			Image:       "http://example.com/test.jpg",
		}

		// POST: create cake
		jsonValue, _ := json.Marshal(cake)
		req := httptest.NewRequest("POST", "/cakes", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+suite.token)

		resp, _ := suite.app.Test(req)
		body, _ := io.ReadAll(resp.Body)

		var respCreate utils.Response
		_ = json.Unmarshal(body, &respCreate)

		// convert respCreate.Data into model.CakeModel
		dataBytes, _ := json.Marshal(respCreate.Data)
		var createdCake model.CakeModel
		_ = json.Unmarshal(dataBytes, &createdCake)

		// GET: retrieve the cake
		req = httptest.NewRequest("GET", "/cakes/"+strconv.Itoa(int(createdCake.ID)), nil)
		resp, err := suite.app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		body, _ = io.ReadAll(resp.Body)
		var getResp utils.Response
		_ = json.Unmarshal(body, &getResp)

		dataBytes, _ = json.Marshal(getResp.Data)
		var fetchedCake model.CakeModel
		_ = json.Unmarshal(dataBytes, &fetchedCake)

		assert.Equal(t, createdCake.Title, fetchedCake.Title)
		assert.Equal(t, createdCake.Description, fetchedCake.Description)
		assert.Equal(t, createdCake.Rating, fetchedCake.Rating)
		assert.Equal(t, createdCake.Category, fetchedCake.Category)
		assert.Equal(t, createdCake.ImageURL, fetchedCake.ImageURL)
	})

	suite.T().Run("Test cake not found", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/cakes/9999", nil)
		resp, err := suite.app.Test(req)
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), fiber.StatusNotFound, resp.StatusCode)
	})
}

func (suite *CakeHandlerTestSuite) TestGetAll() {
	// Test getting all cakes
	perPage := 10
	req := httptest.NewRequest("GET", "/cakes", nil)
	resp, err := suite.app.Test(req)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var response model.PaginationResponse[[]entity.Cake]
	err = json.Unmarshal(body, &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), perPage, len(response.Data))
}

func (suite *CakeHandlerTestSuite) TestUpdate() {
	// Create a test cake first
	suite.T().Run("Test successful cake update", func(t *testing.T) {
		cake := &model.CreateUpdateCakeRequest{
			Title:       "Test Cake",
			Description: "Test Description",
			Rating:      4.5,
			Price:       90000,
			Category:    "birthday_cake",
			ImageURL:    "http://example.com/test.jpg",
		}

		// POST: create cake
		jsonValue, _ := json.Marshal(cake)
		req := httptest.NewRequest("POST", "/cakes", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+suite.token)

		resp, _ := suite.app.Test(req)
		body, _ := io.ReadAll(resp.Body)

		var respCreate utils.Response
		_ = json.Unmarshal(body, &respCreate)

		// convert respCreate.Data into model.CakeModel
		dataBytes, _ := json.Marshal(respCreate.Data)
		var createdCake model.CakeModel
		_ = json.Unmarshal(dataBytes, &createdCake)

		// Update the cake
		updatedCake := createdCake
		updatedCake.Title = "Updated Cake"
		updatedCake.Rating = 4.5
		updatedCake.ImageURL = "http://example.com/updated.jpg"
		updatedCake.Price = 100
		updatedCake.Category = "test"
		updatedCake.Description = "Updated Description"

		testUpdatedCake := model.CreateUpdateCakeRequest{
			Title:       updatedCake.Title,
			Description: updatedCake.Description,
			Rating:      updatedCake.Rating,
			ImageURL:    updatedCake.ImageURL,
			Price:       updatedCake.Price,
			Category:    updatedCake.Category,
		}

		jsonValue, _ = json.Marshal(testUpdatedCake)
		req = httptest.NewRequest("PUT", "/cakes/"+strconv.Itoa(int(createdCake.ID)), bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+suite.token)

		resp, err := suite.app.Test(req)
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), fiber.StatusOK, resp.StatusCode)

		body, _ = io.ReadAll(resp.Body)
		var response utils.Response
		_ = json.Unmarshal(body, &response)

		dataBytes, _ = json.Marshal(response.Data)
		var updated model.CakeModel
		_ = json.Unmarshal(dataBytes, &updated)

		assert.Equal(suite.T(), "Updated Cake", updated.Title)

		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), "Updated Cake", updated.Title)
		assert.Equal(suite.T(), 4.5, updated.Rating)
	})

	// Test updating non-existent cake - should return 404
	suite.T().Run("Update non-existent cake", func(t *testing.T) {
		cake := &model.CreateUpdateCakeRequest{
			Title:       "Test Cake",
			Description: "Test Description",
			Price:       90000,
			Category:    "birthday_cake",
			Rating:      4.5,
			ImageURL:    "http://example.com/test.jpg",
		}
		jsonValue, _ := json.Marshal(cake)
		req := httptest.NewRequest("PUT", "/cakes/9999999", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+suite.token)
		resp, err := suite.app.Test(req)
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), fiber.StatusNotFound, resp.StatusCode)
	})
}

func (suite *CakeHandlerTestSuite) TestDelete() {
	// Create a test cake first
	cake := &model.CreateUpdateCakeRequest{
		Title:       "Test Cake",
		Description: "Test Description",
		Price:       90000,
		Category:    "birthday_cake",
		Rating:      4.5,
		ImageURL:    "http://example.com/test.jpg",
	}
	jsonValue, _ := json.Marshal(cake)
	req := httptest.NewRequest("POST", "/cakes", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.token)
	resp, _ := suite.app.Test(req)

	body, _ := io.ReadAll(resp.Body)
	var res utils.Response
	json.Unmarshal(body, &res)

	jsonBody, _ := json.Marshal(res.Data)
	var result model.CakeModel
	_ = json.Unmarshal(jsonBody, &result)

	// Delete the cake
	req = httptest.NewRequest("DELETE", "/cakes/"+strconv.Itoa(int(result.ID)), nil)
	resp, _ = suite.app.Test(req)
	req.Header.Set("Authorization", "Bearer "+suite.token)
	assert.Equal(suite.T(), fiber.StatusOK, resp.StatusCode)

	// Verify cake is deleted
	req = httptest.NewRequest("GET", "/cakes/"+strconv.Itoa(int(result.ID)), nil)
	resp, _ = suite.app.Test(req)
	assert.Equal(suite.T(), fiber.StatusNotFound, resp.StatusCode)
}
