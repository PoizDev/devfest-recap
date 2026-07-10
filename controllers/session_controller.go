package controllers

import (
	"dfrecap/config"
	"dfrecap/db"
	"dfrecap/models"
	"dfrecap/services"
	"dfrecap/utils"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/skip2/go-qrcode"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func Scan(c *gin.Context) {
	userIDVal, exists := c.Get("id")
	if !exists {
		config.Log.Warn("Yetkisiz erişim denemesi (Kullanıcı kimliği yok)", zap.String("action", "scan"))
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
		config.Log.Warn("Geçersiz kullanıcı kimliği tipi", zap.String("action", "scan"))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "geçersiz kullanıcı kimliği tipi"})
		return
	}

	var body struct {
		QRCodeKey string `json:"qr_code_key" binding:"required"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		config.Log.Warn("Karekod okutma verisi doğrulanamadı", zap.String("action", "scan"), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "karekod anahtarı gereklidir"})
		return
	}

	scan, err := services.ScanQRCode(userID, body.QRCodeKey)
	if err != nil {
		if err.Error() == "bu seansı daha önce zaten okuttunuz" {
			config.Log.Warn("Çifte okutma (Race Condition veya tekrar) engellendi", zap.Uint("userID", userID), zap.String("qrCodeKey", body.QRCodeKey))
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Seans başarıyla okutuldu!",
		"scan":    scan,
	})
}

func CreateSession(c *gin.Context) {
	var body struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		Type        string `json:"type" binding:"required"`
		Category    string `json:"category" binding:"required"`
		Points      int    `json:"points" binding:"required"`
		QRCodeKey   string `json:"qr_code_key" binding:"required"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		config.Log.Warn("Seans oluşturma verisi geçersiz", zap.String("action", "create_session"), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if body.Type != "stand" && body.Type != "session" && body.Type != "workshop" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "geçersiz seans tipi ('stand', 'session' veya 'workshop' olmalıdır)"})
		return
	}

	if body.Category != "backend" && body.Category != "mobile" && body.Category != "ai" && body.Category != "cloud" {
		config.Log.Warn("Geçersiz kategori tipi", zap.String("category", body.Category))
		c.JSON(http.StatusBadRequest, gin.H{"error": "geçersiz kategori ('backend', 'mobile', 'ai' veya 'cloud' olmalıdır)"})
		return
	}

	session := models.Session{
		Name:        body.Name,
		Description: body.Description,
		Type:        body.Type,
		Category:    body.Category,
		Points:      body.Points,
		QRCodeKey:   body.QRCodeKey,
	}

	if err := db.DB.Create(&session).Error; err != nil {
		config.Log.Error("Seans oluşturulurken hata", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "seans oluşturulamadı"})
		return
	}

	config.Log.Info("Yeni seans başarıyla oluşturuldu", zap.String("action", "create_session"), zap.Uint("sessionID", session.ID))
	c.JSON(http.StatusCreated, gin.H{
		"message":     "Seans başarıyla oluşturuldu",
		"session":     session,
		"qr_code_url": fmt.Sprintf("/api/admin/sessions/%d/qrcode", session.ID),
	})
}

func ListSessions(c *gin.Context) {
	var sessions []models.Session
	if err := db.DB.Find(&sessions).Error; err != nil {
		config.Log.Error("Seanslar listelenirken hata", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "seanslar yüklenemedi"})
		return
	}

	c.JSON(http.StatusOK, sessions)
}

func DeleteSession(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		config.Log.Warn("Geçersiz seans ID silme isteği", zap.String("idStr", idStr))
		c.JSON(http.StatusBadRequest, gin.H{"error": "geçersiz seans ID'si"})
		return
	}

	err = db.DB.Transaction(func(tx *gorm.DB) error {
		// Önce ilgili scan kayıtlarını sil
		if err := tx.Where("session_id = ?", id).Delete(&models.Scan{}).Error; err != nil {
			return err
		}
		result := tx.Delete(&models.Session{}, id)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return nil
	})

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "seans bulunamadı"})
			return
		}
		config.Log.Error("Seans silinirken hata", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "seans silinemedi"})
		return
	}

	config.Log.Info("Seans başarıyla silindi", zap.String("action", "delete_session"), zap.Int("sessionID", id))
	c.JSON(http.StatusOK, gin.H{"message": "Seans ve ilişkili taramalar başarıyla silindi"})
}

func GetSessionQRCode(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "geçersiz seans ID'si"})
		return
	}

	var session models.Session
	if err := db.DB.First(&session, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			config.Log.Warn("Karekod için seans bulunamadı", zap.Int("sessionID", id))
			c.JSON(http.StatusNotFound, gin.H{"error": "seans bulunamadı"})
			return
		}
		config.Log.Error("Seansa özel karekod üretilirken veritabanı hatası", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "seans yüklenemedi"})
		return
	}

	sizeStr := c.DefaultQuery("size", "256")
	size, err := strconv.Atoi(sizeStr)
	if err != nil || size <= 0 || size > 1024 {
		size = 256
	}

	encryptedKey, err := utils.Encrypt(session.QRCodeKey)
	if err != nil {
		config.Log.Error("Seans karekod anahtarı şifrelenirken kriptografik bir hata meydana geldi", zap.String("action", "encrypt_qr_key"), zap.Int("sessionID", id), zap.String("qrCodeKey", session.QRCodeKey), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "karekod şifrelenemedi"})
		return
	}

	//' Seansın şifreli QRCodeKey'i kullanılarak karekod görseli oluşturulur
	png, err := qrcode.Encode(encryptedKey, qrcode.Medium, size)
	if err != nil {
		config.Log.Error("Karekod png oluşturulamadı", zap.Error(err), zap.Int("sessionID", id))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "karekod görseli üretilemedi"})
		return
	}

	c.Data(http.StatusOK, "image/png", png)
}
