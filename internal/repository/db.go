package repository

import (
	"database/sql"
	"fmt"

	"easy-stream/internal/config"

	_ "github.com/lib/pq"
)

// NewPostgresDB 创建 PostgreSQL 连接
func NewPostgresDB(cfg config.DatabaseConfig) (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
