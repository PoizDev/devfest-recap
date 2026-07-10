package db

import (
	"context"
	"dfrecap/config"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

var RDB *redis.Client

func ConnectRedis(cfg config.RedisConfig) {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	}

	password := os.Getenv("REDIS_PASSWORD")

	RDB = redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := RDB.Ping(ctx).Err(); err != nil {
		config.Log.Error("Redis bağlantısı başarısız", zap.Error(err))
		RDB = nil
		return
	}

	config.Log.Info("Redis bağlantısı başarılı")
}
