package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"sync"
)

var (
	encryptionKey []byte
	keyOnce       sync.Once
)

func getEncryptionKey() []byte {
	keyOnce.Do(func() {
		secret := os.Getenv("QR_SECRET")
		if secret == "" {
			panic("qr secret env yok")
		}
		hash := sha256.Sum256([]byte(secret))
		encryptionKey = hash[:]
	})
	return encryptionKey
}

// ' Verilen düz metni AES-GCM yöntemiyle şifreleyerek hex formatında döner
func Encrypt(plainText string) (string, error) {
	key := getEncryptionKey()
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	cipherText := aesGCM.Seal(nonce, nonce, []byte(plainText), nil)
	return hex.EncodeToString(cipherText), nil
}

// ' Hex formatındaki şifreli metni çözerek düz metin haline getirir
func Decrypt(cipherTextHex string) (string, error) {
	key := getEncryptionKey()
	cipherText, err := hex.DecodeString(cipherTextHex)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := aesGCM.NonceSize()
	if len(cipherText) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, actualCipherText := cipherText[:nonceSize], cipherText[nonceSize:]
	plainTextBytes, err := aesGCM.Open(nil, nonce, actualCipherText, nil)
	if err != nil {
		return "", err
	}

	return string(plainTextBytes), nil
}
