package test

import (
	"bytes"
	configs "cakestore/internal/config"
	"cakestore/internal/database"
	controller "cakestore/internal/delivery/http"
	"cakestore/internal/domain/entity"
	"cakestore/internal/domain/model"
	"cakestore/internal/repository"
	"cakestore/internal/usecase"
	"cakestore/utils"
	"context"
	"encoding/json"
	"io"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type MenuHandlerTestSuite struct {
	suite.Suite
	app     *fiber.App
	db      *gorm.DB
	handler *controller.MenuController
	useCase usecase.MenuUseCase
	repo    repository.MenuRepository
	logger  *logrus.Logger
	token   string
}

func (suite *MenuHandlerTestSuite) SetupTest() {
	cfg := configs.LoadConfig()

	db := database.ConnectPostgres(cfg)

	// Run migrations
	err := db.AutoMigrate(&entity.Menu{})
	assert.NoError(suite.T(), err)

	ctx := context.Background()
	redis := database.NewRedisCacheService(ctx, "")

	suite.db = db
	suite.logger = utils.NewLogger()
	suite.repo = repository.NewMenuRepository(db, suite.logger)
	suite.useCase = usecase.NewMenuUseCase(suite.repo, suite.logger, redis)
	suite.handler = controller.NewMenuController(suite.useCase, suite.logger)

	// Initialize Fiber app
	suite.app = fiber.New()

	// Generate a test token
	token, err := utils.GenerateToken(123, "test@example.com", "Test User", "customer")
	suite.Require().NoError(err)
	suite.token = token

	// Setup routes
	suite.app.Post("/menus", suite.handler.CreateMenu)
	suite.app.Get("/menus/:id", suite.handler.GetMenuByID)
	suite.app.Get("/menus", suite.handler.GetAllMenus)
	suite.app.Put("/menus/:id", suite.handler.UpdateMenu)
	suite.app.Delete("/menus/:id", suite.handler.DeleteMenu)
}

func TestSanity(t *testing.T) {
	t.Log("Sanity test runs")
}

func TestMenuHandlerSuite(t *testing.T) {
	suite.Run(t, new(MenuHandlerTestSuite))
}

func (suite *MenuHandlerTestSuite) TestCreate() {
	suite.T().Run("Test successful menu creation", func(t *testing.T) {
		request := model.CreateUpdateMenuRequest{
			Title:       "Test Menu",
			Description: "Test Description",
			Price:       90000,
			Category:    "birthday_cake",
			Rating:      4.5,
			ImageURL:    "http://example.com/test.jpg",
		}

		jsonValue, _ := json.Marshal(request)
		req := httptest.NewRequest("POST", "/menus", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+suite.token)

		resp, err := suite.app.Test(req)
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), fiber.StatusCreated, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		var response utils.Response
		_ = json.Unmarshal(body, &response)

		jsonMenu, _ := json.Marshal(response.Data)
		var menu entity.Menu
		_ = json.Unmarshal(jsonMenu, &menu)

		assert.NotZero(suite.T(), menu.ID)
		assert.Equal(suite.T(), request.Title, menu.Title)

		assert.NoError(suite.T(), err)
		assert.NotZero(suite.T(), menu.ID)
		assert.Equal(suite.T(), request.Title, menu.Title)
	})

	suite.T().Run("Test invalid request body", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/menus", bytes.NewBuffer([]byte(`{invalid json}`)))
		req.Header.Set("Content-Type", "application/json")
		resp, err := suite.app.Test(req)
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), fiber.StatusBadRequest, resp.StatusCode)
	})
}

func (suite *MenuHandlerTestSuite) TestGetByID() {
	suite.T().Run("Test successful menu retrieval", func(t *testing.T) {
		menu := &entity.Menu{
			Title:       "Test Cake",
			Description: "Test Description",
			Rating:      4.5,
			Price:       90000,
			Category:    "birthday_cake",
			Image:       "http://example.com/test.jpg",
		}

		// POST: create menu
		jsonValue, _ := json.Marshal(menu)
		req := httptest.NewRequest("POST", "/menus", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+suite.token)

		resp, _ := suite.app.Test(req)
		body, _ := io.ReadAll(resp.Body)

		var respCreate utils.Response
		_ = json.Unmarshal(body, &respCreate)

		dataBytes, _ := json.Marshal(respCreate.Data)
		var createdMenu model.MenuModel
		_ = json.Unmarshal(dataBytes, &createdMenu)

		// GET: retrieve the menu
		req = httptest.NewRequest("GET", "/menus/"+strconv.Itoa(int(createdMenu.ID)), nil)
		resp, err := suite.app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		body, _ = io.ReadAll(resp.Body)
		var getResp utils.Response
		_ = json.Unmarshal(body, &getResp)

		dataBytes, _ = json.Marshal(getResp.Data)
		var fetchedMenu model.MenuModel
		_ = json.Unmarshal(dataBytes, &fetchedMenu)

		assert.Equal(t, createdMenu.Title, fetchedMenu.Title)
		assert.Equal(t, createdMenu.Description, fetchedMenu.Description)
		assert.Equal(t, createdMenu.Rating, fetchedMenu.Rating)
		assert.Equal(t, createdMenu.Category, fetchedMenu.Category)
		assert.Equal(t, createdMenu.ImageURL, fetchedMenu.ImageURL)
	})

	suite.T().Run("Test menu not found", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/menus/9999", nil)
		resp, err := suite.app.Test(req)
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), fiber.StatusNotFound, resp.StatusCode)
	})
}

