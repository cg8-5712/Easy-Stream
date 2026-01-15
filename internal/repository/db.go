package repository

import (
	"database/sql"
	"fmt"
	"log"

	"easy-stream/internal/config"

	_ "github.com/lib/pq"
)

// 当前数据库版本
const CurrentDBVersion = 1

// Migration 表示一个数据库迁移
type Migration struct {
	Version     int
	Description string
	Up          func(*sql.Tx) error
}

// 定义所有迁移
var migrations = []Migration{
	// 版本1是初始版本，由 init-db.sql 创建，无需迁移
}

// NewPostgresDB 创建 PostgreSQL 连接并执行迁移
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

	// 执行数据库迁移
	if err := runMigrations(db); err != nil {
		return nil, fmt.Errorf("数据库迁移失败: %w", err)
	}

	return db, nil
}

// runMigrations 检查并执行数据库迁移
func runMigrations(db *sql.DB) error {
	// 确保迁移表存在
	if err := ensureMigrationTable(db); err != nil {
		return err
	}

	// 获取当前版本
	currentVersion, err := getCurrentVersion(db)
	if err != nil {
		return err
	}

	log.Printf("当前数据库版本: %d, 目标版本: %d", currentVersion, CurrentDBVersion)

	// 执行需要的迁移
	for _, m := range migrations {
		if m.Version > currentVersion {
			log.Printf("执行迁移 v%d: %s", m.Version, m.Description)
			if err := executeMigration(db, m); err != nil {
				return fmt.Errorf("迁移 v%d 失败: %w", m.Version, err)
			}
			log.Printf("迁移 v%d 完成", m.Version)
		}
	}

	return nil
}

// ensureMigrationTable 确保迁移表存在
func ensureMigrationTable(db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version     INTEGER PRIMARY KEY,
			description VARCHAR(256) NOT NULL,
			applied_at  TIMESTAMP DEFAULT NOW()
		)`
	_, err := db.Exec(query)
	return err
}

// getCurrentVersion 获取当前数据库版本
func getCurrentVersion(db *sql.DB) (int, error) {
	var version int
	err := db.QueryRow("SELECT COALESCE(MAX(version), 0) FROM schema_migrations").Scan(&version)
	if err != nil {
		return 0, err
	}
	return version, nil
}

// executeMigration 执行单个迁移
func executeMigration(db *sql.DB, m Migration) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 执行迁移
	if err := m.Up(tx); err != nil {
		return err
	}

	// 记录迁移版本
	_, err = tx.Exec(
		"INSERT INTO schema_migrations (version, description) VALUES ($1, $2)",
		m.Version, m.Description,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}
