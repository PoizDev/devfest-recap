package services

import (
	"context"
	"dfrecap/config"
	"dfrecap/db"
	"dfrecap/models"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RecapStats struct {
	TotalScans        int            `json:"total_scans"`
	TotalPoints       int            `json:"total_points"`
	Rank              int            `json:"rank"`
	Percentile        float64        `json:"percentile"`
	TotalParticipants int            `json:"total_participants"`
	CategoryStats     map[string]int `json:"category_stats"`
	DominantCategory  string         `json:"dominant_category"`
	FavoriteRoom      string         `json:"favorite_room"`
	FavoriteSpeaker   string         `json:"favorite_speaker"`
	FavoriteTag       string         `json:"favorite_tag"`
	TotalTalks        int            `json:"total_talks"`
	TypeStats         map[string]int `json:"type_stats"`
	TagStats          map[string]int `json:"tag_stats"`
	FirstScanTime     string         `json:"first_scan_time"`
	LastScanTime      string         `json:"last_scan_time"`
	Badges            []BadgeDetails `json:"badges"`
	ScannedSessions   []SessionInfo  `json:"scanned_sessions"`
}

type BadgeDetails struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
}

type SessionInfo struct {
	Name         string   `json:"name"`
	Type         string   `json:"type"`
	Category     string   `json:"category"`
	Points       int      `json:"points"`
	RoomName     string   `json:"room_name,omitempty"`
	SpeakerName  string   `json:"speaker_name,omitempty"`
	SpeakerTitle string   `json:"speaker_title,omitempty"`
	Tags         []string `json:"tags,omitempty"`
	ScannedAt    string   `json:"scanned_at"`
}

