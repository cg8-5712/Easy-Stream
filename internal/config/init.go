package config

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
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
  host: "localhost"
  port: "5432"
  user: "postgres"
  password: "your_password"
  dbname: "easy_stream"
  sslmode: "disable"

redis:
  host: "localhost"
  port: "6379"
  password: ""
  db: 0

jwt:
  secret: "your-secret-key-change-this"
  expirehour: 24

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
	if cfg.JWT.ExpireHour <= 0 {
		return fmt.Errorf("jwt.expirehour must be greater than 0")
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

// testDatabaseConnection tests PostgreSQL database connection
func testDatabaseConnection(cfg DatabaseConfig) ServiceStatus {
	status := ServiceStatus{
		Name: "PostgreSQL",
	}

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	gormConfig := &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	}

	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
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
	defer sqlDB.Close()

	if err := sqlDB.Ping(); err != nil {
		status.Status = "✗ Failed"
		status.Error = err.Error()
		return status
	}

	status.Status = "✓ Connected"
	status.Details = fmt.Sprintf("%s:%s/%s", cfg.Host, cfg.Port, cfg.DBName)
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
	defer client.Close()

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
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		status.Status = "✗ Failed"
		status.Error = fmt.Sprintf("HTTP %d", resp.StatusCode)
		return status
	}

	status.Status = "✓ Connected"
	status.Details = fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	return status
}
