package utils

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log"
)

func HashAPIKey(key string) string {
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:])
}

func GenerateHMAC(message, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(message))
	return hex.EncodeToString(h.Sum(nil))
}

func GenerateAPIKey() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		log.Println("Cannot generate the API Key")
		return "", fmt.Errorf("generate api key: %w", err)
	}
	return "sk_" + hex.EncodeToString(b), nil
}

func VerifyHMAC(message, providedSignature, secret string) bool {
	expectedMAC := hmac.New(sha256.New, []byte(secret))
	expectedMAC.Write([]byte(message))
	expected := expectedMAC.Sum(nil)

	provided, err := hex.DecodeString(providedSignature)
	if err != nil {
		return false
	}

	return hmac.Equal(expected, provided) // constant-time comparison
}

const SecretLength = 32

func GenerateSecret() (string, error) {
	bytes := make([]byte, SecretLength)

	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	encoded := base64.RawURLEncoding.EncodeToString(bytes)
	return "whsec" + encoded, nil // "whsec" -> webhook secret prefix
}
