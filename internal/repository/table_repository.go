package repository

import (
	"cakestore/internal/domain/entity"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type TableRepository interface {
	Count() (int64, error)
	Create(table *entity.Table) error
	GetByID(id uint) (*entity.Table, error)
	GetAll() ([]entity.Table, error)
	Update(table *entity.Table) error
	Delete(id uint) error
	GetAvailableTables(reserveTime time.Time, duration time.Duration) ([]entity.Table, error)
	UpdateAvailability(id uint, isAvailable bool) error
}

type tableRepository struct {
	db  *gorm.DB
	log *logrus.Logger
}

func NewTableRepository(db *gorm.DB, log *logrus.Logger) TableRepository {
	return &tableRepository{db: db, log: log}
}

func (r *tableRepository) Count() (int64, error) {
	var count int64
	if err := r.db.Model(&entity.Table{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *tableRepository) Create(table *entity.Table) error {
	return r.db.Create(table).Error
}

func (r *tableRepository) GetByID(id uint) (*entity.Table, error) {
	var table entity.Table
	if err := r.db.First(&table, id).Error; err != nil {
		return nil, err
	}
	return &table, nil
}

func (r *tableRepository) GetAll() ([]entity.Table, error) {
	var tables []entity.Table
	if err := r.db.Find(&tables).Error; err != nil {
		return nil, err
	}
	return tables, nil
}

func (r *tableRepository) Update(table *entity.Table) error {
	return r.db.Save(table).Error
}

func (r *tableRepository) Delete(id uint) error {
	return r.db.Delete(&entity.Table{}, id).Error
}

func (r *tableRepository) GetAvailableTables(reserveTime time.Time, duration time.Duration) ([]entity.Table, error) {
	var tables []entity.Table
	endTime := reserveTime.Add(duration)

	subQuery := r.db.Model(&entity.Reservation{}).Select("table_id").Where(
		"(reserved_at BETWEEN ? AND ?) OR (reserved_at + duration * interval '1 minute' BETWEEN ? AND ?)",
		reserveTime, endTime, reserveTime, endTime,
	)

	if err := r.db.Where("id NOT IN (?) AND is_available = ?", subQuery, true).Find(&tables).Error; err != nil {
		return nil, err
	}

	return tables, nil
}

func (r *tableRepository) UpdateAvailability(id uint, isAvailable bool) error {
	return r.db.Model(&entity.Table{}).Where("id = ?", id).Update("is_available", isAvailable).Error
}
