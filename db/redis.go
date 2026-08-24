package db

import (
	"context"
	"log"

	"watch-api/config"

	"github.com/redis/go-redis/v9"
)

func InitRedis(cfg config.Config) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	// 测试连接
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("❌ 连接 Redis 失败: %v", err)
	}

	log.Println("✅ Redis 连接成功")
	return rdb
}
