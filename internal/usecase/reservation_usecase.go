package usecase

import (
	"cakestore/internal/domain/entity"
	"cakestore/internal/domain/model"
	"cakestore/internal/repository"
	"errors"

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
}

func NewReservationUseCase(
	repo repository.ReservationRepository,
	logger *logrus.Logger,
	tableRepository repository.TableRepository,
) ReservationUseCase {
	return &reservationUseCase{
		repo:            repo,
		logger:          logger,
		tableRepository: tableRepository,
	}
}

func (u *reservationUseCase) AdminGetAllCustomerReservations(params *model.PaginationQuery) (*model.PaginationResponse[[]model.ReservationResponse], error) {
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

	return &model.PaginationResponse[[]model.ReservationResponse]{
		Data:       responses,
		Total:      result.Total,
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalPages: result.TotalPages,
	}, nil
}

func (u *reservationUseCase) Create(customerID uint, request *model.CreateReservationRequest) (*model.ReservationResponse, error) {
	if err := request.Validate(); err != nil {
		return nil, err
	}

	// Get table by ID
	table, err := u.tableRepository.GetByID(request.TableID)
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

	reservation := &entity.Reservation{
		CustomerID:   customerID,
		TableID:      &request.TableID,
		Table:        table,
		TableNumber:  table.TableNumber,
		GuestCount:   request.GuestCount,
		ReserveDate:  request.ReserveDate,
		Status:       entity.ReservationStatusPending,
		SpecialNotes: request.SpecialNotes,
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
	reservation, err := u.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	return &model.ReservationResponse{
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
	}, nil
}

func (u *reservationUseCase) GetAll(params *model.ReservationQueryParams) (*model.PaginationResponse[[]model.ReservationResponse], error) {
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

	return &model.PaginationResponse[[]model.ReservationResponse]{
		Data:       responses,
		Total:      result.Total,
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalPages: result.TotalPages,
	}, nil
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
	return u.repo.Delete(id)
}
