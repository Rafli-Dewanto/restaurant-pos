package seeder

import (
	"cakestore/internal/domain/entity"
	"cakestore/internal/repository"
	"time"

	"github.com/sirupsen/logrus"
)

type TableSeeder struct {
	repo   repository.TableRepository
	logger *logrus.Logger
}

func NewTableSeeder(repo repository.TableRepository, logger *logrus.Logger) *TableSeeder {
	return &TableSeeder{
		repo:   repo,
		logger: logger,
	}
}

func (s *TableSeeder) Seed() error {
	// if seeded, return nil
	num, err := s.repo.Count()
	if err != nil {
		s.logger.Errorf("Failed to count tables: %v", err)
		return err
	}
	if num > 0 {
		return nil
	}

	var tables []entity.Table
	now := time.Now()

	for i := 1; i <= 20; i++ {
		table := entity.Table{
			TableNumber: i,
			Capacity:    2 + (i % 5 * 2), // capacity: 2, 4, 6, 8, 10 in a pattern
			IsAvailable: true,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		tables = append(tables, table)
	}

	for _, table := range tables {
		if err := s.repo.Create(&table); err != nil {
			s.logger.Errorf("Failed to seed table %d: %v", table.TableNumber, err)
			return err
		}
		s.logger.Infof("Seeded table %d", table.TableNumber)
	}

	return nil
}
