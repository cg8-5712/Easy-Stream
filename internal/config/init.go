package config

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// InitConfig creates a default config.yaml file
func InitConfig() error {
	// Check if config.yaml already exists
	if _, err := os.Stat("config.yaml"); err == nil {
		return fmt.Errorf("config.yaml already exists")
	}

	defaultConfig := `# Easy-Stream Configuration File

server:
  host: "0.0.0.0"
  port: "8080"
  mode: "debug"  # debug / release

database:
  type: "sqlite"  # postgres / mysql / sqlite
  # SQLite 配置（默认）
  filepath: "./easy_stream.db"
  # PostgreSQL / MySQL 配置（可选）
  # host: "localhost"
  # port: "5432"  # PostgreSQL: 5432, MySQL: 3306
  # user: "postgres"
  # password: "your_password"
  # dbname: "easy_stream"
  # sslmode: "disable"  # PostgreSQL only

redis:
  host: "localhost"
  port: "6379"
  password: ""
  db: 0

jwt:
  secret: "your-secret-key-change-this"

# 管理员账号配置（可选，留空则自动生成随机密码）
admin:
  username: "admin"
  password: ""  # 留空则在首次初始化时自动生成

zlmediakit:
  host: "localhost"
  port: "80"
  secret: "035c73f7-bb6b-4889-a715-d9eb2d1925cc"
  hookbaseurl: "http://localhost:8080/api/v1/hooks"
  recordmode: "local"  # local / remote
  recordpath: "./records"

log:
  level: "info"  # debug / info / warn / error

storage:
  targets: []
  # Example S3 configuration:
  # targets:
  #   - type: "s3"
  #     bucket: "your-bucket"
  #     region: "us-east-1"
  #     accesskey: "your-access-key"
  #     secretkey: "your-secret-key"
  #     endpoint: ""  # Optional, for S3-compatible services
`

	return os.WriteFile("config.yaml", []byte(defaultConfig), 0644)
}

// SaveConfig saves configuration to config.yaml file
func SaveConfig(cfg *Config) error {
	configContent := fmt.Sprintf(`# Easy-Stream Configuration File

server:
  host: "%s"
  port: "%s"
  mode: "%s"

database:
  type: "%s"
  filepath: "%s"
  host: "%s"
  port: "%s"
  user: "%s"
  password: "%s"
  dbname: "%s"
  sslmode: "%s"

redis:
  host: "%s"
  port: "%s"
  password: "%s"
  db: %d

jwt:
  secret: "%s"

admin:
  username: "%s"
  password: ""

zlmediakit:
  host: "%s"
  port: "%s"
  secret: "%s"
  hookbaseurl: "%s"
  httpport: "%s"
  httpsport: "%s"
  webrtcport: "%s"
  recordmode: "%s"
  recordlocalpath: "%s"
  recordbaseurl: "%s"
  externalhost: "%s"

log:
  level: "%s"

storage:
  targets: []
`,
		cfg.Server.Host, cfg.Server.Port, cfg.Server.Mode,
		cfg.Database.Type, cfg.Database.FilePath, cfg.Database.Host, cfg.Database.Port,
		cfg.Database.User, cfg.Database.Password, cfg.Database.DBName, cfg.Database.SSLMode,
		cfg.Redis.Host, cfg.Redis.Port, cfg.Redis.Password, cfg.Redis.DB,
		cfg.JWT.Secret,
		cfg.Admin.Username,
		cfg.ZLMediaKit.Host, cfg.ZLMediaKit.Port, cfg.ZLMediaKit.Secret, cfg.ZLMediaKit.HookBaseURL,
		cfg.ZLMediaKit.HTTPPort, cfg.ZLMediaKit.HTTPSPort, cfg.ZLMediaKit.WebRTCPort,
		cfg.ZLMediaKit.RecordMode, cfg.ZLMediaKit.RecordLocalPath, cfg.ZLMediaKit.RecordBaseURL,
		cfg.ZLMediaKit.ExternalHost,
		cfg.Log.Level,
	)

	return os.WriteFile("config.yaml", []byte(configContent), 0644)
}