func GetUserRecap(userID uint) (*RecapStats, error) {
	// PERF-04: Önce Redis cache kontrol et
	if db.RDB != nil {
		ctx := context.Background()
		cacheKey := fmt.Sprintf("recap:%d", userID)
		if cached, err := db.RDB.Get(ctx, cacheKey).Result(); err == nil {
			var stats RecapStats
			if json.Unmarshal([]byte(cached), &stats) == nil {
				return &stats, nil
			}
		}
	}

	var scans []models.Scan
	err := db.DB.Preload("Session").Where("user_id = ?", userID).Order("scanned_at asc").Find(&scans).Error
	if err != nil {
		return nil, err
	}

	totalScans := len(scans)
	totalPoints := 0
	categoryStats := map[string]int{
		"backend": 0,
		"mobile":  0,
		"ai":      0,
		"cloud":   0,
	}

	var scannedSessions []SessionInfo
	typeCount := map[string]int{
		"stand":    0,
		"session":  0,
		"workshop": 0,
	}

	roomStats := make(map[string]int)
	speakerStats := make(map[string]int)
	tagStats := make(map[string]int)
	totalTalks := 0

	firstScanTime := ""
	lastScanTime := ""

	if len(scans) > 0 {
		firstScanTime = scans[0].ScannedAt.Format("15:04")
		lastScanTime = scans[len(scans)-1].ScannedAt.Format("15:04")
	}

	for _, scan := range scans {
		sess := scan.Session
		totalPoints += sess.Points
		categoryStats[sess.Category]++
		typeCount[sess.Type]++
		if sess.Type == "session" {
			totalTalks++
		}

		if sess.RoomName != "" {
			roomStats[sess.RoomName]++
		}
		if sess.SpeakerName != "" {
			speakerStats[sess.SpeakerName]++
		}
		for _, tag := range sess.Tags {
			tagStats[tag]++
		}

		scannedSessions = append(scannedSessions, SessionInfo{
			Name:         sess.Name,
			Type:         sess.Type,
			Category:     sess.Category,
			Points:       sess.Points,
			RoomName:     sess.RoomName,
			SpeakerName:  sess.SpeakerName,
			SpeakerTitle: sess.SpeakerTitle,
			Tags:         sess.Tags,
			ScannedAt:    scan.ScannedAt.Format("15:04"),
		})
	}

	dominantCategory := "Keşfedilmemiş"
	maxScans := 0
	for cat, count := range categoryStats {
		if count > maxScans {
			maxScans = count
			dominantCategory = cat
		}
	}

	favoriteRoom := "Keşfedilmemiş"
	maxRoomScans := 0
	for rName, count := range roomStats {
		if count > maxRoomScans {
			maxRoomScans = count
			favoriteRoom = rName
		}
	}

	favoriteSpeaker := "Keşfedilmemiş"
	maxSpeakerScans := 0
	for sName, count := range speakerStats {
		if count > maxSpeakerScans {
			maxSpeakerScans = count
			favoriteSpeaker = sName
		}
	}

	favoriteTag := "Keşfedilmemiş"
	maxTagScans := 0
	for tag, count := range tagStats {
		if count > maxTagScans {
			maxTagScans = count
			favoriteTag = tag
		}
	}

	categoryNamesTR := map[string]string{
		"backend": "Backend & Sunucu Teknolojileri",
		"mobile":  "Mobil Uygulama Geliştirme",
		"ai":      "Yapay Zeka & Veri Bilimi",
		"cloud":   "Bulut Bilişim & DevOps",
	}
	if dominantCategory != "Keşfedilmemiş" {
		dominantCategory = categoryNamesTR[dominantCategory]
	}

	rank := 0
	if db.RDB != nil {
		ctx := context.Background()
		redisRank, err := db.RDB.ZRevRank(ctx, "leaderboard", fmt.Sprintf("%d", userID)).Result()
		if err == nil {
			rank = int(redisRank) + 1
		}
	}

	if rank == 0 {
		var dbRank int64
		db.DB.Table("scans").
			Select("scans.user_id, sum(sessions.points) as pts").
			Joins("join sessions on sessions.id = scans.session_id").
			Group("scans.user_id").
			Having("sum(sessions.points) > ?", totalPoints).
			Count(&dbRank)
		rank = int(dbRank) + 1
	}

	var totalParticipants int64
	db.DB.Model(&models.Scan{}).Select("count(distinct(user_id))").Count(&totalParticipants)
	if totalParticipants == 0 {
		totalParticipants = 1
	}

	percentile := 100.0 - (float64(rank-1)/float64(totalParticipants))*100.0
	if percentile < 0 {
		percentile = 0
	} else if percentile > 100 {
		percentile = 100
	}

	var badges []BadgeDetails

	if totalScans >= 1 {
		badges = append(badges, BadgeDetails{
			Name:        "İlk Adım",
			Description: "DevFest macerasına ilk QR kodunu taratarak başladın!",
			Icon:        "🏁",
		})
	}

	if typeCount["stand"] >= 5 {
		badges = append(badges, BadgeDetails{
			Name:        "Stant Avcısı",
			Description: "Neredeyse tüm stantları gezip harika etkileşimler topladın!",
			Icon:        "🎯",
		})
	}

	if categoryStats["ai"] >= 2 {
		badges = append(badges, BadgeDetails{
			Name:        "AI Meraklısı",
			Description: "Yapay zeka ve model oturumlarına özel ilgi gösterdin.",
			Icon:        "🧠",
		})
	}

	if categoryStats["cloud"] >= 2 {
		badges = append(badges, BadgeDetails{
			Name:        "Bulut Gezgini",
			Description: "Bulut çözümleri ve Kubernetes stantlarını fethettin!",
			Icon:        "☁️",
		})
	}

	if categoryStats["backend"] >= 2 {
		badges = append(badges, BadgeDetails{
			Name:        "Sistem Mimarı",
			Description: "Go, veritabanları ve ölçeklenebilirlik oturumlarının yıldızı oldun.",
			Icon:        "⚙️",
		})
	}

	if categoryStats["mobile"] >= 2 {
		badges = append(badges, BadgeDetails{
			Name:        "Mobil Kodlayıcı",
			Description: "Flutter, Kotlin ve mobil dünya oturumlarını kaçırmadın.",
			Icon:        "📱",
		})
	}

	if categoryStats["backend"] >= 1 && categoryStats["mobile"] >= 1 && categoryStats["ai"] >= 1 && categoryStats["cloud"] >= 1 {
		badges = append(badges, BadgeDetails{
			Name:        "Full-Stack Teknoloji Gurusu",
			Description: "Tüm alanlardan bilgi toplayan gerçek bir teknoloji sevdalısı!",
			Icon:        "🔥",
		})
	}

	if db.RDB == nil {
		config.Log.Debug("Redis aktif olmadığı için leaderboard sıralaması veritabanı yedeğinden hesaplandı.")
	}

	result := &RecapStats{
		TotalScans:        totalScans,
		TotalPoints:       totalPoints,
		Rank:              rank,
		Percentile:        percentile,
		TotalParticipants: int(totalParticipants),
		CategoryStats:     categoryStats,
		DominantCategory:  dominantCategory,
		FavoriteRoom:      favoriteRoom,
		FavoriteSpeaker:   favoriteSpeaker,
		FavoriteTag:       favoriteTag,
		TotalTalks:        totalTalks,
		TypeStats:         typeCount,
		TagStats:          tagStats,
		FirstScanTime:     firstScanTime,
		LastScanTime:      lastScanTime,
		Badges:            badges,
		ScannedSessions:   scannedSessions,
	}

	// PERF-04: Sonucu Redis'e cache'le (60s TTL)
	if db.RDB != nil {
		ctx := context.Background()
		cacheKey := fmt.Sprintf("recap:%d", userID)
		if jsonBytes, err := json.Marshal(result); err == nil {
			db.RDB.Set(ctx, cacheKey, jsonBytes, 60*time.Second)
		}
	}

	return result, nil
}

func SyncLeaderboardFromDB() error {
	if db.RDB == nil {
		return redis.Nil
	}

	type ScoreStruct struct {
		UserID uint
		Pts    int64
	}

	var scores []ScoreStruct
	err := db.DB.Table("scans").
		Select("scans.user_id, sum(sessions.points) as pts").
		Joins("join sessions on sessions.id = scans.session_id").
		Group("scans.user_id").
		Scan(&scores).Error

	if err != nil {
		return err
	}

	ctx := context.Background()

	if len(scores) == 0 {
		db.RDB.Del(ctx, "leaderboard")
		return nil
	}

	// PERF-05: Geçici key'e yaz, sonra atomik RENAME yap
	tempKey := "leaderboard:sync_tmp"
	pipe := db.RDB.Pipeline()
	pipe.Del(ctx, tempKey)

	for _, s := range scores {
		pipe.ZAdd(ctx, tempKey, redis.Z{
			Score:  float64(s.Pts),
			Member: fmt.Sprintf("%d", s.UserID),
		})
	}

	pipe.Rename(ctx, tempKey, "leaderboard")

	_, err = pipe.Exec(ctx)
	return err
}
