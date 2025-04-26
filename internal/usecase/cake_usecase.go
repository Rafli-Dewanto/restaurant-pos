package usecase

import (
	"cakestore/internal/constants"
	"cakestore/internal/domain/entity"
	"cakestore/internal/domain/model"
	"cakestore/internal/repository"
	"errors"

	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
)

type CakeUseCase interface {
	GetAllCakes(params *model.CakeQueryParams) (*model.PaginationResponse[[]entity.Cake], error)
	GetCakeByID(id int64) (*entity.Cake, error)
	CreateCake(cake *entity.Cake) error
	UpdateCake(cake *entity.Cake) error
	SoftDeleteCake(id int64) error
}

type cakeUseCase struct {
	repo     repository.CakeRepository
	logger   *logrus.Logger
	validate *validator.Validate
}

func NewCakeUseCase(repo repository.CakeRepository, logger *logrus.Logger) CakeUseCase {
	return &cakeUseCase{
		repo:     repo,
		logger:   logger,
		validate: validator.New(),
	}
}

func (uc *cakeUseCase) GetAllCakes(params *model.CakeQueryParams) (*model.PaginationResponse[[]entity.Cake], error) {
	if params == nil {
		params = &model.CakeQueryParams{}
	}

	response, err := uc.repo.GetAll(params)
	if err != nil {
		uc.logger.Errorf("Error fetching cakes with params: %v, error: %v", params, err)
		return nil, err
	}

	return response, nil
}

func (uc *cakeUseCase) GetCakeByID(id int64) (*entity.Cake, error) {
	cake, err := uc.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return nil, constants.ErrNotFound
		}
		uc.logger.Errorf("Error fetching cake with ID %d: %v", id, err)
		return nil, err
	}
	uc.logger.Infof("Successfully fetched cake with ID %d", id)
	return cake, nil
}

func (uc *cakeUseCase) CreateCake(cake *entity.Cake) error {
	if err := uc.validate.Struct(cake); err != nil {
		uc.logger.Errorf("Validation failed for cake: %v", err)
		return err
	}

	if err := uc.repo.Create(cake); err != nil {
		uc.logger.Errorf("Error creating cake: %v", err)
		return err
	}
	uc.logger.Infof("Successfully created a new cake: %s", cake.Title)
	return nil
}

func (uc *cakeUseCase) UpdateCake(cake *entity.Cake) error {
	if err := uc.validate.Struct(cake); err != nil {
		uc.logger.Errorf("Validation failed for cake: %v", err)
		return err
	}

	if err := uc.repo.UpdateCake(cake); err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return constants.ErrNotFound
		}
		uc.logger.Errorf("Error updating cake: %v", err)
		return err
	}
	uc.logger.Infof("Successfully updated cake with ID %d", cake.ID)
	return nil
}

func (uc *cakeUseCase) SoftDeleteCake(id int64) error {
	if err := uc.repo.SoftDelete(id); err != nil {
		uc.logger.Errorf("Error deleting cake with ID %d: %v", id, err)
		return err
	}
	uc.logger.Infof("Successfully deleted cake with ID %d", id)
	return nil
}
