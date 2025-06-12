package usecase

import (
	"cakestore/internal/domain/entity"
	"cakestore/internal/domain/model"
	"cakestore/internal/repository"
	"cakestore/utils"
	"time"

	"github.com/sirupsen/logrus"
)

type TableUseCase interface {
	Create(request *model.CreateTableRequest) (*model.TableResponse, error)
	GetByID(id uint) (*model.TableResponse, error)
	GetAll(params *model.TableQueryParams) (*model.PaginationResponse[[]model.TableResponse], error)
	Update(id uint, request *model.UpdateTableRequest) (*model.TableResponse, error)
	Delete(id uint) error
	GetAvailableTables(reserveTime time.Time, duration time.Duration) ([]model.TableResponse, error)
	UpdateAvailability(id uint, isAvailable bool) error
}

type tableUseCase struct {
	tableRepo repository.TableRepository
	log       *logrus.Logger
}

func NewTableUseCase(tableRepo repository.TableRepository, log *logrus.Logger) TableUseCase {
	return &tableUseCase{
		tableRepo: tableRepo,
		log:       log,
	}
}

func (u *tableUseCase) Create(request *model.CreateTableRequest) (*model.TableResponse, error) {
	table := &entity.Table{
		TableNumber: request.TableNumber,
		Capacity:    request.Capacity,
		IsAvailable: true,
	}

	if err := u.tableRepo.Create(table); err != nil {
		return nil, err
	}

	return model.ToTableResponse(table), nil
}

func (u *tableUseCase) GetByID(id uint) (*model.TableResponse, error) {
	table, err := u.tableRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	return model.ToTableResponse(table), nil
}

func (u *tableUseCase) GetAll(params *model.TableQueryParams) (*model.PaginationResponse[[]model.TableResponse], error) {
	tables, err := u.tableRepo.GetAll()
	if err != nil {
		return nil, err
	}

	var filteredTables []entity.Table
	for _, table := range tables {
		if params.Capacity > 0 && table.Capacity != params.Capacity {
			continue
		}
		if params.IsAvailable != nil && table.IsAvailable != *params.IsAvailable {
			continue
		}
		filteredTables = append(filteredTables, table)
	}

	paginatedTables := utils.CreatePaginationMeta(params.Page, params.Limit, int64(len(filteredTables)))

	var tableResponses []model.TableResponse
	for _, table := range filteredTables {
		tableResponses = append(tableResponses, *model.ToTableResponse(&table))
	}

	return &model.PaginationResponse[[]model.TableResponse]{
		Data:       tableResponses,
		Total:      int64(len(filteredTables)),
		Page:       params.Page,
		PageSize:   params.Limit,
		TotalPages: paginatedTables.LastPage,
	}, nil
}

func (u *tableUseCase) Update(id uint, request *model.UpdateTableRequest) (*model.TableResponse, error) {
	table, err := u.tableRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	table.TableNumber = request.TableNumber
	table.Capacity = request.Capacity
	table.IsAvailable = request.IsAvailable

	if err := u.tableRepo.Update(table); err != nil {
		return nil, err
	}

	return model.ToTableResponse(table), nil
}

func (u *tableUseCase) Delete(id uint) error {
	return u.tableRepo.Delete(id)
}

func (u *tableUseCase) GetAvailableTables(reserveTime time.Time, duration time.Duration) ([]model.TableResponse, error) {
	tables, err := u.tableRepo.GetAvailableTables(reserveTime, duration)
	if err != nil {
		return nil, err
	}

	var tableResponses []model.TableResponse
	for _, table := range tables {
		tableResponses = append(tableResponses, *model.ToTableResponse(&table))
	}

	return tableResponses, nil
}

func (u *tableUseCase) UpdateAvailability(id uint, isAvailable bool) error {
	return u.tableRepo.UpdateAvailability(id, isAvailable)
}
