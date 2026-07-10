package db

import (
	"dfrecap/config"
	"os"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect(dbconf config.DatabaseConfig, envPath string) {
	err := godotenv.Load(envPath)
	if err != nil && !os.IsNotExist(err) {
		config.Log.Warn(".env dosyası bulunamadı, sistem ortam değişkenleri kullanılacak", zap.Error(err))
	}
	dsn := os.Getenv("dsn")
	if dsn == "" {
		config.Log.Fatal("DSN ortam değişkeni tanımlı değil, uygulama başlatılamaz")
	}

	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		config.Log.Fatal("Database başlatılırken hata oluştu, ", zap.Error(err))
		return
	}

	sqlDB, err := DB.DB()
	if err != nil {
		config.Log.Error("sql.DB alınamadı, ", zap.Error(err))
		return
	}
	sqlDB.SetMaxIdleConns(dbconf.MaxIdleConns)
	sqlDB.SetMaxOpenConns(dbconf.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(dbconf.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(dbconf.ConnMaxIdleTime)

	config.Log.Debug("Connection Pooling ayarları uygulandı.")
	config.Log.Info("Database bağlantısı başarılı.")
}
