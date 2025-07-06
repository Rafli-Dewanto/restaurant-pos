package usecase

import (
	"cakestore/internal/database"
	"cakestore/internal/domain/entity"
	"cakestore/internal/domain/model"
	"cakestore/internal/repository"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

type InventoryUseCase interface {
	Create(request *model.CreateInventoryRequest) (*model.InventoryResponse, error)
	GetByID(id uint) (*model.InventoryResponse, error)
	GetAll(params *model.InventoryQueryParams) (*model.PaginationResponse[[]model.InventoryResponse], error)
	Update(id uint, request *model.UpdateInventoryRequest) (*model.InventoryResponse, error)
	Delete(id uint) error
	UpdateStock(id uint, quantity float64) error
	GetLowStockIngredients() ([]model.InventoryResponse, error)
}

type inventoryUseCase struct {
	repo   repository.InventoryRepository
	logger *logrus.Logger
	cache  database.RedisCache
}

func NewInventoryUseCase(repo repository.InventoryRepository, logger *logrus.Logger, cache database.RedisCache) InventoryUseCase {
	return &inventoryUseCase{
		repo:   repo,
		logger: logger,
		cache:  cache,
	}
}

func (u *inventoryUseCase) Create(request *model.CreateInventoryRequest) (*model.InventoryResponse, error) {
	ingredient := &entity.Inventory{
		Name:            request.Name,
		Quantity:        request.Quantity,
		Unit:            request.Unit,
		MinimumStock:    request.MinimumStock,
		ReorderPoint:    request.ReorderPoint,
		UnitPrice:       request.UnitPrice,
		LastRestockDate: time.Now(),
	}

	if err := u.repo.Create(ingredient); err != nil {
		return nil, err
	}

	return &model.InventoryResponse{
		ID:              ingredient.ID,
		Name:            ingredient.Name,
		Quantity:        ingredient.Quantity,
		Unit:            ingredient.Unit,
		MinimumStock:    ingredient.MinimumStock,
		ReorderPoint:    ingredient.ReorderPoint,
		UnitPrice:       ingredient.UnitPrice,
		LastRestockDate: ingredient.LastRestockDate,
		CreatedAt:       ingredient.CreatedAt,
		UpdatedAt:       ingredient.UpdatedAt,
	}, nil
}

func (u *inventoryUseCase) GetByID(id uint) (*model.InventoryResponse, error) {
	start := time.Now()
	defer func() {
		u.logger.Infof("GetByID took %v", time.Since(start))
	}()

	// Try to get the ingredient from the cache first
	cacheKey := fmt.Sprintf("inventory:%d", id)
	var ingredient model.InventoryResponse
	if err := u.cache.Get(context.Background(), cacheKey, &ingredient); err == nil {
		u.logger.Info("Ingredient fetched from cache")
		return &ingredient, nil
	}

	// If not in cache, get from the database
	ingredientEntity, err := u.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Store the ingredient in the cache for future requests
	ingredientModel := &model.InventoryResponse{
		ID:              ingredientEntity.ID,
		Name:            ingredientEntity.Name,
		Quantity:        ingredientEntity.Quantity,
		Unit:            ingredientEntity.Unit,
		MinimumStock:    ingredientEntity.MinimumStock,
		ReorderPoint:    ingredientEntity.ReorderPoint,
		UnitPrice:       ingredientEntity.UnitPrice,
		LastRestockDate: ingredientEntity.LastRestockDate,
		CreatedAt:       ingredientEntity.CreatedAt,
		UpdatedAt:       ingredientEntity.UpdatedAt,
	}
	if err := u.cache.Set(context.Background(), cacheKey, ingredientModel, 5*time.Minute); err != nil {
		u.logger.Errorf("Error setting cache for ingredient ID %d: %v", id, err)
	}

	return ingredientModel, nil
}

func (u *inventoryUseCase) GetAll(params *model.InventoryQueryParams) (*model.PaginationResponse[[]model.InventoryResponse], error) {
	start := time.Now()
	defer func() {
		u.logger.Infof("GetAll took %v", time.Since(start))
	}()

	// Try to get the ingredients from the cache first
	cacheKey := fmt.Sprintf("inventory:all:page:%d:limit:%d", params.Page, params.Limit)
	var cachedData model.PaginationResponse[[]model.InventoryResponse]
	if err := u.cache.Get(context.Background(), cacheKey, &cachedData); err == nil {
		u.logger.Info("Ingredients fetched from cache")
		return &cachedData, nil
	}

	// If not in cache, get from the database
	result, err := u.repo.GetAll(params)
	if err != nil {
		return nil, err
	}

	responses := make([]model.InventoryResponse, len(result.Data))
	for i, ingredient := range result.Data {
		responses[i] = model.InventoryResponse{
			ID:              ingredient.ID,
			Name:            ingredient.Name,
			Quantity:        ingredient.Quantity,
			Unit:            ingredient.Unit,
			MinimumStock:    ingredient.MinimumStock,
			ReorderPoint:    ingredient.ReorderPoint,
			UnitPrice:       ingredient.UnitPrice,
			LastRestockDate: ingredient.LastRestockDate,
			CreatedAt:       ingredient.CreatedAt,
			UpdatedAt:       ingredient.UpdatedAt,
		}
	}

	paginatedResponse := &model.PaginationResponse[[]model.InventoryResponse]{
		Data:       responses,
		Total:      result.Total,
		Page:       result.Page,
		TotalPages: result.TotalPages,
	}

	// Store the ingredients in the cache for future requests
	if err := u.cache.Set(context.Background(), cacheKey, paginatedResponse, 5*time.Minute); err != nil {
		u.logger.Errorf("Error setting cache for all ingredients: %v", err)
	}

	return paginatedResponse, nil
}

func (u *inventoryUseCase) Update(id uint, request *model.UpdateInventoryRequest) (*model.InventoryResponse, error) {
	existing, err := u.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	if request.Name != "" {
		existing.Name = request.Name
	}
	if request.Quantity > 0 {
		existing.Quantity = request.Quantity
	}
	if request.Unit != "" {
		existing.Unit = request.Unit
	}
	if request.MinimumStock > 0 {
		existing.MinimumStock = request.MinimumStock
	}
	if request.ReorderPoint > 0 {
		existing.ReorderPoint = request.ReorderPoint
	}
	if request.UnitPrice > 0 {
		existing.UnitPrice = request.UnitPrice
	}

	if err := u.repo.Update(existing); err != nil {
		return nil, err
	}

	// Invalidate cache
	cacheKey := fmt.Sprintf("inventory:%d", id)
	if err := u.cache.Delete(context.Background(), cacheKey); err != nil {
		u.logger.Errorf("Error deleting cache for ingredient ID %d: %v", id, err)
	}
	if err := u.cache.Delete(context.Background(), "inventory:all:*"); err != nil {
		u.logger.Errorf("Error deleting cache for all ingredients: %v", err)
	}
	if err := u.cache.Delete(context.Background(), "low_stock_ingredients"); err != nil {
		u.logger.Errorf("Error deleting cache for low stock ingredients: %v", err)
	}

	return &model.InventoryResponse{
		ID:              existing.ID,
		Name:            existing.Name,
		Quantity:        existing.Quantity,
		Unit:            existing.Unit,
		MinimumStock:    existing.MinimumStock,
		ReorderPoint:    existing.ReorderPoint,
		UnitPrice:       existing.UnitPrice,
		LastRestockDate: existing.LastRestockDate,
		CreatedAt:       existing.CreatedAt,
		UpdatedAt:       existing.UpdatedAt,
	}, nil
}

func (u *inventoryUseCase) Delete(id uint) error {
	if err := u.repo.Delete(id); err != nil {
		return err
	}

	// Invalidate cache
	cacheKey := fmt.Sprintf("inventory:%d", id)
	if err := u.cache.Delete(context.Background(), cacheKey); err != nil {
		u.logger.Errorf("Error deleting cache for ingredient ID %d: %v", id, err)
	}
	if err := u.cache.Delete(context.Background(), "inventory:all:*"); err != nil {
		u.logger.Errorf("Error deleting cache for all ingredients: %v", err)
	}
	if err := u.cache.Delete(context.Background(), "low_stock_ingredients"); err != nil {
		u.logger.Errorf("Error deleting cache for low stock ingredients: %v", err)
	}

	return nil
}

func (u *inventoryUseCase) UpdateStock(id uint, quantity float64) error {
	ingredient, err := u.repo.GetByID(id)
	if err != nil {
		return err
	}

	if ingredient.Quantity+quantity < 0 {
		return errors.New("insufficient stock")
	}

	if err := u.repo.UpdateStock(id, quantity); err != nil {
		return err
	}

	// Invalidate cache
	cacheKey := fmt.Sprintf("inventory:%d", id)
	if err := u.cache.Delete(context.Background(), cacheKey); err != nil {
		u.logger.Errorf("Error deleting cache for ingredient ID %d: %v", id, err)
	}
	if err := u.cache.Delete(context.Background(), "inventory:all:*"); err != nil {
		u.logger.Errorf("Error deleting cache for all ingredients: %v", err)
	}
	if err := u.cache.Delete(context.Background(), "low_stock_ingredients"); err != nil {
		u.logger.Errorf("Error deleting cache for low stock ingredients: %v", err)
	}

	return nil
}

func (u *inventoryUseCase) GetLowStockIngredients() ([]model.InventoryResponse, error) {
	start := time.Now()
	defer func() {
		u.logger.Infof("GetLowStockIngredients took %v", time.Since(start))
	}()

	// Try to get the ingredients from the cache first
	cacheKey := "low_stock_ingredients"
	var ingredients []model.InventoryResponse
	if err := u.cache.Get(context.Background(), cacheKey, &ingredients); err == nil {
		u.logger.Info("Low stock ingredients fetched from cache")
		return ingredients, nil
	}

	// If not in cache, get from the database
	ingredientEntities, err := u.repo.GetLowStockIngredients()
	if err != nil {
		return nil, err
	}

	responses := make([]model.InventoryResponse, len(ingredientEntities))
	for i, ingredient := range ingredientEntities {
		responses[i] = model.InventoryResponse{
			ID:              ingredient.ID,
			Name:            ingredient.Name,
			Quantity:        ingredient.Quantity,
			Unit:            ingredient.Unit,
			MinimumStock:    ingredient.MinimumStock,
			ReorderPoint:    ingredient.ReorderPoint,
			UnitPrice:       ingredient.UnitPrice,
			LastRestockDate: ingredient.LastRestockDate,
			CreatedAt:       ingredient.CreatedAt,
			UpdatedAt:       ingredient.UpdatedAt,
		}
	}

	// Store the ingredients in the cache for future requests
	if err := u.cache.Set(context.Background(), cacheKey, responses, 5*time.Minute); err != nil {
		u.logger.Errorf("Error setting cache for low stock ingredients: %v", err)
	}

	return responses, nil
}
