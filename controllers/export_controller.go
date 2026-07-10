package controllers

import (
	"dfrecap/config"
	"dfrecap/db"
	"dfrecap/models"
	"dfrecap/services"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ParticipantRecapExport struct {
	UserID    uint                 `json:"user_id"`
	Username  string               `json:"username"`
	FullName  string               `json:"full_name"`
	Mail      string               `json:"mail"`
	RecapData *services.RecapStats `json:"recap"`
}

func ExportRecaps(c *gin.Context) {
	var participants []models.Users
	err := db.DB.Where("role = ?", "user").Find(&participants).Error
	if err != nil {
		config.Log.Error("Katılımcı listesi dışa aktarılırken hata", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "katılımcılar yüklenemedi"})
		return
	}

	var exportList []ParticipantRecapExport

	for _, user := range participants {
		recap, err := services.GetUserRecap(user.ID)
		if err != nil {
			config.Log.Warn("Katılımcı için recap oluşturulamadı, atlanıyor", zap.Uint("userID", user.ID), zap.Error(err))
			continue
		}

		exportList = append(exportList, ParticipantRecapExport{
			UserID:    user.ID,
			Username:  user.Username,
			FullName:  user.FullName,
			Mail:      user.Mail,
			RecapData: recap,
		})
	}

	config.Log.Info("Katılımcı wrapped verileri dışa aktarıldı", zap.Int("count", len(exportList)))
	c.JSON(http.StatusOK, exportList)
}
