package handlers

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
