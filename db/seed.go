package db

import (
	"dfrecap/config"
	"dfrecap/models"
	"os"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

func SeedAdminUser() {
	var count int64
	DB.Model(&models.Users{}).Count(&count)

	if count == 0 {
		config.Log.Info("Veritabanında hiç kullanıcı bulunamadığından varsayılan admin oluştuurluyor.")

		adminUser := os.Getenv("DEFAULT_ADMIN_USER")
		adminPass := os.Getenv("DEFAULT_ADMIN_PASS")
		adminMail := os.Getenv("DEFAULT_ADMIN_MAIL")

		if adminUser == "" || adminPass == "" || adminMail == "" {
			config.Log.Fatal("DEFAULT_ADMIN_USER, DEFAULT_ADMIN_PASS veya DEFAULT_ADMIN_MAIL tanımlı değil. Güvenlik sebebiyle sistem başlatılamıyor.")
			panic("Güvenlik ihlali: Varsayılan admin bilgileri ortam değişkenlerinde bulunamadı.")
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(adminPass), 10)
		if err != nil {
			config.Log.Fatal("varsayılan admin şifresi hashlenirken hata oluştu", zap.Error(err))
		}

		admin := models.Users{
			Username: adminUser,
			Password: string(hash),
			Mail:     adminMail,
			Role:     "admin",
		}

		result := DB.Create(&admin)
		if result.Error != nil {
			config.Log.Fatal("varsayılan admin oluşturulurken hata oluştu", zap.Error(err))
		}

		config.Log.Info("Varsayılan admin oluşturuldu", zap.String("username", adminUser))
	}

	config.Log.Info("Veritabanında kullanıcı mevcut, seed işlemi atlandı.")
}
