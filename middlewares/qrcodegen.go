package middlewares

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/skip2/go-qrcode"
)

func GenerateQRCode(c *gin.Context) {
	text := c.Query("text")
	if text == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "text parametresi zorunludur"})
		return
	}

	sizeStr := c.DefaultQuery("size", "256")
	size, err := strconv.Atoi(sizeStr)
	if err != nil || size <= 0 || size > 1024 {
		size = 256
	}

	//' Gelişmiş orta düzey hata toleransı ile karekodu üretir
	png, err := qrcode.Encode(text, qrcode.Medium, size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "karekod üretilemedi"})
		return
	}

	c.Data(http.StatusOK, "image/png", png)
}