func (suite *MenuHandlerTestSuite) TestGetAll() {
	// Test getting all menus
	perPage := 10
	req := httptest.NewRequest("GET", "/menus", nil)
	resp, err := suite.app.Test(req)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var response model.PaginationResponse[[]entity.Menu]
	err = json.Unmarshal(body, &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), perPage, len(response.Data))
}

func (suite *MenuHandlerTestSuite) TestUpdate() {
	// Create a test menu first
	suite.T().Run("Test successful menu update", func(t *testing.T) {
		menu := &model.CreateUpdateMenuRequest{
			Title:       "Test Cake",
			Description: "Test Description",
			Rating:      4.5,
			Price:       90000,
			Category:    "birthday_cake",
			ImageURL:    "http://example.com/test.jpg",
		}

		// POST: create menu
		jsonValue, _ := json.Marshal(menu)
		req := httptest.NewRequest("POST", "/menus", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+suite.token)

		resp, _ := suite.app.Test(req)
		body, _ := io.ReadAll(resp.Body)

		var respCreate utils.Response
		_ = json.Unmarshal(body, &respCreate)

		dataBytes, _ := json.Marshal(respCreate.Data)
		var createdMenu model.MenuModel
		_ = json.Unmarshal(dataBytes, &createdMenu)

		// Update the menu
		updatedMenu := createdMenu
		updatedMenu.Title = "Updated Cake"
		updatedMenu.Rating = 4.5
		updatedMenu.ImageURL = "http://example.com/updated.jpg"
		updatedMenu.Price = 100
		updatedMenu.Category = "test"
		updatedMenu.Description = "Updated Description"

		testUpdatedMenu := model.CreateUpdateMenuRequest{
			Title:       updatedMenu.Title,
			Description: updatedMenu.Description,
			Rating:      updatedMenu.Rating,
			ImageURL:    updatedMenu.ImageURL,
			Price:       updatedMenu.Price,
			Category:    updatedMenu.Category,
		}

		jsonValue, _ = json.Marshal(testUpdatedMenu)
		req = httptest.NewRequest("PUT", "/menus/"+strconv.Itoa(int(createdMenu.ID)), bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+suite.token)

		resp, err := suite.app.Test(req)
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), fiber.StatusOK, resp.StatusCode)

		body, _ = io.ReadAll(resp.Body)
		var response utils.Response
		_ = json.Unmarshal(body, &response)

		dataBytes, _ = json.Marshal(response.Data)
		var updated model.MenuModel
		_ = json.Unmarshal(dataBytes, &updated)

		assert.Equal(suite.T(), "Updated Menu", updated.Title)

		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), "Updated Menu", updated.Title)
		assert.Equal(suite.T(), 4.5, updated.Rating)
	})

	// Test updating non-existent menu - should return 404
	suite.T().Run("Update non-existent menu", func(t *testing.T) {
		menu := &model.CreateUpdateMenuRequest{
			Title:       "Test Cake",
			Description: "Test Description",
			Price:       90000,
			Category:    "birthday_cake",
			Rating:      4.5,
			ImageURL:    "http://example.com/test.jpg",
		}
		jsonValue, _ := json.Marshal(menu)
		req := httptest.NewRequest("PUT", "/menus/9999999", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+suite.token)
		resp, err := suite.app.Test(req)
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), fiber.StatusNotFound, resp.StatusCode)
	})
}

func (suite *MenuHandlerTestSuite) TestDelete() {
	// Create a test menu first
	menu := &model.CreateUpdateMenuRequest{
		Title:       "Test Cake",
		Description: "Test Description",
		Price:       90000,
		Category:    "birthday_cake",
		Rating:      4.5,
		ImageURL:    "http://example.com/test.jpg",
	}
	jsonValue, _ := json.Marshal(menu)
	req := httptest.NewRequest("POST", "/menus", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.token)
	resp, _ := suite.app.Test(req)

	body, _ := io.ReadAll(resp.Body)
	var res utils.Response
	json.Unmarshal(body, &res)

	jsonBody, _ := json.Marshal(res.Data)
	var result model.MenuModel
	_ = json.Unmarshal(jsonBody, &result)

	// Delete the menu
	req = httptest.NewRequest("DELETE", "/menus/"+strconv.Itoa(int(result.ID)), nil)
	resp, _ = suite.app.Test(req)
	req.Header.Set("Authorization", "Bearer "+suite.token)
	assert.Equal(suite.T(), fiber.StatusOK, resp.StatusCode)

	// Verify menu is deleted
	req = httptest.NewRequest("GET", "/menus/"+strconv.Itoa(int(result.ID)), nil)
	resp, _ = suite.app.Test(req)
	assert.Equal(suite.T(), fiber.StatusNotFound, resp.StatusCode)
}
