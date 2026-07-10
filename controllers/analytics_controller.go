package controllers

import (
	"dfrecap/config"
	"dfrecap/db"
	"dfrecap/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func GetAnalyticsOverview(c *gin.Context) {
	var totalUsers int64
	db.DB.Model(&models.Users{}).Where("role = ?", "user").Count(&totalUsers)

	var totalActiveParticipants int64
	db.DB.Model(&models.Scan{}).Select("count(distinct(user_id))").Count(&totalActiveParticipants)

	var totalScans int64
	db.DB.Model(&models.Scan{}).Count(&totalScans)

	avgScansPerUser := 0.0
	if totalActiveParticipants > 0 {
		avgScansPerUser = float64(totalScans) / float64(totalActiveParticipants)
	}

	c.JSON(http.StatusOK, gin.H{
		"total_registered_users":  totalUsers,
		"total_active_participants": totalActiveParticipants,
		"total_scans":             totalScans,
		"average_scans_per_user":    avgScansPerUser,
	})
}

func GetAnalyticsTags(c *gin.Context) {
	var scans []models.Scan
	if err := db.DB.Preload("Session").Find(&scans).Error; err != nil {
		config.Log.Error("Taramalar çekilirken hata oluştu", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "veri çekilemedi"})
		return
	}

	tagStats := make(map[string]int)
	for _, scan := range scans {
		for _, tag := range scan.Session.Tags {
			tagStats[tag]++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"tag_popularity": tagStats,
	})
}

func GetAnalyticsRooms(c *gin.Context) {
	type RoomStat struct {
		RoomName string `json:"room_name"`
		Scans    int    `json:"scans"`
	}

	var results []RoomStat
	err := db.DB.Table("scans").
		Select("sessions.room_name, count(scans.id) as scans").
		Joins("join sessions on sessions.id = scans.session_id").
		Where("sessions.room_name != ''").
		Group("sessions.room_name").
		Order("scans desc").
		Scan(&results).Error

	if err != nil {
		config.Log.Error("Salon istatistikleri çekilirken hata oluştu", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "veri çekilemedi"})
		return
	}

	c.JSON(http.StatusOK, results)
}
