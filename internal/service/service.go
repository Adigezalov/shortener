package service

import (
	"errors"
	"fmt"
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

func NewURLService(baseURL string, filePath string) *URLServiceImpl {
	storage := storage.NewMemoryStorage(filePath)

	return &URLServiceImpl{
		storage:  storage,
		baseURL:  baseURL,
		filePath: filePath,
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

	// Проверяем существование URL в памяти (данные уже загружены при старте)
	if shortID, exists := s.storage.Exists(originalURL); exists {
		return s.baseURL + shortID, ErrURLAlreadyExist
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

	// Загружаем текущие записи из файла
	records, _ := storage.LoadFromFile(s.filePath)

	// Добавляем новую запись
	records = append(records, storage.URLRecord{
		UUID:        shortID,
		ShortURL:    s.baseURL + shortID,
		OriginalURL: originalURL,
	})

	// Сохраняем обновленные записи обратно в файл
	if err := storage.SaveToFile(s.filePath, records); err != nil {
		return "", fmt.Errorf("failed to save records to file: %w", err)
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
