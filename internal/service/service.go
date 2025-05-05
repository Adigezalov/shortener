package service

import (
	"errors"
	"github.com/Adigezalov/shortener/internal/storage"
	"github.com/Adigezalov/shortener/pkg/utils"
	"net/http"
	"net/url"
	"strings"
)

var (
	ErrInvalidURL      = errors.New("invalid URL")
	ErrEmptyURL        = errors.New("empty URL")
	ErrURLAlreadyExist = errors.New("URL already exists")
)

// URLServiceImpl - реализация URLService
type URLServiceImpl struct {
	storage storage.URLReadWriter
	baseURL string
}

func NewURLService(storage storage.URLReadWriter, baseURL string) *URLServiceImpl {
	return &URLServiceImpl{
		storage: storage,
		baseURL: baseURL,
	}
}

// SetBaseURL устанавливает базовый URL для сервиса (должен вызываться перед использованием)
func (s *URLServiceImpl) SetBaseURL(req *http.Request) {
	s.baseURL = utils.GetBaseURL(req)
}

// ShortenURL создает короткую версию URL
func (s *URLServiceImpl) ShortenURL(originalURL string) (string, error) {
	originalURL = strings.TrimSpace(originalURL)
	if originalURL == "" {
		return "", ErrEmptyURL
	}

	// Проверяем валидность URL
	if _, err := url.ParseRequestURI(originalURL); err != nil {
		return "", ErrInvalidURL
	}

	// Проверяем, не сокращали ли уже этот URL
	if shortID, exists := s.storage.Exists(originalURL); exists {
		return s.baseURL + shortID, ErrURLAlreadyExist
	}

	// Генерируем новый короткий ID
	shortID, err := utils.GenerateShortID()
	if err != nil {
		return "", err
	}

	// Сохраняем в хранилище
	if err := s.storage.Save(shortID, originalURL); err != nil {
		return "", err
	}

	return s.baseURL + shortID, nil
}

// GetOriginalURL возвращает оригинальный URL по короткому идентификатору
func (s *URLServiceImpl) GetOriginalURL(shortID string) (string, error) {
	if shortID == "" {
		return "", ErrInvalidURL
	}

	originalURL, exists := s.storage.Get(shortID)

	if !exists {
		return "", ErrInvalidURL
	}

	return originalURL, nil
}

// GetShortURLIfExists возвращает короткий URL если оригинальный уже существует
func (s *URLServiceImpl) GetShortURLIfExists(originalURL string) (string, bool) {
	originalURL = strings.TrimSpace(originalURL)
	if originalURL == "" {
		return "", false
	}

	if shortID, exists := s.storage.Exists(originalURL); exists {
		return s.baseURL + shortID, true
	}

	return "", false
}
