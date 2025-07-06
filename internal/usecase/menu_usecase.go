package usecase

import (
	"cakestore/internal/constants"
	"cakestore/internal/database"
	"cakestore/internal/domain/entity"
	"cakestore/internal/domain/model"
	"cakestore/internal/repository"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
)

type MenuUseCase interface {
	GetAllMenus(params *model.MenuQueryParams) (*model.PaginationResponse[[]entity.Menu], error)
	GetMenuByID(id int64) (*entity.Menu, error)
	CreateMenu(menu *entity.Menu) error
	UpdateMenu(menu *entity.Menu) error
	SoftDeleteMenu(id int64) error
}

type menuUseCase struct {
	repo     repository.MenuRepository
	logger   *logrus.Logger
	validate *validator.Validate
	cache    database.RedisCache
}

func NewMenuUseCase(repo repository.MenuRepository, logger *logrus.Logger, cache database.RedisCache) MenuUseCase {
	return &menuUseCase{
		repo:     repo,
		logger:   logger,
		validate: validator.New(),
		cache:    cache,
	}
}

func (uc *menuUseCase) GetAllMenus(params *model.MenuQueryParams) (*model.PaginationResponse[[]entity.Menu], error) {
	start := time.Now()
	defer func() {
		uc.logger.Infof("GetAllMenus took %v", time.Since(start))
	}()

	if params == nil {
		params = &model.MenuQueryParams{}
	}

	// Try to get the menus from the cache first
	cacheKey := fmt.Sprintf("menus:all:page:%d:limit:%d", params.Page, params.Limit)
	var cachedData model.PaginationResponse[[]entity.Menu]
	if err := uc.cache.Get(context.Background(), cacheKey, &cachedData); err == nil {
		uc.logger.Info("Menus fetched from cache")
		return &cachedData, nil
	}

	// If not in cache, get from the database
	response, err := uc.repo.GetAll(params)
	if err != nil {
		uc.logger.Errorf("Error fetching menus with params: %v, error: %v", params, err)
		return nil, err
	}

	// Store the menus in the cache for future requests
	if err := uc.cache.Set(context.Background(), cacheKey, response, 5*time.Minute); err != nil {
		uc.logger.Errorf("Error setting cache for all menus: %v", err)
	}

	return response, nil
}

func (uc *menuUseCase) GetMenuByID(id int64) (*entity.Menu, error) {
	start := time.Now()
	defer func() {
		uc.logger.Infof("GetMenuByID took %v", time.Since(start))
	}()

	// Try to get the menu from the cache first
	cacheKey := fmt.Sprintf("menu:%d", id)
	var menu entity.Menu
	if err := uc.cache.Get(context.Background(), cacheKey, &menu); err == nil {
		uc.logger.Info("Menu fetched from cache")
		return &menu, nil
	}

	// If not in cache, get from the database
	menuEntity, err := uc.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return nil, constants.ErrNotFound
		}
		uc.logger.Errorf("Error fetching menu with ID %d: %v", id, err)
		return nil, err
	}

	// Store the menu in the cache for future requests
	if err := uc.cache.Set(context.Background(), cacheKey, menuEntity, 5*time.Minute); err != nil {
		uc.logger.Errorf("Error setting cache for menu ID %d: %v", id, err)
	}

	uc.logger.Infof("Successfully fetched menu with ID %d", id)
	return menuEntity, nil
}

func (uc *menuUseCase) CreateMenu(menu *entity.Menu) error {
	if err := uc.validate.Struct(menu); err != nil {
		uc.logger.Errorf("Validation failed for menu: %v", err)
		return err
	}

	if err := uc.repo.Create(menu); err != nil {
		uc.logger.Errorf("Error creating menu: %v", err)
		return err
	}
	uc.logger.Infof("Successfully created a new menu: %s", menu.Title)
	return nil
}

func (uc *menuUseCase) UpdateMenu(menu *entity.Menu) error {
	if err := uc.validate.Struct(menu); err != nil {
		uc.logger.Errorf("Validation failed for menu: %v", err)
		return err
	}

	if err := uc.repo.UpdateMenu(menu); err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return constants.ErrNotFound
		}
		uc.logger.Errorf("Error updating menu: %v", err)
		return err
	}

	// Invalidate cache
	cacheKey := fmt.Sprintf("menu:%d", menu.ID)
	if err := uc.cache.Delete(context.Background(), cacheKey); err != nil {
		uc.logger.Errorf("Error deleting cache for menu ID %d: %v", menu.ID, err)
	}
	if err := uc.cache.Delete(context.Background(), "menus:all:*"); err != nil {
		uc.logger.Errorf("Error deleting cache for all menus: %v", err)
	}

	uc.logger.Infof("Successfully updated menu with ID %d", menu.ID)
	return nil
}

func (uc *menuUseCase) SoftDeleteMenu(id int64) error {
	if err := uc.repo.SoftDelete(id); err != nil {
		uc.logger.Errorf("Error deleting menu with ID %d: %v", id, err)
		return err
	}

	// Invalidate cache
	cacheKey := fmt.Sprintf("menu:%d", id)
	if err := uc.cache.Delete(context.Background(), cacheKey); err != nil {
		uc.logger.Errorf("Error deleting cache for menu ID %d: %v", id, err)
	}
	if err := uc.cache.Delete(context.Background(), "menus:all:*"); err != nil {
		uc.logger.Errorf("Error deleting cache for all menus: %v", err)
	}

	uc.logger.Infof("Successfully deleted menu with ID %d", id)
	return nil
}
