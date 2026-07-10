package services

import (
	"context"
	"dfrecap/config"
	"dfrecap/db"
	"dfrecap/models"
	"dfrecap/utils"
	"errors"
	"fmt"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

func ScanQRCode(userID uint, qrCodeKey string) (*models.Scan, error) {
	decryptedKey, err := utils.Decrypt(qrCodeKey)
	if err != nil {
		config.Log.Warn("Geçersiz veya manipüle edilmiş QR okutma girişimi",
			zap.String("action", "scan_rejected"),
			zap.Uint("userID", userID),
			zap.String("rawPayload", qrCodeKey),
			zap.Error(err),
		)
		return nil, errors.New("geçersiz veya okunamayan QR formatı")
	}

	config.Log.Info("QR kod payload'u başarıyla deşifre edildi",
		zap.String("action", "decrypt_qr"),
		zap.Uint("userID", userID),
	)
	qrCodeKey = decryptedKey

	var session models.Session
	if err := db.DB.Where("qr_code_key = ?", qrCodeKey).First(&session).Error; err != nil {
		config.Log.Warn("Geçersiz QR Kod okutuldu", zap.String("qrCodeKey", qrCodeKey), zap.Uint("userID", userID))
		return nil, errors.New("geçersiz veya bulunamayan seans QR kodu")
	}

	scan := &models.Scan{
		UserID:    userID,
		SessionID: session.ID,
	}

	err = db.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(scan).Error; err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				config.Log.Warn("Çifte tarama girişimi", zap.Uint("userID", userID), zap.Uint("sessionID", session.ID))
				return errors.New("bu seansı daha önce zaten okuttunuz")
			}
			config.Log.Error("Tarama kaydı veritabanına eklenemedi", zap.Error(err), zap.Uint("userID", userID), zap.Uint("sessionID", session.ID))
			return errors.New("tarama işlemi kaydedilirken hata oluştu")
		}

		if db.RDB != nil {
			ctx := context.Background()
			err := db.RDB.ZIncrBy(ctx, "leaderboard", float64(session.Points), fmt.Sprintf("%d", userID)).Err()
			if err != nil {
				config.Log.Error("Redis leaderboard güncellenemedi, işlem geri alınıyor", zap.Error(err), zap.Uint("userID", userID))
				return errors.New("puanınız eklenirken sunucu kaynaklı bir hata oluştu, lütfen tekrar deneyin")
			}
			db.RDB.Del(ctx, fmt.Sprintf("recap:%d", userID))
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	scan.Session = session

	config.Log.Info("QR Kod tarama başarılı", zap.Uint("userID", userID), zap.Uint("sessionID", session.ID), zap.Int("points", session.Points))
	return scan, nil
}
