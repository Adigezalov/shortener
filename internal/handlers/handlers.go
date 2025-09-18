// Package handlers предоставляет HTTP обработчики для сервиса сокращения URL.
//
// Пакет содержит обработчики для всех эндпоинтов API:
//   - Создание коротких URL (POST /, POST /api/shorten)
//   - Пакетное создание URL (POST /api/shorten/batch)
//   - Редирект по короткому URL (GET /{id})
//   - Получение URL пользователя (GET /api/user/urls)
//   - Удаление URL пользователя (DELETE /api/user/urls)
//   - Проверка состояния БД (GET /ping)
package handlers

import "github.com/Adigezalov/shortener/internal/models"

// URLStorage определяет интерфейс для хранения и управления URL.
//
// Интерфейс поддерживает различные типы хранилищ: в памяти, файловое и базу данных.
// Все методы должны быть потокобезопасными.
type URLStorage interface {
	// Add добавляет новый URL в хранилище.
	// Возвращает ID, флаг существования URL и ошибку.
	Add(id string, url string) (string, bool, error)

	// AddWithUser добавляет URL с привязкой к пользователю.
	// Возвращает ID, флаг существования URL и ошибку.
	AddWithUser(id string, url string, userID string) (string, bool, error)

	// Get возвращает оригинальный URL по короткому ID.
	// Возвращает URL и флаг существования.
	Get(id string) (string, bool)

	// FindByOriginalURL ищет короткий ID по оригинальному URL.
	// Возвращает ID и флаг существования.
	FindByOriginalURL(url string) (string, bool)

	// GetUserURLs возвращает все URL пользователя.
	GetUserURLs(userID string) ([]models.UserURL, error)

	// DeleteUserURLs помечает URL как удаленные для указанного пользователя.
	DeleteUserURLs(userID string, shortURLs []string) error

	// IsDeleted проверяет, помечен ли URL как удаленный.
	IsDeleted(shortURL string) (bool, error)

	// Close закрывает хранилище и освобождает ресурсы.
	Close() error
}

// URLShortener определяет интерфейс для сокращения URL.
//
// Интерфейс предоставляет методы для генерации коротких идентификаторов
// и построения полных коротких URL.
type URLShortener interface {
	// Shorten генерирует короткий идентификатор для URL.
	// Возвращает уникальный короткий ID.
	Shorten(url string) string

	// BuildShortURL строит полный короткий URL из идентификатора.
	// Возвращает полный URL вида "http://example.com/abc123".
	BuildShortURL(id string) string
}

// Pinger определяет интерфейс для проверки подключения к базе данных.
//
// Используется для реализации health check эндпоинта.
type Pinger interface {
	// Ping проверяет доступность базы данных.
	// Возвращает ошибку, если подключение недоступно.
	Ping() error
}

// Handler содержит обработчики HTTP запросов для сервиса сокращения URL.
//
// Структура инкапсулирует зависимости: хранилище URL, сервис сокращения
// и интерфейс для проверки БД. Все обработчики являются методами этой структуры.
type Handler struct {
	storage   URLStorage
	shortener URLShortener
	db        Pinger
}

// New создает новый экземпляр обработчика HTTP запросов.
//
// Параметры:
//   - storage: реализация интерфейса URLStorage для хранения URL
//   - shortener: реализация интерфейса URLShortener для сокращения URL
//   - db: реализация интерфейса Pinger для проверки БД (может быть nil)
//
// Возвращает готовый к использованию Handler.
//
// Пример использования:
//
//	storage := memory.NewMemoryStorage("")
//	shortener := shortener.New("http://localhost:8080")
//	handler := handlers.New(storage, shortener, nil)
func New(storage URLStorage, shortener URLShortener, db Pinger) *Handler {
	return &Handler{
		storage:   storage,
		shortener: shortener,
		db:        db,
	}
}
