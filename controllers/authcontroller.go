package controllers

import (
	"dfrecap/config"
	"dfrecap/services"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/microcosm-cc/bluemonday"
	"go.uber.org/zap"
)

var jwtConfig config.JWTConfig

func SetJWTConfig(cfg config.JWTConfig) {
	jwtConfig = cfg
}

func Login(c *gin.Context) {
	var body struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		config.Log.Warn("Giriş verisi doğrulanamadı", zap.String("action", "login"), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "geçersiz parametreler"})
		return
	}
	token, err := services.Login(body.Username, body.Password, jwtConfig)
	if err != nil {
		config.Log.Warn("Giriş başarısız", zap.String("action", "login"), zap.String("username", body.Username), zap.Error(err))
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	config.Log.Info("Kullanıcı girişi başarılı", zap.String("action", "login"), zap.String("username", body.Username))
	c.SetCookie("Auth", token, 60*60*24*7, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"message": "Giriş başarılı", "token": token})
}

func Register(c *gin.Context) {
	var body struct {
		Username string `json:"username" binding:"required"`
		FullName string `json:"full_name" binding:"required"`
		Mail     string `json:"mail" binding:"required,email"`
		Password string `json:"password" binding:"required,min=6"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		config.Log.Warn("Kayıt verisi doğrulanamadı", zap.String("action", "register"), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	p := bluemonday.StrictPolicy()
	sanitizedFullName := p.Sanitize(body.FullName)
	sanitizedUsername := p.Sanitize(body.Username)

	user, err := services.Register(sanitizedUsername, sanitizedFullName, body.Mail, body.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	config.Log.Info("Yeni kullanıcı kaydı başarılı", zap.String("action", "register"), zap.Uint("userID", user.ID), zap.String("username", body.Username))
	c.JSON(http.StatusCreated, gin.H{"message": "Kayıt başarılı", "user_id": user.ID})
}

func Logout(c *gin.Context) {
	c.SetCookie("Auth", "", -1, "/", "", true, true)
	config.Log.Info("Kullanıcı çıkışı yapıldı", zap.String("action", "logout"))
	c.JSON(http.StatusOK, gin.H{"message": "Çıkış başarılı"})
}

//! FIX: (Security) Login fonksiyonundaki c.SetCookie işleminde 'secure' parametresi production ortamında 'true' yapılarak çerezin sadece HTTPS üzerinden taşınması sağlanmalıdır.
