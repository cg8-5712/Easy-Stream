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

// DeleteStreamAccessTokens 删除指定直播的所有访问令牌（直播结束时调用）
func (r *RedisClient) DeleteStreamAccessTokens(streamKey string) error {
	ctx := context.Background()
	pattern := fmt.Sprintf("stream_access:%s:*", streamKey)

	var cursor uint64
	for {
		keys, nextCursor, err := r.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return err
		}

		if len(keys) > 0 {
			if err := r.Del(ctx, keys...).Err(); err != nil {
				return err
			}
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	return nil
}

// GetStreamKeyByAccessToken 通过访问令牌获取 stream_key
func (r *RedisClient) GetStreamKeyByAccessToken(token string) (string, error) {
	ctx := context.Background()
	pattern := fmt.Sprintf("stream_access:*:%s", token)

	keys, _, err := r.Scan(ctx, 0, pattern, 10).Result()
	if err != nil {
		return "", err
	}

	if len(keys) == 0 {
		return "", nil
	}

	// 从 key 中提取 stream_key，格式: stream_access:{streamKey}:{token}
	key := keys[0]
	// 去掉前缀 "stream_access:" 和后缀 ":{token}"
	prefix := "stream_access:"
	suffix := ":" + token
	if len(key) > len(prefix)+len(suffix) {
		return key[len(prefix) : len(key)-len(suffix)], nil
	}
	return "", nil
}

