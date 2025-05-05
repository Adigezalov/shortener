package app

import "net/http"

// URLService определяет контракт для сервиса сокращения URL
type URLService interface {
	SetBaseURL(req *http.Request)
	ShortenURL(originalURL string) (string, error)
	GetOriginalURL(shortID string) (string, error)
	GetShortURLIfExists(originalURL string) (string, bool)
}
