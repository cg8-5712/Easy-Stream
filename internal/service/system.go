package service

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"net/http"
	"time"

	"easy-stream/internal/config"
	"easy-stream/internal/repository"
	"easy-stream/internal/zlm"
)

// SystemService 系统服务
type SystemService struct {
	db        *sql.DB
	redis     *repository.RedisClient
	zlmClient *zlm.Client
	zlmConfig config.ZLMediaKitConfig
	startTime time.Time
}

// NewSystemService 创建系统服务
func NewSystemService(db *sql.DB, redis *repository.RedisClient, zlmConfig config.ZLMediaKitConfig) *SystemService {
	return &SystemService{
		db:        db,
		redis:     redis,
		zlmClient: zlm.NewClient(zlmConfig.Host, zlmConfig.Port, zlmConfig.Secret),
		zlmConfig: zlmConfig,
		startTime: time.Now(),
	}
}

// HealthStatus 健康状态
type HealthStatus struct {
	Status    string                    `json:"status"` // healthy / degraded / unhealthy
	Timestamp time.Time                 `json:"timestamp"`
	Uptime    string                    `json:"uptime"`
	Version   string                    `json:"version"`
	Services  map[string]*ServiceHealth `json:"services"`
	Network   *NetworkHealth            `json:"network"`
}

// ServiceHealth 服务健康状态
type ServiceHealth struct {
	Status  string `json:"status"`            // up / down / auth_failed
	Latency string `json:"latency"`           // 响应延迟
	Message string `json:"message,omitempty"` // 错误信息
}

// NetworkHealth 网络健康状态
type NetworkHealth struct {
	DNS      *ServiceHealth `json:"dns"`
	Internet *ServiceHealth `json:"internet"`
}

// CheckHealth 检查所有服务健康状态
func (s *SystemService) CheckHealth() *HealthStatus {
	health := &HealthStatus{
		Status:    "healthy",
		Timestamp: time.Now(),
		Uptime:    s.formatUptime(),
		Version:   "2.0",
		Services:  make(map[string]*ServiceHealth),
		Network:   &NetworkHealth{},
	}

	// 检查 PostgreSQL
	health.Services["postgresql"] = s.checkPostgres()

	// 检查 Redis
	health.Services["redis"] = s.checkRedis()

	// 检查 ZLMediaKit
	health.Services["zlmediakit"] = s.checkZLMediaKit()

	// 检查网络
	health.Network.DNS = s.checkDNS()
	health.Network.Internet = s.checkInternet()

	// 计算总体状态
	health.Status = s.calculateOverallStatus(health)

	return health
}

// checkPostgres 检查 PostgreSQL 连接和认证
func (s *SystemService) checkPostgres() *ServiceHealth {
	start := time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 不仅 Ping，还执行一个简单查询来验证凭据和权限
	var result int
	err := s.db.QueryRowContext(ctx, "SELECT 1").Scan(&result)
	latency := time.Since(start)

	if err != nil {
		// 检查是否是认证错误
		errMsg := err.Error()
		if contains(errMsg, "password authentication failed") ||
			contains(errMsg, "authentication failed") ||
			contains(errMsg, "FATAL") {
			return &ServiceHealth{
				Status:  "auth_failed",
				Latency: latency.String(),
				Message: "authentication failed: invalid username or password",
			}
		}
		return &ServiceHealth{
			Status:  "down",
			Latency: latency.String(),
			Message: fmt.Sprintf("connection failed: %v", err),
		}
	}

	return &ServiceHealth{
		Status:  "up",
		Latency: latency.String(),
	}
}

// checkRedis 检查 Redis 连接和认证
func (s *SystemService) checkRedis() *ServiceHealth {
	start := time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 使用一个临时 key 进行完整的读写测试，验证认证和权限
	testKey := "_health_check_test_"
	testValue := fmt.Sprintf("test_%d", time.Now().UnixNano())

	// 尝试写入
	err := s.redis.Set(ctx, testKey, testValue, 10*time.Second).Err()
	if err != nil {
		latency := time.Since(start)
		// 检查是否是认证错误
		errMsg := err.Error()
		if contains(errMsg, "NOAUTH") ||
			contains(errMsg, "AUTH") ||
			contains(errMsg, "invalid password") ||
			contains(errMsg, "WRONGPASS") {
			return &ServiceHealth{
				Status:  "auth_failed",
				Latency: latency.String(),
				Message: "authentication failed: invalid password",
			}
		}
		return &ServiceHealth{
			Status:  "down",
			Latency: latency.String(),
			Message: fmt.Sprintf("connection failed: %v", err),
		}
	}

	// 尝试读取验证
	val, err := s.redis.Get(ctx, testKey).Result()
	if err != nil {
		latency := time.Since(start)
		return &ServiceHealth{
			Status:  "down",
			Latency: latency.String(),
			Message: fmt.Sprintf("read verification failed: %v", err),
		}
	}

	// 验证值是否正确
	if val != testValue {
		latency := time.Since(start)
		return &ServiceHealth{
			Status:  "down",
			Latency: latency.String(),
			Message: "data verification failed: value mismatch",
		}
	}

	// 清理测试 key
	s.redis.Del(ctx, testKey)

	latency := time.Since(start)
	return &ServiceHealth{
		Status:  "up",
		Latency: latency.String(),
	}
}

