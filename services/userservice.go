package services

import (
	"dfrecap/config"
	"dfrecap/db"
	"dfrecap/models"
	"errors"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

func Register(username, fullName, mail, password string) (*models.Users, error) {
	var count int64
	db.DB.Model(&models.Users{}).Where("username = ?", username).Count(&count)
	if count > 0 {
		return nil, errors.New("kullanıcı adı zaten kullanımda")
	}

	db.DB.Model(&models.Users{}).Where("mail = ?", mail).Count(&count)
	if count > 0 {
		return nil, errors.New("e-posta adresi zaten kullanımda")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		config.Log.Error("Şifre hashlenirken hata oluştu", zap.Error(err))
		return nil, errors.New("kayıt işlemi sırasında bir hata oluştu")
	}

	user := &models.Users{
		Username: username,
		FullName: fullName,
		Password: string(hash),
		Mail:     mail,
		Role:     "user",
	}

	if err := db.DB.Create(user).Error; err != nil {
		config.Log.Error("Kullanıcı kaydedilemedi", zap.Error(err))
		return nil, errors.New("veritabanına kayıt yapılamadı")
	}

	config.Log.Info("Yeni katılımcı başarıyla kaydoldu", zap.String("username", username), zap.String("mail", mail))
	return user, nil
}
