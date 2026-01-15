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
	{
		Version:     1,
		Description: "初始化数据库结构",
		Up:          migrationV1,
	},
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

// migrationV1 创建初始表结构
func migrationV1(tx *sql.Tx) error {
	// 创建用户表
	if _, err := tx.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id              SERIAL PRIMARY KEY,
			username        VARCHAR(64) UNIQUE NOT NULL,
			password_hash   VARCHAR(256) NOT NULL,
			email           VARCHAR(128),
			phone           VARCHAR(32),
			real_name       VARCHAR(64),
			avatar          VARCHAR(256),
			last_login_at   TIMESTAMP,
			created_at      TIMESTAMP DEFAULT NOW(),
			updated_at      TIMESTAMP DEFAULT NOW()
		)
	`); err != nil {
		return err
	}

	// 创建用户表索引
	if _, err := tx.Exec(`CREATE INDEX IF NOT EXISTS idx_users_username ON users(username)`); err != nil {
		return err
	}

	// 创建推流表
	if _, err := tx.Exec(`
		CREATE TABLE IF NOT EXISTS streams (
			id                      SERIAL PRIMARY KEY,
			stream_key              VARCHAR(64) UNIQUE NOT NULL,
			name                    VARCHAR(128) NOT NULL,
			description             TEXT,
			device_id               VARCHAR(64),
			status                  VARCHAR(16) DEFAULT 'idle',
			visibility              VARCHAR(16) DEFAULT 'public',
			password                VARCHAR(256),
			record_enabled          BOOLEAN DEFAULT FALSE,
			record_files            JSONB DEFAULT '[]',
			protocol                VARCHAR(16),
			bitrate                 INTEGER DEFAULT 0,
			fps                     INTEGER DEFAULT 0,
			streamer_name           VARCHAR(64) NOT NULL,
			streamer_contact        VARCHAR(128),
			scheduled_start_time    TIMESTAMP NOT NULL,
			scheduled_end_time      TIMESTAMP NOT NULL,
			auto_kick_delay         INTEGER DEFAULT 30,
			actual_start_time       TIMESTAMP,
			actual_end_time         TIMESTAMP,
			last_frame_at           TIMESTAMP,
			current_viewers         INTEGER DEFAULT 0,
			total_viewers           INTEGER DEFAULT 0,
			peak_viewers            INTEGER DEFAULT 0,
			created_by              INTEGER REFERENCES users(id),
			created_at              TIMESTAMP DEFAULT NOW(),
			updated_at              TIMESTAMP DEFAULT NOW()
		)
	`); err != nil {
		return err
	}

	// 创建推流表索引
	indexes := []string{
		`CREATE INDEX IF NOT EXISTS idx_streams_status ON streams(status)`,
		`CREATE INDEX IF NOT EXISTS idx_streams_visibility ON streams(visibility)`,
		`CREATE INDEX IF NOT EXISTS idx_streams_device_id ON streams(device_id)`,
		`CREATE INDEX IF NOT EXISTS idx_streams_created_by ON streams(created_by)`,
		`CREATE INDEX IF NOT EXISTS idx_streams_scheduled_start ON streams(scheduled_start_time)`,
		`CREATE INDEX IF NOT EXISTS idx_streams_scheduled_end ON streams(scheduled_end_time)`,
		`CREATE INDEX IF NOT EXISTS idx_streams_record_enabled ON streams(record_enabled)`,
	}
	for _, idx := range indexes {
		if _, err := tx.Exec(idx); err != nil {
			return err
		}
	}

	// 创建操作日志表
	if _, err := tx.Exec(`
		CREATE TABLE IF NOT EXISTS operation_logs (
			id              SERIAL PRIMARY KEY,
			user_id         INTEGER REFERENCES users(id),
			action          VARCHAR(64) NOT NULL,
			target_type     VARCHAR(32),
			target_id       VARCHAR(64),
			detail          JSONB,
			created_at      TIMESTAMP DEFAULT NOW()
		)
	`); err != nil {
		return err
	}

	// 创建操作日志表索引
	logIndexes := []string{
		`CREATE INDEX IF NOT EXISTS idx_logs_user_id ON operation_logs(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_logs_created_at ON operation_logs(created_at)`,
		`CREATE INDEX IF NOT EXISTS idx_logs_action ON operation_logs(action)`,
	}
	for _, idx := range logIndexes {
		if _, err := tx.Exec(idx); err != nil {
			return err
		}
	}

	return nil
}
