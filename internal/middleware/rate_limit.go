package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// RateLimitConfig 速率限制配置
type RateLimitConfig struct {
	// 时间窗口内允许的最大请求数
	MaxRequests int
	// 时间窗口大小
	Window time.Duration
	// 用于生成限流键的函数（默认使用 IP 地址）
	KeyFunc func(*gin.Context) string
}

// RateLimiter 创建速率限制中间件
func RateLimiter(redisClient *redis.Client, config RateLimitConfig) gin.HandlerFunc {
	// 默认使用 IP 地址作为限流键
	if config.KeyFunc == nil {
		config.KeyFunc = func(c *gin.Context) string {
			return c.ClientIP()
		}
	}

	return func(c *gin.Context) {
		ctx := context.Background()
		key := fmt.Sprintf("rate_limit:%s", config.KeyFunc(c))

		// 使用 Redis 的 INCR 和 EXPIRE 实现滑动窗口限流
		pipe := redisClient.Pipeline()
		incrCmd := pipe.Incr(ctx, key)
		pipe.Expire(ctx, key, config.Window)
		_, err := pipe.Exec(ctx)

		if err != nil {
			// Redis 错误时，记录日志但允许请求通过（fail-open）
			c.Next()
			return
		}

		count := incrCmd.Val()
		if count > int64(config.MaxRequests) {
			// 超过限制，返回 429 Too Many Requests
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "rate limit exceeded",
				"retry_after": int(config.Window.Seconds()),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// LoginRateLimit 登录接口速率限制：每分钟5次
func LoginRateLimit(redisClient *redis.Client) gin.HandlerFunc {
	return RateLimiter(redisClient, RateLimitConfig{
		MaxRequests: 5,
		Window:      1 * time.Minute,
	})
}

// ShareCodeVerifyRateLimit 分享码验证速率限制：每分钟10次
func ShareCodeVerifyRateLimit(redisClient *redis.Client) gin.HandlerFunc {
	return RateLimiter(redisClient, RateLimitConfig{
		MaxRequests: 10,
		Window:      1 * time.Minute,
	})
}

// ShareLinkVerifyRateLimit 分享链接验证速率限制：每分钟20次
func ShareLinkVerifyRateLimit(redisClient *redis.Client) gin.HandlerFunc {
	return RateLimiter(redisClient, RateLimitConfig{
		MaxRequests: 20,
		Window:      1 * time.Minute,
	})
}
