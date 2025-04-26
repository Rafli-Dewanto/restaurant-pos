package database

import (
	configs "cakestore/internal/config"
	"fmt"

	"github.com/jmoiron/sqlx"
)

func NewDB(dbConfig *configs.Config) (*sqlx.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%v)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbConfig.DBUser, dbConfig.DBPassword, dbConfig.DBHost, dbConfig.DBPort, dbConfig.DBName)

	db, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("cannot connect to db: %w", err)
	}

	return db, nil
}
