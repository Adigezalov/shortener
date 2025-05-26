package shortener

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

// Service сервис для сокращения URL
type Service struct {
	baseURL string
}

// New создает новый сервис сокращения URL
func New(baseURL string) *Service {
	return &Service{
		baseURL: baseURL,
	}
}

// Shorten генерирует короткий идентификатор для URL
func (s *Service) Shorten(url string) string {
	// Генерируем случайную последовательность байт
	b := make([]byte, 6)
	rand.Read(b)

	// Кодируем в base64 и берем первые 8 символов
	// Используем URL-safe версию base64 (без '/' и '+')
	encoded := base64.RawURLEncoding.EncodeToString(b)
	return encoded[:8]
}

// BuildShortURL строит полный короткий URL из идентификатора
func (s *Service) BuildShortURL(id string) string {
	return fmt.Sprintf("%s/%s", s.baseURL, id)
}
