package usecase

import (
	"cakestore/internal/constants"
	"cakestore/internal/domain/entity"
	"cakestore/internal/domain/model"
	"cakestore/internal/repository"
	"errors"
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
}

func NewMenuUseCase(repo repository.MenuRepository, logger *logrus.Logger) MenuUseCase {
	return &menuUseCase{
		repo:     repo,
		logger:   logger,
		validate: validator.New(),
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

	response, err := uc.repo.GetAll(params)
	if err != nil {
		uc.logger.Errorf("Error fetching menus with params: %v, error: %v", params, err)
		return nil, err
	}

	return response, nil
}

func (uc *menuUseCase) GetMenuByID(id int64) (*entity.Menu, error) {
	start := time.Now()
	defer func() {
		uc.logger.Infof("GetMenuByID took %v", time.Since(start))
	}()

	menu, err := uc.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return nil, constants.ErrNotFound
		}
		uc.logger.Errorf("Error fetching menu with ID %d: %v", id, err)
		return nil, err
	}
	uc.logger.Infof("Successfully fetched menu with ID %d", id)
	return menu, nil
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
	uc.logger.Infof("Successfully updated menu with ID %d", menu.ID)
	return nil
}

func (uc *menuUseCase) SoftDeleteMenu(id int64) error {
	if err := uc.repo.SoftDelete(id); err != nil {
		uc.logger.Errorf("Error deleting menu with ID %d: %v", id, err)
		return err
	}
	uc.logger.Infof("Successfully deleted menu with ID %d", id)
	return nil
}
