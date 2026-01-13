package repository

import (
	"context"
	"fmt"
	"strconv"
	"time"

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

// SetRefreshToken 存储 Refresh Token
func (r *RedisClient) SetRefreshToken(userID int64, token string, expiration time.Duration) error {
	ctx := context.Background()
	key := fmt.Sprintf("refresh_token:%s", token)
	return r.Set(ctx, key, userID, expiration).Err()
}

// GetUserIDByRefreshToken 通过 Refresh Token 获取用户 ID
func (r *RedisClient) GetUserIDByRefreshToken(token string) (int64, error) {
	ctx := context.Background()
	key := fmt.Sprintf("refresh_token:%s", token)
	val, err := r.Get(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(val, 10, 64)
}

// DeleteRefreshToken 删除 Refresh Token
func (r *RedisClient) DeleteRefreshToken(token string) error {
	ctx := context.Background()
	key := fmt.Sprintf("refresh_token:%s", token)
	return r.Del(ctx, key).Err()
}

// SetStreamAccessToken 存储私有直播访问令牌
func (r *RedisClient) SetStreamAccessToken(streamKey, token string, expiration time.Duration) error {
	ctx := context.Background()
	key := fmt.Sprintf("stream_access:%s:%s", streamKey, token)
	return r.Set(ctx, key, "1", expiration).Err()
}

// VerifyStreamAccessToken 验证私有直播访问令牌
func (r *RedisClient) VerifyStreamAccessToken(streamKey, token string) (bool, error) {
	ctx := context.Background()
	key := fmt.Sprintf("stream_access:%s:%s", streamKey, token)
	val, err := r.Get(ctx, key).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return val == "1", nil
}

