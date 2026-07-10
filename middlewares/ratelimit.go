package middlewares

import (
	"context"
	"dfrecap/config"
	"dfrecap/db"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// RateLimitMiddleware Redis tabanlı sliding window rate limiter.
// Redis yoksa veya hata olursa isteği geçirir (fail-open).
func RateLimitMiddleware(cfg config.RateLimitConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		if db.RDB == nil {
			c.Next()
			return
		}

		ctx := context.Background()
		key := fmt.Sprintf("rl:%s", c.ClientIP())
		now := time.Now()
		windowStart := now.Add(-cfg.WindowDuration)

		pipe := db.RDB.Pipeline()
		pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", windowStart.UnixMilli()))
		pipe.ZAdd(ctx, key, redis.Z{
			Score:  float64(now.UnixMilli()),
			Member: fmt.Sprintf("%d", now.UnixNano()),
		})
		countCmd := pipe.ZCard(ctx, key)
		pipe.Expire(ctx, key, cfg.WindowDuration+time.Second)

		_, err := pipe.Exec(ctx)
		if err != nil {
			config.Log.Warn("Rate limit kontrolü başarısız, istek geçiriliyor", zap.Error(err))
			c.Next()
			return
		}

		if countCmd.Val() > cfg.RequestsPerMinute {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Çok fazla istek gönderdiniz, lütfen biraz bekleyin",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
