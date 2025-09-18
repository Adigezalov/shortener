// Package shortener предоставляет сервис для сокращения URL.
//
// Пакет реализует алгоритм генерации коротких идентификаторов
// на основе криптографически стойкого генератора случайных чисел
// и base64 кодирования.
package shortener

import (
	"crypto/rand"
	"encoding/base64"
	"strings"
)

// Service предоставляет функциональность сокращения URL.
//
// Сервис генерирует короткие идентификаторы длиной 8 символов
// используя URL-safe base64 кодирование. Каждый идентификатор
// статистически уникален благодаря использованию crypto/rand.
//
// Пример использования:
//
//	service := shortener.New("https://example.com")
//	shortID := service.Shorten("https://very-long-url.com/path")
//	fullURL := service.BuildShortURL(shortID)
//	// fullURL = "https://example.com/abc12345"
type Service struct {
	baseURL string // Базовый URL для формирования коротких ссылок
}

// New создает новый экземпляр сервиса сокращения URL.
//
// Параметр baseURL должен содержать полный базовый адрес
// включая протокол (http:// или https://) без завершающего слеша.
//
// Пример:
//
//	service := shortener.New("https://short.ly")
//	service := shortener.New("http://localhost:8080")
//
// Возвращает готовый к использованию сервис.
func New(baseURL string) *Service {
	return &Service{
		baseURL: baseURL,
	}
}

// Shorten генерирует уникальный короткий идентификатор для URL.
//
// Метод создает 8-символьный идентификатор используя:
//  1. 6 случайных байт из crypto/rand
//  2. URL-safe base64 кодирование (без символов '/' и '+')
//  3. Обрезание до 8 символов
//
// Параметр url в данной реализации не используется для генерации ID,
// что обеспечивает полную случайность идентификаторов.
//
// Возвращает короткий идентификатор, например: "abc12345"
//
// Пример:
//
//	id := service.Shorten("https://example.com/very/long/path")
//	// id = "kJ8xN2mP" (случайный 8-символьный ID)
func (s *Service) Shorten(url string) string {
	// Генерируем случайную последовательность байт
	b := make([]byte, 6)
	rand.Read(b)

	// Кодируем в base64 и берем первые 8 символов
	// Используем URL-safe версию base64 (без '/' и '+')
	encoded := base64.RawURLEncoding.EncodeToString(b)
	return encoded[:8]
}

// BuildShortURL строит полный короткий URL из идентификатора.
//
// Метод объединяет базовый URL сервиса с переданным идентификатором,
// формируя полную ссылку для использования в HTTP ответах.
//
// Использует strings.Builder с предварительным выделением памяти
// для оптимальной производительности.
//
// Параметр id должен содержать короткий идентификатор, полученный
// от метода Shorten.
//
// Возвращает полный URL вида "https://example.com/abc12345"
//
// Пример:
//
//	fullURL := service.BuildShortURL("abc12345")
//	// fullURL = "https://example.com/abc12345"
func (s *Service) BuildShortURL(id string) string {
	var builder strings.Builder
	builder.Grow(len(s.baseURL) + 1 + len(id)) // Предварительно выделяем память
	builder.WriteString(s.baseURL)
	builder.WriteByte('/')
	builder.WriteString(id)
	return builder.String()
}
