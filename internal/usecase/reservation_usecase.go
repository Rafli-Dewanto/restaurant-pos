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

type ReservationUseCase interface {
	Create(customerID uint, request *model.CreateReservationRequest) (*model.ReservationResponse, error)
	GetByID(id uint) (*model.ReservationResponse, error)
	GetAll(params *model.ReservationQueryParams) (*model.PaginationResponse[[]model.ReservationResponse], error)
	AdminGetAllCustomerReservations(params *model.PaginationQuery) (*model.PaginationResponse[[]model.ReservationResponse], error)
	Update(id uint, request *model.UpdateReservationRequest) (*model.ReservationResponse, error)
	Delete(id uint) error
}

type reservationUseCase struct {
	repo            repository.ReservationRepository
	tableRepository repository.TableRepository
	logger          *logrus.Logger
	cache           database.RedisCache
}

func NewReservationUseCase(
	repo repository.ReservationRepository,
	logger *logrus.Logger,
	tableRepository repository.TableRepository,
	cache database.RedisCache,
) ReservationUseCase {
	return &reservationUseCase{
		repo:            repo,
		logger:          logger,
		tableRepository: tableRepository,
		cache:           cache,
	}
}

func (u *reservationUseCase) AdminGetAllCustomerReservations(params *model.PaginationQuery) (*model.PaginationResponse[[]model.ReservationResponse], error) {
	start := time.Now()
	defer func() {
		u.logger.Infof("AdminGetAllCustomerReservations took %v", time.Since(start))
	}()

	// Try to get the reservations from the cache first
	cacheKey := fmt.Sprintf("reservations:admin:all:page:%d:limit:%d", params.Page, params.Limit)
	var cachedData model.PaginationResponse[[]model.ReservationResponse]
	if err := u.cache.Get(context.Background(), cacheKey, &cachedData); err == nil {
		u.logger.Info("Admin reservations fetched from cache")
		return &cachedData, nil
	}

	// If not in cache, get from the database
	result, err := u.repo.AdminGetAllCustomerReservations(params)
	if err != nil {
		return nil, err
	}

	responses := make([]model.ReservationResponse, len(result.Data))
	for i, reservation := range result.Data {
		responses[i] = model.ReservationResponse{
			ID:           reservation.ID,
			CustomerID:   reservation.CustomerID,
			Customer:     *model.ToCustomerResponse(&reservation.Customer),
			TableNumber:  reservation.TableNumber,
			GuestCount:   reservation.GuestCount,
			ReserveDate:  reservation.ReserveDate,
			Status:       string(reservation.Status),
			SpecialNotes: reservation.SpecialNotes,
			CreatedAt:    reservation.CreatedAt,
			UpdatedAt:    reservation.UpdatedAt,
		}
	}

	paginatedResponse := &model.PaginationResponse[[]model.ReservationResponse]{
		Data:       responses,
		Total:      result.Total,
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalPages: result.TotalPages,
	}

	// Store the reservations in the cache for future requests
	if err := u.cache.Set(context.Background(), cacheKey, paginatedResponse, 5*time.Minute); err != nil {
		u.logger.Errorf("Error setting cache for admin reservations: %v", err)
	}

	return paginatedResponse, nil
}

