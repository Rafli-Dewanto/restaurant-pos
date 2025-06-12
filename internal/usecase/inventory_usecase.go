package usecase

import (
	"cakestore/internal/domain/entity"
	"cakestore/internal/domain/model"
	"cakestore/internal/repository"
	"errors"
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
}

func NewInventoryUseCase(repo repository.InventoryRepository, logger *logrus.Logger) InventoryUseCase {
	return &inventoryUseCase{
		repo:   repo,
		logger: logger,
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
	ingredient, err := u.repo.GetByID(id)
	if err != nil {
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

func (u *inventoryUseCase) GetAll(params *model.InventoryQueryParams) (*model.PaginationResponse[[]model.InventoryResponse], error) {
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

	return &model.PaginationResponse[[]model.InventoryResponse]{
		Data:       responses,
		Total:      result.Total,
		Page:       result.Page,
		TotalPages: result.TotalPages,
	}, nil
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
	return u.repo.Delete(id)
}

func (u *inventoryUseCase) UpdateStock(id uint, quantity float64) error {
	ingredient, err := u.repo.GetByID(id)
	if err != nil {
		return err
	}

	if ingredient.Quantity+quantity < 0 {
		return errors.New("insufficient stock")
	}

	return u.repo.UpdateStock(id, quantity)
}

func (u *inventoryUseCase) GetLowStockIngredients() ([]model.InventoryResponse, error) {
	ingredients, err := u.repo.GetLowStockIngredients()
	if err != nil {
		return nil, err
	}

	responses := make([]model.InventoryResponse, len(ingredients))
	for i, ingredient := range ingredients {
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

	return responses, nil
}
