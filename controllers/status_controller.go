package controllers

import (
	"context"
	"dfrecap/db"
	"net/http"

	"github.com/gin-gonic/gin"
)

var recapMode bool

func SetRecapMode(mode bool) {
	recapMode = mode
}

func GetStatus(c *gin.Context) {
	dbOK := false
	if db.DB != nil {
		sqlDB, err := db.DB.DB()
		if err == nil {
			dbOK = sqlDB.Ping() == nil
		}
	}

	redisOK := false
	if db.RDB != nil {
		ctx := context.Background()
		redisOK = db.RDB.Ping(ctx).Err() == nil
	}

	status := "healthy"
	httpCode := http.StatusOK
	if !dbOK {
		status = "unhealthy"
		httpCode = http.StatusServiceUnavailable
	}

	c.JSON(httpCode, gin.H{
		"status":     status,
		"recap_mode": recapMode,
		"database":   dbOK,
		"redis":      redisOK,
	})
}
