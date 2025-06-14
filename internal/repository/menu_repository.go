package repository

import (
	"cakestore/internal/constants"
	"cakestore/internal/domain/entity"
	"cakestore/internal/domain/model"
	"errors"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type MenuRepository interface {
	GetAll(params *model.MenuQueryParams) (*model.PaginationResponse[[]entity.Menu], error)
	GetByID(id int64) (*entity.Menu, error)
	Create(menu *entity.Menu) error
	UpdateMenu(menu *entity.Menu) error
	SoftDelete(id int64) error
}

type menuRepository struct {
	db  *gorm.DB
	log *logrus.Logger
}

func NewMenuRepository(db *gorm.DB, log *logrus.Logger) MenuRepository {
	return &menuRepository{db: db, log: log}
}

func (c *menuRepository) GetAll(params *model.MenuQueryParams) (*model.PaginationResponse[[]entity.Menu], error) {
	var menus []entity.Menu
	var total int64

	query := c.db.Model(&entity.Menu{}).Where("deleted_at IS NULL")

	if params.Title != "" {
		query = query.Where("LOWER(title) LIKE LOWER(?)", "%"+params.Title+"%")
	}
	if params.MinPrice > 0 {
		query = query.Where("price >= ?", params.MinPrice)
	}
	if params.MaxPrice > 0 {
		query = query.Where("price <= ?", params.MaxPrice)
	}
	if params.Category != "" {
		query = query.Where("LOWER(category) LIKE LOWER(?)", "%"+params.Category+"%")
	}

	// Get total count for pagination
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	// Set default values for pagination
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PageSize < 1 {
		if params.Limit == 0 {
			params.PageSize = 10
		} else {
			params.PageSize = params.Limit
		}
	}

	// Apply pagination and ordering
	offset := (params.Page - 1) * params.PageSize
	err := query.Order("rating DESC, title ASC").Offset(int(offset)).Limit(int(params.PageSize)).Find(&menus).Error
	if err != nil {
		return nil, err
	}

	totalPages := int64(total) / params.PageSize
	if int64(total)%params.PageSize != 0 {
		totalPages++
	}

	return &model.PaginationResponse[[]entity.Menu]{
		Data:       menus,
		Total:      total,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: totalPages,
	}, nil
}

func (c *menuRepository) GetByID(id int64) (*entity.Menu, error) {
	var menu entity.Menu
	err := c.db.Where("deleted_at IS NULL").First(&menu, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, constants.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &menu, nil
}

func (c *menuRepository) Create(menu *entity.Menu) error {
	return c.db.Create(menu).Error
}

func (c *menuRepository) UpdateMenu(menu *entity.Menu) error {
	result := c.db.Model(&entity.Menu{}).
		Where("id = ?", menu.ID).
		Updates(map[string]interface{}{
			"title":       menu.Title,
			"description": menu.Description,
			"rating":      menu.Rating,
			"image":       menu.Image,
			"quantity":    menu.Quantity,
			"price":       menu.Price,
			"category":    menu.Category,
			"updated_at":  time.Now(),
		})

	if result.RowsAffected == 0 {
		return constants.ErrNotFound
	}

	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (c *menuRepository) SoftDelete(id int64) error {
	result := c.db.Model(&entity.Menu{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"deleted_at": time.Now(),
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("no rows updated, menu not found")
	}

	return nil
}