// checkZLMediaKit 检查 ZLMediaKit 连接和 Secret 验证
func (s *SystemService) checkZLMediaKit() *ServiceHealth {
	start := time.Now()

	// 尝试获取服务器配置来验证连接和 Secret
	resp, err := s.zlmClient.GetServerConfig()
	latency := time.Since(start)

	if err != nil {
		// 网络连接失败
		return &ServiceHealth{
			Status:  "down",
			Latency: latency.String(),
			Message: fmt.Sprintf("connection failed: %v", err),
		}
	}

	// ZLMediaKit API 返回 code:
	// 0 = 成功
	// -1 = secret 验证失败
	// 其他负数 = 其他错误
	if resp.Code != 0 {
		if resp.Code == -1 {
			return &ServiceHealth{
				Status:  "auth_failed",
				Latency: latency.String(),
				Message: "authentication failed: invalid secret",
			}
		}
		return &ServiceHealth{
			Status:  "down",
			Latency: latency.String(),
			Message: fmt.Sprintf("API error: code=%d", resp.Code),
		}
	}

	return &ServiceHealth{
		Status:  "up",
		Latency: latency.String(),
	}
}

// checkDNS 检查 DNS 解析
func (s *SystemService) checkDNS() *ServiceHealth {
	start := time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resolver := &net.Resolver{}
	_, err := resolver.LookupHost(ctx, "www.baidu.com")
	latency := time.Since(start)

	if err != nil {
		return &ServiceHealth{
			Status:  "down",
			Latency: latency.String(),
			Message: fmt.Sprintf("DNS resolution failed: %v", err),
		}
	}

	return &ServiceHealth{
		Status:  "up",
		Latency: latency.String(),
	}
}

// checkInternet 检查外网连接
func (s *SystemService) checkInternet() *ServiceHealth {
	start := time.Now()

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get("https://www.baidu.com")
	latency := time.Since(start)

	if err != nil {
		return &ServiceHealth{
			Status:  "down",
			Latency: latency.String(),
			Message: fmt.Sprintf("internet connection failed: %v", err),
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return &ServiceHealth{
			Status:  "down",
			Latency: latency.String(),
			Message: fmt.Sprintf("unexpected status code: %d", resp.StatusCode),
		}
	}

	return &ServiceHealth{
		Status:  "up",
		Latency: latency.String(),
	}
}

// calculateOverallStatus 计算总体健康状态
func (s *SystemService) calculateOverallStatus(health *HealthStatus) string {
	criticalDown := false
	anyDown := false

	// PostgreSQL 和 Redis 是关键服务
	pgStatus := health.Services["postgresql"].Status
	redisStatus := health.Services["redis"].Status

	if pgStatus == "down" || pgStatus == "auth_failed" {
		criticalDown = true
	}
	if redisStatus == "down" || redisStatus == "auth_failed" {
		criticalDown = true
	}

	// 检查所有服务
	for _, svc := range health.Services {
		if svc.Status == "down" || svc.Status == "auth_failed" {
			anyDown = true
		}
	}

	// 检查网络
	if health.Network.DNS.Status == "down" || health.Network.Internet.Status == "down" {
		anyDown = true
	}

	if criticalDown {
		return "unhealthy"
	}
	if anyDown {
		return "degraded"
	}
	return "healthy"
}

// formatUptime 格式化运行时间
func (s *SystemService) formatUptime() string {
	uptime := time.Since(s.startTime)

	days := int(uptime.Hours()) / 24
	hours := int(uptime.Hours()) % 24
	minutes := int(uptime.Minutes()) % 60
	seconds := int(uptime.Seconds()) % 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm %ds", days, hours, minutes, seconds)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
	}
	if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}

// contains 检查字符串是否包含子串（不区分大小写）
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && containsIgnoreCase(s, substr)))
}

func containsIgnoreCase(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if equalFoldAt(s, i, substr) {
			return true
		}
	}
	return false
}

func equalFoldAt(s string, start int, substr string) bool {
	for i := 0; i < len(substr); i++ {
		c1 := s[start+i]
		c2 := substr[i]
		if c1 != c2 {
			// 简单的大小写转换
			if c1 >= 'A' && c1 <= 'Z' {
				c1 += 'a' - 'A'
			}
			if c2 >= 'A' && c2 <= 'Z' {
				c2 += 'a' - 'A'
			}
			if c1 != c2 {
				return false
			}
		}
	}
	return true
}