func (u *reservationUseCase) Create(customerID uint, request *model.CreateReservationRequest) (*model.ReservationResponse, error) {
	if err := request.Validate(); err != nil {
		return nil, err
	}

	var table *entity.Table
	var tableNumber int

	// Only get table and set relation if TableID is provided (not zero)
	if request.TableID != 0 {
		var err error
		table, err = u.tableRepository.GetByID(request.TableID)
		if err != nil {
			u.logger.Errorf("Error getting table: %v", err)
			return nil, err
		}

		// Check table availability
		isAvailable, err := u.repo.CheckTableAvailability(request.TableID, request.ReserveDate)
		if err != nil {
			return nil, err
		}
		if !isAvailable {
			return nil, errors.New("table is not available for the selected date")
		}
		tableNumber = table.TableNumber
	}

	reservation := &entity.Reservation{
		CustomerID:   customerID,
		GuestCount:   request.GuestCount,
		ReserveDate:  request.ReserveDate,
		Status:       entity.ReservationStatusPending,
		SpecialNotes: request.SpecialNotes,
	}

	// Only set TableID, Table, and TableNumber if TableID is provided
	if request.TableID != 0 {
		reservation.TableID = &request.TableID
		reservation.Table = table
		reservation.TableNumber = tableNumber
	}

	if err := u.repo.Create(reservation); err != nil {
		return nil, err
	}

	// Get the created reservation with customer details
	createdReservation, err := u.repo.GetByID(reservation.ID)
	if err != nil {
		return nil, err
	}

	return &model.ReservationResponse{
		ID:           createdReservation.ID,
		CustomerID:   createdReservation.CustomerID,
		Customer:     *model.ToCustomerResponse(&createdReservation.Customer),
		TableNumber:  createdReservation.TableNumber,
		GuestCount:   createdReservation.GuestCount,
		ReserveDate:  createdReservation.ReserveDate,
		Status:       string(createdReservation.Status),
		SpecialNotes: createdReservation.SpecialNotes,
		CreatedAt:    createdReservation.CreatedAt,
		UpdatedAt:    createdReservation.UpdatedAt,
	}, nil
}

func (u *reservationUseCase) GetByID(id uint) (*model.ReservationResponse, error) {
	start := time.Now()
	defer func() {
		u.logger.Infof("GetByID took %v", time.Since(start))
	}()

	// Try to get the reservation from the cache first
	cacheKey := fmt.Sprintf("reservation:%d", id)
	var reservation model.ReservationResponse
	if err := u.cache.Get(context.Background(), cacheKey, &reservation); err == nil {
		u.logger.Info("Reservation fetched from cache")
		return &reservation, nil
	}

	// If not in cache, get from the database
	reservationEntity, err := u.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Store the reservation in the cache for future requests
	reservationModel := &model.ReservationResponse{
		ID:           reservationEntity.ID,
		CustomerID:   reservationEntity.CustomerID,
		Customer:     *model.ToCustomerResponse(&reservationEntity.Customer),
		TableNumber:  reservationEntity.TableNumber,
		GuestCount:   reservationEntity.GuestCount,
		ReserveDate:  reservationEntity.ReserveDate,
		Status:       string(reservationEntity.Status),
		SpecialNotes: reservationEntity.SpecialNotes,
		CreatedAt:    reservationEntity.CreatedAt,
		UpdatedAt:    reservationEntity.UpdatedAt,
	}
	if err := u.cache.Set(context.Background(), cacheKey, reservationModel, 5*time.Minute); err != nil {
		u.logger.Errorf("Error setting cache for reservation ID %d: %v", id, err)
	}

	return reservationModel, nil
}

func (u *reservationUseCase) GetAll(params *model.ReservationQueryParams) (*model.PaginationResponse[[]model.ReservationResponse], error) {
	start := time.Now()
	defer func() {
		u.logger.Infof("GetAll took %v", time.Since(start))
	}()

	// Try to get the reservations from the cache first
	cacheKey := fmt.Sprintf("reservations:all:page:%d:limit:%d", params.Page, params.Limit)
	var cachedData model.PaginationResponse[[]model.ReservationResponse]
	if err := u.cache.Get(context.Background(), cacheKey, &cachedData); err == nil {
		u.logger.Info("Reservations fetched from cache")
		return &cachedData, nil
	}

	// If not in cache, get from the database
	result, err := u.repo.GetAll(params)
	if err != nil {
		return nil, err
	}

	responses := make([]model.ReservationResponse, len(result.Data))
	for i, reservation := range result.Data {
		responses[i] = model.ReservationResponse{
			ID:           reservation.ID,
			CustomerID:   reservation.CustomerID,
			Customer:     *model.ToCustomerResponse(&reservation.Customer),
			TableNumber:  reservation.TableNumber,
			GuestCount:   reservation.GuestCount,
			ReserveDate:  reservation.ReserveDate,
			Status:       string(reservation.Status),
			SpecialNotes: reservation.SpecialNotes,
			CreatedAt:    reservation.CreatedAt,
			UpdatedAt:    reservation.UpdatedAt,
		}
	}

	paginatedResponse := &model.PaginationResponse[[]model.ReservationResponse]{
		Data:       responses,
		Total:      result.Total,
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalPages: result.TotalPages,
	}

	// Store the reservations in the cache for future requests
	if err := u.cache.Set(context.Background(), cacheKey, paginatedResponse, 5*time.Minute); err != nil {
		u.logger.Errorf("Error setting cache for all reservations: %v", err)
	}

	return paginatedResponse, nil
}

