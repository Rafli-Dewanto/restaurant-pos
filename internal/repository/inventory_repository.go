package repository

import (
	"cakestore/internal/domain/entity"
	"cakestore/internal/domain/model"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type InventoryRepository interface {
	Create(ingredient *entity.Inventory) error
	GetByID(id uint) (*entity.Inventory, error)
	GetAll(params *model.InventoryQueryParams) (*model.PaginationResponse[[]entity.Inventory], error)
	Update(ingredient *entity.Inventory) error
	Delete(id uint) error
	UpdateStock(id uint, quantity float64) error
	GetLowStockIngredients() ([]entity.Inventory, error)
	Count() (int64, error)
}

type inventoryRepository struct {
	db     *gorm.DB
	logger *logrus.Logger
}

func NewInventoryRepository(db *gorm.DB, logger *logrus.Logger) InventoryRepository {
	return &inventoryRepository{
		db:     db,
		logger: logger,
	}
}

func (r *inventoryRepository) Count() (int64, error) {
	var count int64
	if err := r.db.Model(&entity.Inventory{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *inventoryRepository) Create(ingredient *entity.Inventory) error {
	return r.db.Create(ingredient).Error
}

func (r *inventoryRepository) GetByID(id uint) (*entity.Inventory, error) {
	var ingredient entity.Inventory
	if err := r.db.First(&ingredient, id).Error; err != nil {
		return nil, err
	}
	return &ingredient, nil
}

func (r *inventoryRepository) GetAll(params *model.InventoryQueryParams) (*model.PaginationResponse[[]entity.Inventory], error) {
	var ingredients []entity.Inventory
	var total int64

	query := r.db.Model(&entity.Inventory{})

	if params.Search != "" {
		query = query.Where("name ILIKE ?", "%"+params.Search+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	offset := (params.Page - 1) * params.Limit
	if err := query.Offset(int(offset)).Limit(int(params.Limit)).Find(&ingredients).Error; err != nil {
		return nil, err
	}

	totalPages := (total + params.Limit - 1) / params.Limit

	return &model.PaginationResponse[[]entity.Inventory]{
		Data:       ingredients,
		Total:      total,
		Page:       params.Page,
		TotalPages: totalPages,
	}, nil
}

func (r *inventoryRepository) Update(ingredient *entity.Inventory) error {
	return r.db.Save(ingredient).Error
}

func (r *inventoryRepository) Delete(id uint) error {
	return r.db.Delete(&entity.Inventory{}, id).Error
}

func (r *inventoryRepository) UpdateStock(id uint, quantity float64) error {
	return r.db.Model(&entity.Inventory{}).Where("id = ?", id).UpdateColumn("quantity", gorm.Expr("quantity + ?", quantity)).Error
}

func (r *inventoryRepository) GetLowStockIngredients() ([]entity.Inventory, error) {
	var ingredients []entity.Inventory
	if err := r.db.Where("quantity <= minimum_stock").Find(&ingredients).Error; err != nil {
		return nil, err
	}
	return ingredients, nil
}
