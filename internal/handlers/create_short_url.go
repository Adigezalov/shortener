package handlers

import (
	"github.com/Adigezalov/shortener/internal/logger"
	"go.uber.org/zap"
	"io"
	"net/http"
)

// CreateShortURL обрабатывает POST запрос на создание сокращенного URL (text/plain)
func (h *Handler) CreateShortURL(w http.ResponseWriter, r *http.Request) {
	// Читаем оригинальный URL из тела запроса
	body, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Logger.Error("Ошибка чтения тела запроса", zap.Error(err))
		http.Error(w, "Ошибка чтения запроса", http.StatusBadRequest)
		return
	}

	originalURL := string(body)
	if originalURL == "" {
		http.Error(w, "URL не может быть пустым", http.StatusBadRequest)
		return
	}

	// Создаем или получаем существующий короткий ID
	var id string
	var exists bool

	// Сначала проверяем, есть ли такой URL в хранилище
	id, exists = h.storage.FindByOriginalURL(originalURL)
	if !exists {
		// Если URL не найден, генерируем новый ID
		id = h.shortener.Shorten(originalURL)
		// Добавляем URL в хранилище
		id, _ = h.storage.Add(id, originalURL)
	}

	// Строим полный короткий URL
	shortURL := h.shortener.BuildShortURL(id)

	// Отправляем результат
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shortURL))

	logger.Logger.Info("URL сокращен",
		zap.String("original_url", originalURL),
		zap.String("short_url", shortURL),
		zap.Bool("existing", exists),
	)
}
