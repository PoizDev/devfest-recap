package controllers

import (
	"context"
	"dfrecap/config"
	"dfrecap/db"
	"dfrecap/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type LeaderboardEntry struct {
	Rank      int    `json:"rank"`
	UserID    uint   `json:"user_id"`
	Username  string `json:"username"`
	FullName  string `json:"full_name"`
	Points    int    `json:"points"`
	ScanCount int    `json:"scan_count"`
}

func GetLeaderboard(c *gin.Context) {
	ctx := context.Background()

	if db.RDB != nil {
		redisEntries, err := db.RDB.ZRevRangeWithScores(ctx, "leaderboard", 0, 9).Result()
		if err == nil && len(redisEntries) > 0 {
			var userIDs []uint
			pointsMap := make(map[uint]int)

			for _, entry := range redisEntries {
				uID, parseErr := strconv.Atoi(entry.Member.(string))
				if parseErr != nil {
					continue
				}
				userID := uint(uID)
				userIDs = append(userIDs, userID)
				pointsMap[userID] = int(entry.Score)
			}

			var users []models.Users
			if len(userIDs) > 0 {
				db.DB.Where("id IN ?", userIDs).Find(&users)
			}

			userMap := make(map[uint]models.Users)
			for _, user := range users {
				userMap[user.ID] = user
			}

			// Batch fetch scan counts for these users
			var scanCounts []struct {
				UserID uint
				Count  int
			}
			scanCountMap := make(map[uint]int)
			if len(userIDs) > 0 {
				db.DB.Table("scans").Select("user_id, count(id) as count").Where("user_id IN ?", userIDs).Group("user_id").Scan(&scanCounts)
				for _, sc := range scanCounts {
					scanCountMap[sc.UserID] = sc.Count
				}
			}

			var response []LeaderboardEntry
			for i, entry := range redisEntries {
				uID, _ := strconv.Atoi(entry.Member.(string))
				userID := uint(uID)
				user, exists := userMap[userID]
				
				username := "Bilinmeyen Kullanıcı"
				fullName := "Katılımcı"
				if exists {
					username = user.Username
					fullName = user.FullName
				}

				response = append(response, LeaderboardEntry{
					Rank:      i + 1,
					UserID:    userID,
					Username:  username,
					FullName:  fullName,
					Points:    pointsMap[userID],
					ScanCount: scanCountMap[userID],
				})
			}

			c.JSON(http.StatusOK, response)
			return
		}
	}

	// Redis yoksa veya boşsa, PostgreSQL ile hesaplama yap (Fallback)
	config.Log.Warn("Redis liderlik tablosu okunamadı, veritabanından hesaplanıyor...", zap.String("action", "leaderboard_fallback"))

	type FallbackEntry struct {
		UserID    uint
		Username  string
		FullName  string
		Points    int64
		ScanCount int
	}

	var rawEntries []FallbackEntry
	err := db.DB.Table("scans").
		Select("scans.user_id, users.username, users.full_name, sum(sessions.points) as points, count(scans.id) as scan_count").
		Joins("join users on users.id = scans.user_id").
		Joins("join sessions on sessions.id = scans.session_id").
		Group("scans.user_id, users.username, users.full_name").
		Order("points desc").
		Limit(10).
		Scan(&rawEntries).Error

	if err != nil {
		config.Log.Error("Veritabanından liderlik tablosu hesaplanamadı", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "liderlik tablosu yüklenemedi"})
		return
	}

	var response []LeaderboardEntry
	for i, entry := range rawEntries {
		response = append(response, LeaderboardEntry{
			Rank:      i + 1,
			UserID:    entry.UserID,
			Username:  entry.Username,
			FullName:  entry.FullName,
			Points:    int(entry.Points),
			ScanCount: entry.ScanCount,
		})
	}

	c.JSON(http.StatusOK, response)
}
