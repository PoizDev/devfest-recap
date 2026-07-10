package services

import (
	"dfrecap/config"
	"dfrecap/db"
	"dfrecap/models"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DevTVFacilitator struct {
	FacilitatorID uint     `gorm:"column:facilitator_id;primaryKey"`
	Name          string   `gorm:"column:name"`
	Title         string   `gorm:"column:title"`
	Topic         string   `gorm:"column:topic"`
	TopicDetails  string   `gorm:"column:topic_details"`
	Tags          []string `gorm:"column:tags;serializer:json"`
}

type DevTVWorkshop struct {
	WorkshopID   uint   `gorm:"column:workshop_id;primaryKey"`
	WorkshopName string `gorm:"column:workshop_name"`
}

type DevTVTimeSlot struct {
	SlotID        uint      `gorm:"column:slot_id;primaryKey"`
	WorkshopID    uint      `gorm:"column:workshop_id"`
	FacilitatorID uint      `gorm:"column:facilitator_id"`
	SlotStart     time.Time `gorm:"column:slot_start"`
	SlotEnd       time.Time `gorm:"column:slot_end"`
}

var (
	devtvConn *gorm.DB
	devtvOnce sync.Once
	devtvErr  error
)

func getDevTVDB() (*gorm.DB, error) {
	devtvOnce.Do(func() {
		dsn := os.Getenv("DEVTV_DSN")
		if dsn == "" {
			devtvErr = errors.New("DEVTV_DSN ortam değişkeni tanımlı değil")
			return
		}
		devtvConn, devtvErr = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if devtvErr != nil {
			config.Log.Error("DevTV veritabanına bağlanılamadı", zap.Error(devtvErr))
			return
		}
		sqlDB, err := devtvConn.DB()
		if err == nil {
			sqlDB.SetMaxIdleConns(5)
			sqlDB.SetMaxOpenConns(10)
			sqlDB.SetConnMaxLifetime(5 * time.Minute)
		}
		config.Log.Info("DevTV veritabanı bağlantısı kuruldu (singleton)")
	})
	return devtvConn, devtvErr
}

func SyncFromDevTV() (int, int, int, error) {
	devtvDB, err := getDevTVDB()
	if err != nil {
		return 0, 0, 0, fmt.Errorf("DevTV veritabanı bağlantı hatası: %w", err)
	}

	var facilitators []DevTVFacilitator
	if err := devtvDB.Table("facilitators").Find(&facilitators).Error; err != nil {
		return 0, 0, 0, fmt.Errorf("facilitators tablosu okunamadı: %w", err)
	}

	var workshops []DevTVWorkshop
	if err := devtvDB.Table("workshops").Find(&workshops).Error; err != nil {
		return 0, 0, 0, fmt.Errorf("workshops tablosu okunamadı: %w", err)
	}

	var slots []DevTVTimeSlot
	if err := devtvDB.Table("workshop_time_slots").Find(&slots).Error; err != nil {
		return 0, 0, 0, fmt.Errorf("workshop_time_slots tablosu okunamadı: %w", err)
	}

	facMap := make(map[uint]DevTVFacilitator)
	for _, f := range facilitators {
		facMap[f.FacilitatorID] = f
	}

	wsMap := make(map[uint]DevTVWorkshop)
	for _, w := range workshops {
		wsMap[w.WorkshopID] = w
	}

	inserted := 0
	updated := 0
	fetched := 0

	qrKeys := make([]string, 0, len(slots))
	type SlotData struct {
		Key      string
		Category string
		Fac      DevTVFacilitator
		Ws       DevTVWorkshop
		Slot     DevTVTimeSlot
	}
	slotDataList := make([]SlotData, 0, len(slots))

	for _, slot := range slots {
		fac, hasFac := facMap[slot.FacilitatorID]
		ws, hasWs := wsMap[slot.WorkshopID]
		if !hasFac || !hasWs {
			continue
		}

		qrKey := fmt.Sprintf("devtv_slot_%d", slot.SlotID)
		category := detectCategory(fac.Topic, ws.WorkshopName)

		qrKeys = append(qrKeys, qrKey)
		slotDataList = append(slotDataList, SlotData{
			Key:      qrKey,
			Category: category,
			Fac:      fac,
			Ws:       ws,
			Slot:     slot,
		})
	}

	if len(qrKeys) == 0 {
		return 0, 0, 0, nil
	}

	var existingSessions []models.Session
	if err := db.DB.Where("qr_code_key IN ?", qrKeys).Find(&existingSessions).Error; err != nil {
		return 0, 0, 0, fmt.Errorf("mevcut oturumlar sorgulanamadı: %w", err)
	}

	existingMap := make(map[string]*models.Session)
	for i := range existingSessions {
		existingMap[existingSessions[i].QRCodeKey] = &existingSessions[i]
	}

	var toInsert []models.Session

	for _, sd := range slotDataList {
		existing, exists := existingMap[sd.Key]

		if exists {
			hasChanges := false

			if existing.Name != sd.Fac.Topic {
				existing.Name = sd.Fac.Topic
				hasChanges = true
			}
			if existing.Description != sd.Fac.TopicDetails {
				existing.Description = sd.Fac.TopicDetails
				hasChanges = true
			}
			if existing.Category != sd.Category {
				existing.Category = sd.Category
				hasChanges = true
			}
			if existing.RoomName != sd.Ws.WorkshopName {
				existing.RoomName = sd.Ws.WorkshopName
				hasChanges = true
			}
			if existing.SpeakerName != sd.Fac.Name {
				existing.SpeakerName = sd.Fac.Name
				hasChanges = true
			}
			if existing.SpeakerTitle != sd.Fac.Title {
				existing.SpeakerTitle = sd.Fac.Title
				hasChanges = true
			}
			if fmt.Sprintf("%v", existing.Tags) != fmt.Sprintf("%v", sd.Fac.Tags) {
				existing.Tags = sd.Fac.Tags
				hasChanges = true
			}
			if existing.StartTime == nil || !existing.StartTime.Equal(sd.Slot.SlotStart) {
				existing.StartTime = &sd.Slot.SlotStart
				hasChanges = true
			}
			if existing.EndTime == nil || !existing.EndTime.Equal(sd.Slot.SlotEnd) {
				existing.EndTime = &sd.Slot.SlotEnd
				hasChanges = true
			}

			if hasChanges {
				if err := db.DB.Save(existing).Error; err != nil {
					config.Log.Error("Session güncelleme hatası", zap.String("qrKey", sd.Key), zap.Error(err))
					continue
				}
				updated++
			} else {
				fetched++
			}
		} else {
			newSess := models.Session{
				Name:         sd.Fac.Topic,
				Description:  sd.Fac.TopicDetails,
				Type:         "session",
				Category:     sd.Category,
				Points:       15,
				QRCodeKey:    sd.Key,
				RoomName:     sd.Ws.WorkshopName,
				SpeakerName:  sd.Fac.Name,
				SpeakerTitle: sd.Fac.Title,
				Tags:         sd.Fac.Tags,
				StartTime:    &sd.Slot.SlotStart,
				EndTime:      &sd.Slot.SlotEnd,
			}
			toInsert = append(toInsert, newSess)
		}
	}

	if len(toInsert) > 0 {
		if err := db.DB.CreateInBatches(&toInsert, 100).Error; err != nil {
			config.Log.Error("Session toplu kayıt hatası", zap.Error(err))
		} else {
			inserted = len(toInsert)
		}
	}

	return inserted, updated, fetched, nil
}

func detectCategory(topic, workshopName string) string {
	text := strings.ToLower(topic + " " + workshopName)
	if strings.Contains(text, "ai") || strings.Contains(text, "yapay zeka") || strings.Contains(text, "data") || strings.Contains(text, "intelligence") || strings.Contains(text, "ml") || strings.Contains(text, "model") {
		return "ai"
	}
	if strings.Contains(text, "flutter") || strings.Contains(text, "android") || strings.Contains(text, "ios") || strings.Contains(text, "kotlin") || strings.Contains(text, "swift") || strings.Contains(text, "mobile") || strings.Contains(text, "mobil") {
		return "mobile"
	}
	if strings.Contains(text, "cloud") || strings.Contains(text, "bulut") || strings.Contains(text, "devops") || strings.Contains(text, "kubernetes") || strings.Contains(text, "docker") || strings.Contains(text, "aws") || strings.Contains(text, "gcp") {
		return "cloud"
	}
	return "backend"
}
