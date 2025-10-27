package storage

import "github.com/Adigezalov/shortener/internal/models"

// Stats представляет статистику хранилища
type Stats struct {
	URLs  int // Количество сокращённых URL в сервисе
	Users int // Количество пользователей в сервисе
}

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

	// DeleteUserURLs помечает URL как удаленные для указанного пользователя
	DeleteUserURLs(userID string, shortURLs []string) error

	// IsDeleted проверяет, помечен ли URL как удаленный
	IsDeleted(shortURL string) (bool, error)

	// Stats возвращает статистику хранилища
	Stats() (Stats, error)

	// Close закрывает хранилище и освобождает ресурсы
	Close() error
}
