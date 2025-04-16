package usecase

import (
	"cakestore/internal/entity"
	"cakestore/internal/repository"

	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
)

type CakeUseCase interface {
	GetAllCakes() ([]entity.Cake, error)
	GetCakeByID(id int) (*entity.Cake, error)
	CreateCake(cake *entity.Cake) error
	UpdateCake(cake *entity.Cake) error
	SoftDeleteCake(id int) error
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

func (uc *cakeUseCase) GetAllCakes() ([]entity.Cake, error) {
	cakes, err := uc.repo.GetAll()
	if err != nil {
		uc.logger.Errorf("Error fetching all cakes: %v", err)
		return nil, err
	}
	uc.logger.Info("Successfully fetched all cakes")
	return cakes, nil
}

func (uc *cakeUseCase) GetCakeByID(id int) (*entity.Cake, error) {
	cake, err := uc.repo.GetByID(id)
	if err != nil {
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
		uc.logger.Errorf("Error updating cake: %v", err)
		return err
	}
	uc.logger.Infof("Successfully updated cake with ID %d", cake.ID)
	return nil
}

func (uc *cakeUseCase) SoftDeleteCake(id int) error {
	if err := uc.repo.SoftDelete(id); err != nil {
		uc.logger.Errorf("Error deleting cake with ID %d: %v", id, err)
		return err
	}
	uc.logger.Infof("Successfully deleted cake with ID %d", id)
	return nil
}
