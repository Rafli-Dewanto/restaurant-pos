package usecase

import (
	"cakestore/internal/database"
	"cakestore/internal/domain/entity"
	"cakestore/internal/domain/model"
	"cakestore/internal/repository"
	"cakestore/utils"
	"context"
	"fmt"
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
	cache     database.RedisCache
}

func NewTableUseCase(tableRepo repository.TableRepository, log *logrus.Logger, cache database.RedisCache) TableUseCase {
	return &tableUseCase{
		tableRepo: tableRepo,
		log:       log,
		cache:     cache,
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
	start := time.Now()
	defer func() {
		u.log.Infof("GetByID took %v", time.Since(start))
	}()

	// Try to get the table from the cache first
	cacheKey := fmt.Sprintf("table:%d", id)
	var table model.TableResponse
	if err := u.cache.Get(context.Background(), cacheKey, &table); err == nil {
		u.log.Info("Table fetched from cache")
		return &table, nil
	}

	// If not in cache, get from the database
	tableEntity, err := u.tableRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	response := model.ToTableResponse(tableEntity)

	// Store the table in the cache for future requests
	if err := u.cache.Set(context.Background(), cacheKey, response, 5*time.Minute); err != nil {
		u.log.Errorf("Error setting cache for table ID %d: %v", id, err)
	}

	return response, nil
}

func (u *tableUseCase) GetAll(params *model.TableQueryParams) (*model.PaginationResponse[[]model.TableResponse], error) {
	start := time.Now()
	defer func() {
		u.log.Infof("GetAll took %v", time.Since(start))
	}()

	// Try to get the tables from the cache first
	cacheKey := fmt.Sprintf("tables:all:page:%d:limit:%d", params.Page, params.Limit)
	var cachedData model.PaginationResponse[[]model.TableResponse]
	if err := u.cache.Get(context.Background(), cacheKey, &cachedData); err == nil {
		u.log.Info("Tables fetched from cache")
		return &cachedData, nil
	}

	// If not in cache, get from the database
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

	paginatedResponse := &model.PaginationResponse[[]model.TableResponse]{
		Data:       tableResponses,
		Total:      int64(len(filteredTables)),
		Page:       params.Page,
		PageSize:   params.Limit,
		TotalPages: paginatedTables.LastPage,
	}

	// Store the tables in the cache for future requests
	if err := u.cache.Set(context.Background(), cacheKey, paginatedResponse, 5*time.Minute); err != nil {
		u.log.Errorf("Error setting cache for all tables: %v", err)
	}

	return paginatedResponse, nil
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

	// Invalidate cache
	cacheKey := fmt.Sprintf("table:%d", id)
	if err := u.cache.Delete(context.Background(), cacheKey); err != nil {
		u.log.Errorf("Error deleting cache for table ID %d: %v", id, err)
	}
	if err := u.cache.Delete(context.Background(), "tables:all:*"); err != nil {
		u.log.Errorf("Error deleting cache for all tables: %v", err)
	}
	if err := u.cache.Delete(context.Background(), "available_tables:*"); err != nil {
		u.log.Errorf("Error deleting cache for available tables: %v", err)
	}

	return model.ToTableResponse(table), nil
}

func (u *tableUseCase) Delete(id uint) error {
	if err := u.tableRepo.Delete(id); err != nil {
		return err
	}

	// Invalidate cache
	cacheKey := fmt.Sprintf("table:%d", id)
	if err := u.cache.Delete(context.Background(), cacheKey); err != nil {
		u.log.Errorf("Error deleting cache for table ID %d: %v", id, err)
	}
	if err := u.cache.Delete(context.Background(), "tables:all:*"); err != nil {
		u.log.Errorf("Error deleting cache for all tables: %v", err)
	}
	if err := u.cache.Delete(context.Background(), "available_tables:*"); err != nil {
		u.log.Errorf("Error deleting cache for available tables: %v", err)
	}

	return nil
}

func (u *tableUseCase) GetAvailableTables(reserveTime time.Time, duration time.Duration) ([]model.TableResponse, error) {
	start := time.Now()
	defer func() {
		u.log.Infof("GetAvailableTables took %v", time.Since(start))
	}()

	// Try to get the available tables from the cache first
	cacheKey := fmt.Sprintf("available_tables:%s:%s", reserveTime.Format(time.RFC3339), duration.String())
	var tables []model.TableResponse
	if err := u.cache.Get(context.Background(), cacheKey, &tables); err == nil {
		u.log.Info("Available tables fetched from cache")
		return tables, nil
	}

	// If not in cache, get from the database
	tableEntities, err := u.tableRepo.GetAvailableTables(reserveTime, duration)
	if err != nil {
		return nil, err
	}

	var tableResponses []model.TableResponse
	for _, table := range tableEntities {
		tableResponses = append(tableResponses, *model.ToTableResponse(&table))
	}

	// Store the available tables in the cache for future requests
	if err := u.cache.Set(context.Background(), cacheKey, tableResponses, 5*time.Minute); err != nil {
		u.log.Errorf("Error setting cache for available tables: %v", err)
	}

	return tableResponses, nil
}

func (u *tableUseCase) UpdateAvailability(id uint, isAvailable bool) error {
	if err := u.tableRepo.UpdateAvailability(id, isAvailable); err != nil {
		return err
	}

	// Invalidate cache
	cacheKey := fmt.Sprintf("table:%d", id)
	if err := u.cache.Delete(context.Background(), cacheKey); err != nil {
		u.log.Errorf("Error deleting cache for table ID %d: %v", id, err)
	}
	if err := u.cache.Delete(context.Background(), "tables:all:*"); err != nil {
		u.log.Errorf("Error deleting cache for all tables: %v", err)
	}
	if err := u.cache.Delete(context.Background(), "available_tables:*"); err != nil {
		u.log.Errorf("Error deleting cache for available tables: %v", err)
	}

	return nil
}
