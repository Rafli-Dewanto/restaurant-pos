package repository

import (
	"cakestore/internal/constants"
	"cakestore/internal/domain/entity"
	"cakestore/internal/domain/model"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

type CakeRepository interface {
	GetAll(params *model.CakeQueryParams) (*model.PaginationResponse[[]entity.Cake], error)
	GetByID(id int64) (*entity.Cake, error)
	Create(cake *entity.Cake) error
	UpdateCake(cake *entity.Cake) error
	SoftDelete(id int64) error
}

type cakeRepository struct {
	db  *sqlx.DB
	log *logrus.Logger
}

func NewCakeRepository(db *sqlx.DB, log *logrus.Logger) CakeRepository {
	return &cakeRepository{db: db, log: log}
}

func (r *cakeRepository) GetAll(params *model.CakeQueryParams) (*model.PaginationResponse[[]entity.Cake], error) {
	var cakes []entity.Cake
	var total int64

	// Build dynamic query
	query := "SELECT * FROM cakes WHERE deleted_at IS NULL"
	countQuery := "SELECT COUNT(*) FROM cakes WHERE deleted_at IS NULL"

	var args []interface{}
	where := ""

	if params.Title != "" {
		where += " AND title ILIKE ?"
		args = append(args, "%"+params.Title+"%")
	}
	if params.MinPrice > 0 {
		where += " AND price >= ?"
		args = append(args, params.MinPrice)
	}
	if params.MaxPrice > 0 {
		where += " AND price <= ?"
		args = append(args, params.MaxPrice)
	}
	if params.Category != "" {
		where += " AND category = ?"
		args = append(args, params.Category)
	}

	query += where
	countQuery += where

	// Get total
	err := r.db.Get(&total, countQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to count cakes: %w", err)
	}

	// Pagination
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PageSize < 1 {
		if params.Limit == 0 {
			params.PageSize = 10
		} else {
			params.PageSize = params.Limit
		}
	}
	offset := (params.Page - 1) * params.PageSize

	// Final query with ordering and pagination
	query += " ORDER BY rating DESC, title ASC LIMIT ? OFFSET ?"
	args = append(args, params.PageSize, offset)

	err = r.db.Select(&cakes, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get cakes: %w", err)
	}

	totalPages := total / params.PageSize
	if total%params.PageSize != 0 {
		totalPages++
	}

	return &model.PaginationResponse[[]entity.Cake]{
		Data:       cakes,
		Total:      total,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: totalPages,
	}, nil
}

func (r *cakeRepository) GetByID(id int64) (*entity.Cake, error) {
	var cake entity.Cake
	query := "SELECT * FROM cakes WHERE id = ? AND deleted_at IS NULL"
	err := r.db.Get(&cake, query, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, constants.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get cake by id: %w", err)
	}
	return &cake, nil
}

func (r *cakeRepository) Create(cake *entity.Cake) error {
	query := `
		INSERT INTO cakes (title, description, rating, image, price, category, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, NOW(), NOW())
		RETURNING id
	`
	err := r.db.QueryRow(query,
		cake.Title,
		cake.Description,
		cake.Rating,
		cake.Image,
		cake.Price,
		cake.Category,
	).Scan(&cake.ID)
	if err != nil {
		return fmt.Errorf("failed to create cake: %w", err)
	}
	return nil
}

func (r *cakeRepository) UpdateCake(cake *entity.Cake) error {
	query := `
		UPDATE cakes
		SET title = ?, description = ?, rating = ?, image = ?, updated_at = NOW()
		WHERE id = ? AND deleted_at IS NULL
	`
	result, err := r.db.Exec(query,
		cake.Title,
		cake.Description,
		cake.Rating,
		cake.Image,
		cake.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update cake: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return constants.ErrNotFound
	}

	return nil
}

func (r *cakeRepository) SoftDelete(id int64) error {
	query := `
		UPDATE cakes
		SET deleted_at = NOW()
		WHERE id = ? AND deleted_at IS NULL
	`
	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to soft delete cake: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New("no rows updated, cake not found")
	}

	return nil
}
