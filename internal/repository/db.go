package repository

import (
	"database/sql"
	"embed"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"

	"easy-stream/internal/config"

	_ "github.com/lib/pq"
)

// 当前数据库最新版本
const LatestDBVersion = 5

//go:embed migrations/*.sql
var migrationsFS embed.FS

//go:embed init-db.sql
var initDBSQL string

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
	// 检查是否是空数据库（schema_migrations 表不存在）
	isEmpty, err := isDatabaseEmpty(db)
	if err != nil {
		return err
	}

	if isEmpty {
		// 空数据库：直接执行 init-db.sql 初始化到最新版本
		log.Printf("检测到空数据库，执行初始化脚本...")
		if err := initDatabase(db); err != nil {
			return err
		}
		log.Printf("数据库初始化完成，版本: %d", LatestDBVersion)
		return nil
	}

	// 已有数据：获取当前版本，依次执行迁移脚本
	currentVersion, err := getCurrentVersion(db)
	if err != nil {
		return err
	}

	log.Printf("当前数据库版本: %d, 目标版本: %d", currentVersion, LatestDBVersion)

	if currentVersion >= LatestDBVersion {
		log.Printf("数据库已是最新版本")
		return nil
	}

	// 依次执行迁移脚本
	for v := currentVersion + 1; v <= LatestDBVersion; v++ {
		if err := executeMigrationFile(db, v); err != nil {
			return fmt.Errorf("迁移 v%d 失败: %w", v, err)
		}
		log.Printf("迁移 v%d 完成", v)
	}

	return nil
}

// isDatabaseEmpty 检查数据库是否为空（schema_migrations 表不存在）
func isDatabaseEmpty(db *sql.DB) (bool, error) {
	var exists bool
	query := `
		SELECT EXISTS (
			SELECT FROM information_schema.tables
			WHERE table_schema = 'public'
			AND table_name = 'schema_migrations'
		)
	`
	if err := db.QueryRow(query).Scan(&exists); err != nil {
		return false, err
	}
	return !exists, nil
}

// initDatabase 执行 init-db.sql 初始化数据库
func initDatabase(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 执行初始化脚本
	if _, err := tx.Exec(initDBSQL); err != nil {
		return fmt.Errorf("执行 init-db.sql 失败: %w", err)
	}

	// 记录版本号
	_, err = tx.Exec(
		"INSERT INTO schema_migrations (version, description) VALUES ($1, $2) ON CONFLICT (version) DO NOTHING",
		LatestDBVersion, "初始化数据库到最新版本",
	)
	if err != nil {
		return err
	}

	return tx.Commit()
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

// executeMigrationFile 执行指定版本的迁移文件
func executeMigrationFile(db *sql.DB, version int) error {
	// 查找对应版本的迁移文件
	filename, err := findMigrationFile(version)
	if err != nil {
		return err
	}

	log.Printf("执行迁移文件: %s", filename)

	// 读取迁移文件内容
	content, err := migrationsFS.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("读取迁移文件失败: %w", err)
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 执行迁移脚本
	if _, err := tx.Exec(string(content)); err != nil {
		return fmt.Errorf("执行迁移脚本失败: %w", err)
	}

	// 从文件名提取描述
	description := extractDescription(filename)

	// 记录迁移版本
	_, err = tx.Exec(
		"INSERT INTO schema_migrations (version, description) VALUES ($1, $2)",
		version, description,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// findMigrationFile 查找指定版本的迁移文件
func findMigrationFile(version int) (string, error) {
	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return "", err
	}

	prefix := fmt.Sprintf("%03d_", version)
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), prefix) && strings.HasSuffix(entry.Name(), ".sql") {
			return "migrations/" + entry.Name(), nil
		}
	}

	return "", fmt.Errorf("未找到版本 %d 的迁移文件", version)
}

// extractDescription 从文件名提取描述
func extractDescription(filename string) string {
	// migrations/004_add_share_feature.sql -> add_share_feature
	base := strings.TrimPrefix(filename, "migrations/")
	base = strings.TrimSuffix(base, ".sql")
	parts := strings.SplitN(base, "_", 2)
	if len(parts) == 2 {
		return strings.ReplaceAll(parts[1], "_", " ")
	}
	return base
}

// GetMigrationVersions 获取所有可用的迁移版本（用于调试）
func GetMigrationVersions() ([]int, error) {
	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return nil, err
	}

	var versions []int
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".sql") {
			parts := strings.SplitN(entry.Name(), "_", 2)
			if len(parts) >= 1 {
				if v, err := strconv.Atoi(parts[0]); err == nil {
					versions = append(versions, v)
				}
			}
		}
	}

	sort.Ints(versions)
	return versions, nil
}