// Validate validates the configuration
func Validate(cfg *Config) error {
	// Validate Server
	if cfg.Server.Port == "" {
		return fmt.Errorf("server.port is required")
	}
	if cfg.Server.Mode != "debug" && cfg.Server.Mode != "release" {
		return fmt.Errorf("server.mode must be 'debug' or 'release'")
	}

	// Validate Database
	if cfg.Database.Type == "" {
		return fmt.Errorf("database.type is required")
	}
	validDBTypes := map[string]bool{"postgres": true, "mysql": true, "sqlite": true}
	if !validDBTypes[cfg.Database.Type] {
		return fmt.Errorf("database.type must be one of: postgres, mysql, sqlite")
	}

	// SQLite 只需要 filepath
	if cfg.Database.Type == "sqlite" {
		if cfg.Database.FilePath == "" {
			return fmt.Errorf("database.filepath is required for sqlite")
		}
	} else {
		// PostgreSQL 和 MySQL 需要连接信息
		if cfg.Database.Host == "" {
			return fmt.Errorf("database.host is required")
		}
		if cfg.Database.Port == "" {
			return fmt.Errorf("database.port is required")
		}
		if cfg.Database.User == "" {
			return fmt.Errorf("database.user is required")
		}
		if cfg.Database.DBName == "" {
			return fmt.Errorf("database.dbname is required")
		}
	}

	// Validate Redis
	if cfg.Redis.Host == "" {
		return fmt.Errorf("redis.host is required")
	}
	if cfg.Redis.Port == "" {
		return fmt.Errorf("redis.port is required")
	}

	// Validate JWT
	if cfg.JWT.Secret == "" {
		return fmt.Errorf("jwt.secret is required")
	}
	if cfg.JWT.Secret == "your-secret-key-change-this" {
		return fmt.Errorf("jwt.secret must be changed from default value")
	}

	// Validate ZLMediaKit
	if cfg.ZLMediaKit.Host == "" {
		return fmt.Errorf("zlmediakit.host is required")
	}
	if cfg.ZLMediaKit.Port == "" {
		return fmt.Errorf("zlmediakit.port is required")
	}
	if cfg.ZLMediaKit.RecordMode != "local" && cfg.ZLMediaKit.RecordMode != "remote" {
		return fmt.Errorf("zlmediakit.recordmode must be 'local' or 'remote'")
	}

	// Validate Log
	validLogLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
	if !validLogLevels[cfg.Log.Level] {
		return fmt.Errorf("log.level must be one of: debug, info, warn, error")
	}

	return nil
}

// ServiceStatus represents the status of a service
type ServiceStatus struct {
	Name    string
	Status  string // "✓ Connected" or "✗ Failed"
	Error   string
	Details string
}

// VerifyConnections tests connections to all configured services
func VerifyConnections(cfg *Config) []ServiceStatus {
	var statuses []ServiceStatus

	// Test PostgreSQL connection
	dbStatus := testDatabaseConnection(cfg.Database)
	statuses = append(statuses, dbStatus)

	// Test Redis connection
	redisStatus := testRedisConnection(cfg.Redis)
	statuses = append(statuses, redisStatus)

	// Test ZLMediaKit connection
	zlmStatus := testZLMediaKitConnection(cfg.ZLMediaKit)
	statuses = append(statuses, zlmStatus)

	return statuses
}

// testDatabaseConnection tests database connection
func testDatabaseConnection(cfg DatabaseConfig) ServiceStatus {
	status := ServiceStatus{
		Name: "Database",
	}

	var dialector gorm.Dialector

	// 根据数据库类型选择驱动
	switch cfg.Type {
	case "postgres":
		dsn := fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
		)
		dialector = postgres.Open(dsn)
		status.Name = "PostgreSQL"

	case "mysql":
		dsn := fmt.Sprintf(
			"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName,
		)
		dialector = mysql.Open(dsn)
		status.Name = "MySQL"

	case "sqlite":
		// 确保 SQLite 数据库文件目录存在
		if cfg.FilePath != "" {
			dir := filepath.Dir(cfg.FilePath)
			if err := os.MkdirAll(dir, 0755); err != nil {
				status.Status = "✗ Failed"
				status.Error = fmt.Sprintf("failed to create directory: %v", err)
				return status
			}
		}
		dialector = sqlite.Open(cfg.FilePath)
		status.Name = "SQLite"

	default:
		status.Status = "✗ Failed"
		status.Error = fmt.Sprintf("unsupported database type: %s", cfg.Type)
		return status
	}

	gormConfig := &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	}

	db, err := gorm.Open(dialector, gormConfig)
	if err != nil {
		status.Status = "✗ Failed"
		status.Error = err.Error()
		return status
	}

	sqlDB, err := db.DB()
	if err != nil {
		status.Status = "✗ Failed"
		status.Error = err.Error()
		return status
	}
	defer func() {
		_ = sqlDB.Close()
	}()

	if err := sqlDB.Ping(); err != nil {
		status.Status = "✗ Failed"
		status.Error = err.Error()
		return status
	}

	status.Status = "✓ Connected"
	if cfg.Type == "sqlite" {
		status.Details = cfg.FilePath
	} else {
		status.Details = fmt.Sprintf("%s:%s/%s", cfg.Host, cfg.Port, cfg.DBName)
	}
	return status
}

// testRedisConnection tests Redis connection
func testRedisConnection(cfg RedisConfig) ServiceStatus {
	status := ServiceStatus{
		Name: "Redis",
	}

	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	defer func() {
		_ = client.Close()
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		status.Status = "✗ Failed"
		status.Error = err.Error()
		return status
	}

	status.Status = "✓ Connected"
	status.Details = fmt.Sprintf("%s:%s (DB %d)", cfg.Host, cfg.Port, cfg.DB)
	return status
}

// testZLMediaKitConnection tests ZLMediaKit HTTP API connection
func testZLMediaKitConnection(cfg ZLMediaKitConfig) ServiceStatus {
	status := ServiceStatus{
		Name: "ZLMediaKit",
	}

	url := fmt.Sprintf("http://%s:%s/index/api/getServerConfig", cfg.Host, cfg.Port)
	if cfg.Secret != "" {
		url += "?secret=" + cfg.Secret
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		status.Status = "✗ Failed"
		status.Error = err.Error()
		return status
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		status.Status = "✗ Failed"
		status.Error = fmt.Sprintf("HTTP %d", resp.StatusCode)
		return status
	}

	status.Status = "✓ Connected"
	status.Details = fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	return status
}
