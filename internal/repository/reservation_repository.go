package repository

import (
	"cakestore/internal/domain/entity"
	"cakestore/internal/domain/model"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type ReservationRepository interface {
	Create(reservation *entity.Reservation) error
	GetByID(id uint) (*entity.Reservation, error)
	GetAll(params *model.ReservationQueryParams) (*model.PaginationResponse[[]entity.Reservation], error)
	Update(reservation *entity.Reservation) error
	Delete(id uint) error
	CheckTableAvailability(tableNumber int, reserveDate time.Time) (bool, error)
}

type reservationRepository struct {
	db     *gorm.DB
	logger *logrus.Logger
}

func NewReservationRepository(db *gorm.DB, logger *logrus.Logger) ReservationRepository {
	return &reservationRepository{
		db:     db,
		logger: logger,
	}
}

func (r *reservationRepository) Create(reservation *entity.Reservation) error {
	if err := r.db.Create(reservation).Error; err != nil {
		r.logger.Errorf("Error creating reservation: %v", err)
		return err
	}
	return nil
}

func (r *reservationRepository) GetByID(id uint) (*entity.Reservation, error) {
	var reservation entity.Reservation
	if err := r.db.Preload("Customer").First(&reservation, id).Error; err != nil {
		r.logger.Errorf("Error getting reservation by ID: %v", err)
		return nil, err
	}
	return &reservation, nil
}

func (r *reservationRepository) GetAll(params *model.ReservationQueryParams) (*model.PaginationResponse[[]entity.Reservation], error) {
	var reservations []entity.Reservation
	var total int64

	query := r.db.Model(&entity.Reservation{})

	if params.CustomerID != 0 {
		query = query.Where("customer_id = ?", params.CustomerID)
	}

	if params.Status != "" {
		query = query.Where("status = ?", params.Status)
	}

	if !params.ReserveDate.IsZero() {
		start := params.ReserveDate.Truncate(24 * time.Hour)
		end := start.Add(24 * time.Hour)
		query = query.Where("reserve_date BETWEEN ? AND ?", start, end)
	}

	if params.TableNumber != 0 {
		query = query.Where("table_number = ?", params.TableNumber)
	}

	if err := query.Count(&total).Error; err != nil {
		r.logger.Errorf("Error counting reservations: %v", err)
		return nil, err
	}

	offset := (params.Page - 1) * params.Page
	query = query.Offset(int(offset)).Limit(int(params.Page))
	query = query.Preload("Customer")

	if err := query.Find(&reservations).Error; err != nil {
		r.logger.Errorf("Error getting reservations: %v", err)
		return nil, err
	}

	return &model.PaginationResponse[[]entity.Reservation]{
		Data:       reservations,
		Total:      total,
		Page:       params.Page,
		TotalPages: (total + int64(params.Limit) - 1) / int64(params.Limit),
	}, nil
}

func (r *reservationRepository) Update(reservation *entity.Reservation) error {
	if err := r.db.Save(reservation).Error; err != nil {
		r.logger.Errorf("Error updating reservation: %v", err)
		return err
	}
	return nil
}

func (r *reservationRepository) Delete(id uint) error {
	if err := r.db.Delete(&entity.Reservation{}, id).Error; err != nil {
		r.logger.Errorf("Error deleting reservation: %v", err)
		return err
	}
	return nil
}

func (r *reservationRepository) CheckTableAvailability(tableNumber int, reserveDate time.Time) (bool, error) {
	var count int64
	start := reserveDate.Truncate(24 * time.Hour)
	end := start.Add(24 * time.Hour)

	if err := r.db.Model(&entity.Reservation{}).Where(
		"table_number = ? AND reserve_date BETWEEN ? AND ? AND status NOT IN ?",
		tableNumber,
		start,
		end,
		[]string{string(entity.ReservationStatusCancelled), string(entity.ReservationStatusCompleted)},
	).Count(&count).Error; err != nil {
		r.logger.Errorf("Error checking table availability: %v", err)
		return false, err
	}

	return count == 0, nil
}