func (u *reservationUseCase) Update(id uint, request *model.UpdateReservationRequest) (*model.ReservationResponse, error) {
	existing, err := u.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Check table availability if table number or date is being updated
	if (request.TableNumber != 0 && request.TableNumber != existing.TableNumber) ||
		(!request.ReserveDate.IsZero() && !request.ReserveDate.Equal(existing.ReserveDate)) {
		checkDate := request.ReserveDate
		if checkDate.IsZero() {
			checkDate = existing.ReserveDate
		}
		checkTable := request.TableNumber
		if checkTable == 0 {
			checkTable = existing.TableNumber
		}

		available, err := u.repo.CheckTableAvailability(uint(checkTable), checkDate)
		if err != nil {
			return nil, err
		}
		if !available {
			return nil, errors.New("table is not available for the selected date")
		}
	}

	if request.TableNumber != 0 {
		existing.TableNumber = request.TableNumber
	}
	if request.GuestCount != 0 {
		existing.GuestCount = request.GuestCount
	}
	if !request.ReserveDate.IsZero() {
		existing.ReserveDate = request.ReserveDate
	}
	if request.Status != "" {
		existing.Status = entity.ReservationStatus(request.Status)
	}
	if request.SpecialNotes != "" {
		existing.SpecialNotes = request.SpecialNotes
	}

	if err := u.repo.Update(existing); err != nil {
		return nil, err
	}

	// Invalidate cache
	cacheKey := fmt.Sprintf("reservation:%d", id)
	if err := u.cache.Delete(context.Background(), cacheKey); err != nil {
		u.logger.Errorf("Error deleting cache for reservation ID %d: %v", id, err)
	}
	if err := u.cache.Delete(context.Background(), "reservations:all:*"); err != nil {
		u.logger.Errorf("Error deleting cache for all reservations: %v", err)
	}
	if err := u.cache.Delete(context.Background(), "reservations:admin:all:*"); err != nil {
		u.logger.Errorf("Error deleting cache for admin reservations: %v", err)
	}

	updated, err := u.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	return &model.ReservationResponse{
		ID:           updated.ID,
		CustomerID:   updated.CustomerID,
		Customer:     *model.ToCustomerResponse(&updated.Customer),
		TableNumber:  updated.TableNumber,
		GuestCount:   updated.GuestCount,
		ReserveDate:  updated.ReserveDate,
		Status:       string(updated.Status),
		SpecialNotes: updated.SpecialNotes,
		CreatedAt:    updated.CreatedAt,
		UpdatedAt:    updated.UpdatedAt,
	}, nil
}

func (u *reservationUseCase) Delete(id uint) error {
	if err := u.repo.Delete(id); err != nil {
		return err
	}

	// Invalidate cache
	cacheKey := fmt.Sprintf("reservation:%d", id)
	if err := u.cache.Delete(context.Background(), cacheKey); err != nil {
		u.logger.Errorf("Error deleting cache for reservation ID %d: %v", id, err)
	}
	if err := u.cache.Delete(context.Background(), "reservations:all:*"); err != nil {
		u.logger.Errorf("Error deleting cache for all reservations: %v", err)
	}
	if err := u.cache.Delete(context.Background(), "reservations:admin:all:*"); err != nil {
		u.logger.Errorf("Error deleting cache for admin reservations: %v", err)
	}

	return nil
}
