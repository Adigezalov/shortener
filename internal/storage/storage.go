package storage

import "github.com/Adigezalov/shortener/internal/models"

// URLStorage интерфейс для хранения URL
type URLStorage interface {
	// Add добавляет новый URL в хранилище
	// Возвращает ID, признак того, был ли URL уже в хранилище, и ошибку
	Add(id string, url string) (string, bool, error)

	// AddWithUser добавляет новый URL в хранилище с привязкой к пользователю
	// Возвращает ID, признак того, был ли URL уже в хранилище, и ошибку
	AddWithUser(id string, url string, userID string) (string, bool, error)

	// Get возвращает оригинальный URL по идентификатору
	Get(id string) (string, bool)

	// FindByOriginalURL ищет ID по оригинальному URL
	FindByOriginalURL(url string) (string, bool)

	// GetUserURLs возвращает все URL пользователя
	GetUserURLs(userID string) ([]models.UserURL, error)

	// Close закрывает хранилище и освобождает ресурсы
	Close() error
}
