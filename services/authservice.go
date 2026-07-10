package services

import (
	"dfrecap/config"
	"dfrecap/db"
	"dfrecap/middlewares"
	"dfrecap/models"
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

func Login(username, password string, jwtCfg config.JWTConfig) (string, error) {
	var user models.Users
	if err := db.DB.Where("username = ?", username).First(&user).Error; err != nil {
		config.Log.Warn("Kullanıcı bulunamadı", zap.String("username", username))
		return "", errors.New("geçersiz kullanıcı adı veya şifre")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", errors.New("geçersiz kullanıcı adı veya şifre")
	}


	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return "", errors.New("JWT_SECRET tanımlı değil")
	}

	claims := middlewares.JWTClaims{
		ID:       int(user.ID),
		Username: user.Username,
		Mail:     user.Mail,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(jwtCfg.Expiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		config.Log.Error("Token oluşturulamadı", zap.Error(err))
		return "", err
	}

	return signed, nil
}
