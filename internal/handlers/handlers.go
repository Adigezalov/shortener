package handlers

// URLStorage интерфейс для хранения URL
type URLStorage interface {
	Add(id string, url string) (string, bool)
	Get(id string) (string, bool)
	FindByOriginalURL(url string) (string, bool)
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
}

// New создает новый обработчик HTTP запросов
func New(storage URLStorage, shortener URLShortener) *Handler {
	return &Handler{
		storage:   storage,
		shortener: shortener,
	}
}
