package database

import (
	"context"
	"errors"
	"time"

	"course_server/internal/config"

	"github.com/redis/go-redis/v9"
)

// RDB 全局 Redis 客户端；一期 handler 未使用，仅作连接占位（Day1）。
var RDB *redis.Client

// InitRedis 连接 Redis 并 Ping 校验；addr 为空时返回错误。
func InitRedis(cfg config.RedisConfig) error {
	if cfg.Addr == "" {
		return errors.New("redis addr is empty")
	}
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return err
	}
	RDB = client
	return nil
}
