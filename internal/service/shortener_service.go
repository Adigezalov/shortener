// Package service содержит бизнес-логику приложения.
// Этот слой используется как HTTP, так и gRPC хендлерами.
package service

import (
	"github.com/Adigezalov/shortener/internal/database"
	"github.com/Adigezalov/shortener/internal/models"
	"github.com/Adigezalov/shortener/internal/storage"
)

// URLStorage определяет интерфейс для хранения и управления URL.
type URLStorage interface {
	Add(id string, url string) (string, bool, error)
	AddWithUser(id string, url string, userID string) (string, bool, error)
	Get(id string) (string, bool)
	FindByOriginalURL(url string) (string, bool)
	GetUserURLs(userID string) ([]models.UserURL, error)
	DeleteUserURLs(userID string, shortURLs []string) error
	IsDeleted(shortURL string) (bool, error)
	Stats() (storage.Stats, error)
	Close() error
}

// URLShortener определяет интерфейс для сокращения URL.
type URLShortener interface {
	Shorten(url string) string
	BuildShortURL(id string) string
}

// Pinger определяет интерфейс для проверки подключения к базе данных.
type Pinger interface {
	Ping() error
}

// ShortenerService содержит бизнес-логику для работы с URL.
type ShortenerService struct {
	storage   URLStorage
	shortener URLShortener
	db        Pinger
}

// NewShortenerService создает новый экземпляр сервиса.
func NewShortenerService(storage URLStorage, shortener URLShortener, db Pinger) *ShortenerService {
	return &ShortenerService{
		storage:   storage,
		shortener: shortener,
		db:        db,
	}
}

// CreateShortURLResult содержит результат создания короткого URL.
type CreateShortURLResult struct {
	ShortURL string
	Exists   bool
	Error    error
}

// CreateShortURL создает короткий URL для указанного оригинального URL.
func (s *ShortenerService) CreateShortURL(url string, userID string) CreateShortURLResult {
	if url == "" {
		return CreateShortURLResult{Error: ErrEmptyURL}
	}

	// Генерируем ID
	id := s.shortener.Shorten(url)

	// Добавляем URL с привязкой к пользователю
	id, exists, err := s.storage.AddWithUser(id, url, userID)
	if err != nil && err != database.ErrURLConflict {
		return CreateShortURLResult{Error: err}
	}

	// Строим полный короткий URL
	shortURL := s.shortener.BuildShortURL(id)

	return CreateShortURLResult{
		ShortURL: shortURL,
		Exists:   exists || err == database.ErrURLConflict,
		Error:    nil,
	}
}

// BatchItem представляет элемент пакетного запроса.
type BatchItem struct {
	CorrelationID string
	OriginalURL   string
}

// BatchResult представляет результат пакетного создания URL.
type BatchResult struct {
	CorrelationID string
	ShortURL      string
}

// CreateShortURLBatch создает короткие URL для списка оригинальных URL.
func (s *ShortenerService) CreateShortURLBatch(items []BatchItem, userID string) []BatchResult {
	results := make([]BatchResult, 0, len(items))

	for _, item := range items {
		if item.OriginalURL == "" {
			continue
		}

		// Генерируем ID
		id := s.shortener.Shorten(item.OriginalURL)

		// Добавляем URL с привязкой к пользователю
		id, _, err := s.storage.AddWithUser(id, item.OriginalURL, userID)
		if err != nil && err != database.ErrURLConflict {
			continue
		}

		// Строим полный короткий URL
		shortURL := s.shortener.BuildShortURL(id)

		results = append(results, BatchResult{
			CorrelationID: item.CorrelationID,
			ShortURL:      shortURL,
		})
	}

	return results
}

// GetOriginalURLResult содержит результат получения оригинального URL.
type GetOriginalURLResult struct {
	OriginalURL string
	Deleted     bool
	Found       bool
	Error       error
}

// GetOriginalURL возвращает оригинальный URL по короткому ID.
func (s *ShortenerService) GetOriginalURL(id string) GetOriginalURLResult {
	// Получаем оригинальный URL
	originalURL, found := s.storage.Get(id)
	if !found {
		return GetOriginalURLResult{Found: false}
	}

	// Проверяем, не удален ли URL
	deleted, err := s.storage.IsDeleted(id)
	if err != nil {
		return GetOriginalURLResult{Error: err}
	}

	return GetOriginalURLResult{
		OriginalURL: originalURL,
		Deleted:     deleted,
		Found:       true,
		Error:       nil,
	}
}

// GetUserURLsResult содержит результат получения URL пользователя.
type GetUserURLsResult struct {
	URLs  []models.UserURL
	Error error
}

// GetUserURLs возвращает все URL пользователя.
func (s *ShortenerService) GetUserURLs(userID string) GetUserURLsResult {
	userURLs, err := s.storage.GetUserURLs(userID)
	if err != nil {
		return GetUserURLsResult{Error: err}
	}

	// Преобразуем URL в полные ссылки
	result := make([]models.UserURL, len(userURLs))
	for i, userURL := range userURLs {
		result[i] = models.UserURL{
			ShortURL:    s.shortener.BuildShortURL(userURL.ShortURL),
			OriginalURL: userURL.OriginalURL,
		}
	}

	return GetUserURLsResult{
		URLs:  result,
		Error: nil,
	}
}

// DeleteUserURLs помечает URL пользователя как удаленные.
func (s *ShortenerService) DeleteUserURLs(userID string, shortURLs []string) error {
	if len(shortURLs) == 0 {
		return ErrEmptyList
	}

	return s.storage.DeleteUserURLs(userID, shortURLs)
}

// PingDB проверяет доступность базы данных.
func (s *ShortenerService) PingDB() error {
	if s.db == nil {
		return ErrDBNotConfigured
	}
	return s.db.Ping()
}

// StatsResult содержит статистику сервиса.
type StatsResult struct {
	URLs  int
	Users int
	Error error
}

// GetStats возвращает статистику сервиса.
func (s *ShortenerService) GetStats() StatsResult {
	stats, err := s.storage.Stats()
	if err != nil {
		return StatsResult{Error: err}
	}

	return StatsResult{
		URLs:  stats.URLs,
		Users: stats.Users,
		Error: nil,
	}
}
