package controllers

import (
	"dfrecap/config"
	"dfrecap/services"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func GetRecap(c *gin.Context) {
	if !recapMode {
		c.JSON(http.StatusForbidden, gin.H{"error": "Recap sonuçları henüz açıklanmadı. Bizi izlemeye devam edin!"})
		return
	}

	userIDVal, exists := c.Get("id")
	if !exists {
		config.Log.Warn("Yetkisiz erişim denemesi (Kullanıcı kimliği yok)", zap.String("action", "recap"))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "kullanıcı kimliği bulunamadı"})
		return
	}

	var userID uint
	switch v := userIDVal.(type) {
	case int:
		userID = uint(v)
	case uint:
		userID = v
	default:
		config.Log.Warn("Geçersiz kullanıcı kimliği tipi", zap.String("action", "recap"))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "geçersiz kullanıcı kimliği tipi"})
		return
	}

	recap, err := services.GetUserRecap(userID)
	if err != nil {
		config.Log.Error("Kullanıcı recap bilgisi hesaplanamadı", zap.Error(err), zap.Uint("userID", userID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "recap bilgileri oluşturulamadı"})
		return
	}

	c.JSON(http.StatusOK, recap)
}
