package controllers

import (
	"dfrecap/config"
	"dfrecap/services"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func SyncFromDevTV(c *gin.Context) {
	inserted, updated, fetched, err := services.SyncFromDevTV()
	if err != nil {
		config.Log.Error("Manuel senkronizasyon başarısız", zap.String("action", "manual_sync"), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	config.Log.Info("Manuel senkronizasyon başarılı", zap.String("action", "manual_sync"), zap.Int("inserted", inserted), zap.Int("updated", updated))
	c.JSON(http.StatusOK, gin.H{
		"message":  "DevTV veritabanı ile başarıyla senkronize edildi",
		"inserted": inserted,
		"updated":  updated,
		"fetched":  fetched,
		"total":    inserted + updated + fetched,
	})
}
