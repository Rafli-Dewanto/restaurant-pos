package repository

import (
	"cakestore/internal/domain/entity"
	"errors"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type CakeRepository interface {
	GetAll() ([]entity.Cake, error)
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

func (c *cakeRepository) GetAll() ([]entity.Cake, error) {
	var cakes []entity.Cake
	err := c.db.Order("rating DESC, title ASC").Where("deleted_at IS NULL").Find(&cakes).Error
	if err != nil {
		return nil, err
	}
	return cakes, nil
}

func (c *cakeRepository) GetByID(id int) (*entity.Cake, error) {
	var cake entity.Cake
	err := c.db.Where("deleted_at IS NULL").First(&cake, id).Error
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

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("no rows updated, cake not found")
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
