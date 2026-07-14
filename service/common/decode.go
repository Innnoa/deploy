package common

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"fmt"
)

func Decode(encryptedPassword string) string {
	decryptedPassword := ""
	key := "nEi4qx9ZoOCZSf9UFHwGL1WWaqt62yRWPDN/HsU24nk=" // 替换为Java端生成的Base64密钥

	// 解密密码
	decryptedPassword, err := decryptAESGCM(encryptedPassword, key)
	if err != nil {
		AppLogger.Error(fmt.Sprintln("Error decrypting:", err))

	}

	return decryptedPassword
}

func decryptAESGCM(encryptedText, key string) (string, error) {
	// Base64解码加密后的文本
	encryptedBytes, err := base64.StdEncoding.DecodeString(encryptedText)
	if err != nil {
		return "", err
	}

	// Base64解码密钥
	keyBytes, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return "", err
	}

	// 创建AES解密器
	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return "", err
	}

	// 创建解密模式
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// 解密数据
	nonceSize := gcm.NonceSize()
	if len(encryptedBytes) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}
	nonce, ciphertext := encryptedBytes[:nonceSize], encryptedBytes[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
