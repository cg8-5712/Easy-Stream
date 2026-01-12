package repository

import (
	"context"
	"fmt"

	"easy-stream/internal/config"

	"github.com/redis/go-redis/v9"
)

// RedisClient Redis 客户端包装
type RedisClient struct {
	*redis.Client
}

// NewRedisClient 创建 Redis 连接
func NewRedisClient(cfg config.RedisConfig) (*RedisClient, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}

	return &RedisClient{client}, nil
}
