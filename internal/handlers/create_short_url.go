package handlers

import (
	"errors"
	"github.com/Adigezalov/shortener/internal/database"
	"github.com/Adigezalov/shortener/internal/logger"
	"github.com/Adigezalov/shortener/internal/middleware"
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

	// Получаем ID пользователя из контекста
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		logger.Logger.Error("Не удалось получить ID пользователя из контекста")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Генерируем новый ID и пытаемся добавить URL с привязкой к пользователю
	id := h.shortener.Shorten(originalURL)
	id, exists, err := h.storage.AddWithUser(id, originalURL, userID)

	if err != nil {
		if errors.Is(err, database.ErrURLConflict) {
			// Если URL уже существует, возвращаем короткий URL с кодом конфликта
			shortURL := h.shortener.BuildShortURL(id)
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte(shortURL))
			return
		}
		logger.Logger.Error("Ошибка добавления URL", zap.Error(err))
		http.Error(w, "Ошибка сохранения URL", http.StatusInternalServerError)
		return
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
