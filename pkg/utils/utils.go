package utils

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
)

const ShortIDLength = 8

// GenerateShortID создаёт случайный короткий ID длиной ShortIDLength символов
func GenerateShortID() (string, error) {
	b := make([]byte, ShortIDLength)

	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(b)[:ShortIDLength], nil
}

// GetBaseURL определяет базовый URL
func GetBaseURL(req *http.Request) string {
	scheme := "http"

	if req.TLS != nil || req.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}

	return fmt.Sprintf("%s://%s/", scheme, req.Host)
}
