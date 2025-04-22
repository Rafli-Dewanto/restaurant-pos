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

type CakeRepository interface {
	GetAll(params *model.CakeQueryParams) (*model.PaginationResponse[[]entity.Cake], error)
	GetByID(id int) (*entity.Cake, error)
	Create(cake *entity.Cake) error
	UpdateCake(cake *entity.Cake) error
	SoftDelete(id int) error
}

type cakeRepository struct {
	db  *gorm.DB
	log *logrus.Logger
}

func NewCakeRepository(db *gorm.DB, log *logrus.Logger) CakeRepository {
	return &cakeRepository{db: db, log: log}
}

func (c *cakeRepository) GetAll(params *model.CakeQueryParams) (*model.PaginationResponse[[]entity.Cake], error) {
	var cakes []entity.Cake
	var total int64

	query := c.db.Model(&entity.Cake{}).Where("deleted_at IS NULL")

	if params.Title != "" {
		query = query.Where("title LIKE ?", "%"+params.Title+"%")
	}
	if params.MinPrice > 0 {
		query = query.Where("price >= ?", params.MinPrice)
	}
	if params.MaxPrice > 0 {
		query = query.Where("price <= ?", params.MaxPrice)
	}
	if params.Category != "" {
		query = query.Where("category = ?", params.Category)
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
	err := query.Order("rating DESC, title ASC").Offset(offset).Limit(params.PageSize).Find(&cakes).Error
	if err != nil {
		return nil, err
	}

	totalPages := int(total) / params.PageSize
	if int(total)%params.PageSize != 0 {
		totalPages++
	}

	return &model.PaginationResponse[[]entity.Cake]{
		Data:       cakes,
		Total:      total,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: totalPages,
	}, nil
}

func (c *cakeRepository) GetByID(id int) (*entity.Cake, error) {
	var cake entity.Cake
	err := c.db.Where("deleted_at IS NULL").First(&cake, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, constants.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &cake, nil
}

func (c *cakeRepository) Create(cake *entity.Cake) error {
	return c.db.Create(cake).Error
}

func (c *cakeRepository) UpdateCake(cake *entity.Cake) error {
	result := c.db.Model(&entity.Cake{}).
		Where("id = ?", cake.ID).
		Updates(map[string]interface{}{
			"title":       cake.Title,
			"description": cake.Description,
			"rating":      cake.Rating,
			"image":       cake.Image,
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

func (c *cakeRepository) SoftDelete(id int) error {
	result := c.db.Model(&entity.Cake{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"deleted_at": time.Now(),
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("no rows updated, cake not found")
	}

	return nil
}
