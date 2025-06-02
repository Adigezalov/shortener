package handlers

import (
	"github.com/Adigezalov/shortener/internal/database"
)

// URLStorage интерфейс для хранения URL
type URLStorage interface {
	Add(id string, url string) (string, bool, error)
	Get(id string) (string, bool)
	FindByOriginalURL(url string) (string, bool)
	Close() error
}

// URLShortener интерфейс для сокращения URL
type URLShortener interface {
	Shorten(url string) string
	BuildShortURL(id string) string
}

// Handler обработчик HTTP запросов
type Handler struct {
	storage   URLStorage
	shortener URLShortener
	db        database.DBInterface
}

// New создает новый обработчик HTTP запросов
func New(storage URLStorage, shortener URLShortener, db database.DBInterface) *Handler {
	return &Handler{
		storage:   storage,
		shortener: shortener,
		db:        db,
	}
}
