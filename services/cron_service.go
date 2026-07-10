package services

import (
	"dfrecap/config"
	"os"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

var Cron *cron.Cron

func StartCronJobs() {
	if os.Getenv("DEVTV_DSN") == "" {
		config.Log.Warn("DEVTV_DSN bulunamadı, DevTV arka plan senkronizasyonu pasif durumda")
		return
	}

	Cron = cron.New(cron.WithSeconds())

	config.Log.Info("DevTV arka plan senkronizasyonu aktif edildi (30 saniyede bir çalışacak)")

	go runDevTVSync()

	_, err := Cron.AddFunc("*/30 * * * * *", runDevTVSync)
	if err != nil {
		config.Log.Error("DevTV cron job eklenemedi", zap.Error(err))
		return
	}

	Cron.Start()
}

func StopCronJobs() {
	if Cron != nil {
		Cron.Stop()
		config.Log.Info("Cron görevleri durduruldu")
	}
}

func runDevTVSync() {
	defer func() {
		if r := recover(); r != nil {
			config.Log.Error("DevTV sync job beklenmeyen bir hata ile sonlandı", zap.Any("panic", r))
		}
	}()

	ins, upd, fet, err := SyncFromDevTV()
	if err != nil {
		config.Log.Error("DevTV arka plan senkronizasyon hatası", zap.Error(err))
	} else if ins > 0 || upd > 0 || fet > 0 {
		config.Log.Info("DevTV arka plan senkronizasyonu tamamlandı", zap.Int("eklenen", ins), zap.Int("güncellenen", upd), zap.Int("getirilen", fet))
	}
}
