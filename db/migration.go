package db

import (
	"dfrecap/config"
	"dfrecap/models"
)

func RunMigrations() {
	if DB == nil {
		config.Log.Fatal("db bağlantısı yok")
	}
	DB.AutoMigrate(
		&models.Users{},
		&models.Session{},
		&models.Scan{},
	)
	config.Log.Info("Tablolar migrate edildi")
}
