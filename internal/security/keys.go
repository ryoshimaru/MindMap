package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"strings"
)

func EncryptAPIKey(value string) (string, error) {
	gcm, err := keyCipher()
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	sealed := gcm.Seal(nonce, nonce, []byte(strings.TrimSpace(value)), nil)
	return base64.RawURLEncoding.EncodeToString(sealed), nil
}

func DecryptAPIKey(value string) (string, error) {
	gcm, err := keyCipher()
	if err != nil {
		return "", err
	}
	payload, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil || len(payload) < gcm.NonceSize() {
		return "", fmt.Errorf("invalid encrypted API key")
	}
	nonce, ciphertext := payload[:gcm.NonceSize()], payload[gcm.NonceSize():]
	plain, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}

func keyCipher() (cipher.AEAD, error) {
	secret := strings.TrimSpace(os.Getenv("AI_KEY_ENCRYPTION_KEY"))
	if secret == "" {
		return nil, fmt.Errorf("AI_KEY_ENCRYPTION_KEY is required")
	}
	hash := sha256.Sum256([]byte(secret))
	block, err := aes.NewCipher(hash[:])
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}
