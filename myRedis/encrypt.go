package myRedis

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
)

// encrypt はデータをAES-GCMで暗号化します
func encrypt(plainText []byte, secretKey []byte) ([]byte, error) {
	block, err := aes.NewCipher(secretKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// 暗号化ごとにランダムなNonce（一時的な値）を生成
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	// 暗号化データの先頭にNonceを付与して返す
	return gcm.Seal(nonce, nonce, plainText, nil), nil
}

// decrypt は暗号化されたデータを復号します
func decrypt(cipherText []byte, secretKey []byte) ([]byte, error) {
	block, err := aes.NewCipher(secretKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(cipherText) < nonceSize {
		return nil, fmt.Errorf("cipherText too short")
	}

	// 先頭のNonce部分と、それ以降の暗号文を分解
	nonce, actualCipherText := cipherText[:nonceSize], cipherText[nonceSize:]

	// 復号
	plainText, err := gcm.Open(nil, nonce, actualCipherText, nil)
	if err != nil {
		return nil, err
	}

	return plainText, nil
}

func GenerateSecureToken(length int) string {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}
