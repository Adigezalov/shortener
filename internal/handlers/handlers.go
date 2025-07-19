package handlers

import "github.com/Adigezalov/shortener/internal/models"

// URLStorage интерфейс для хранения URL
type URLStorage interface {
	Add(id string, url string) (string, bool, error)
	AddWithUser(id string, url string, userID string) (string, bool, error)
	Get(id string) (string, bool)
	FindByOriginalURL(url string) (string, bool)
	GetUserURLs(userID string) ([]models.UserURL, error)
	DeleteUserURLs(userID string, shortURLs []string) error
	IsDeleted(shortURL string) (bool, error)
	Close() error
}

// URLShortener интерфейс для сокращения URL
type URLShortener interface {
	Shorten(url string) string
	BuildShortURL(id string) string
}

// Pinger интерфейс для проверки подключения к базе данных
type Pinger interface {
	Ping() error
}

// Handler обработчик HTTP запросов
type Handler struct {
	storage   URLStorage
	shortener URLShortener
	db        Pinger
}

// New создает новый обработчик HTTP запросов
func New(storage URLStorage, shortener URLShortener, db Pinger) *Handler {
	return &Handler{
		storage:   storage,
		shortener: shortener,
		db:        db,
	}
}
