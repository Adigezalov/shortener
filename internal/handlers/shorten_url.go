package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Adigezalov/shortener/internal/database"
	"github.com/Adigezalov/shortener/internal/logger"
	"github.com/Adigezalov/shortener/internal/middleware"
	"github.com/Adigezalov/shortener/internal/models"
	"go.uber.org/zap"
)

// ShortenURL обрабатывает POST запрос на создание сокращенного URL через JSON API.
//
// Эндпоинт: POST /api/shorten
// Content-Type: application/json
// Тело запроса: JSON объект с полем "url"
//
// Ответы:
//   - 201 Created: JSON с коротким URL в поле "result"
//   - 400 Bad Request: некорректный JSON или пустой URL
//   - 409 Conflict: URL уже существует (возвращает существующий короткий URL)
//   - 415 Unsupported Media Type: неправильный Content-Type
//   - 500 Internal Server Error: внутренняя ошибка сервера
//
// Пример запроса:
//
//	POST /api/shorten HTTP/1.1
//	Content-Type: application/json
//
//	{
//	  "url": "https://example.com/very/long/url"
//	}
//
// Пример ответа:
//
//	HTTP/1.1 201 Created
//	Content-Type: application/json
//
//	{
//	  "result": "http://localhost:8080/abc12345"
//	}
func (h *Handler) ShortenURL(w http.ResponseWriter, r *http.Request) {
	// Проверяем заголовок Content-Type
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		http.Error(w, "Неподдерживаемый тип контента", http.StatusUnsupportedMediaType)
		return
	}

	// Читаем запрос
	var request models.ShortenRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&request); err != nil {
		http.Error(w, "Неверный формат JSON", http.StatusBadRequest)
		return
	}

	// Проверяем URL
	if request.URL == "" {
		http.Error(w, "URL не может быть пустым", http.StatusBadRequest)
		return
	}

	// Получаем ID пользователя из контекста
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Генерируем новый ID и пытаемся добавить URL с привязкой к пользователю
	id := h.shortener.Shorten(request.URL)
	id, exists, err := h.storage.AddWithUser(id, request.URL, userID)

	if err != nil {
		if err == database.ErrURLConflict {
			// Если URL уже существует, возвращаем короткий URL с кодом конфликта
			shortURL := h.shortener.BuildShortURL(id)
			response := models.ShortenResponse{
				Result: shortURL,
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			encoder := json.NewEncoder(w)
			if err := encoder.Encode(response); err != nil {
				logger.Logger.Error("Ошибка кодирования JSON", zap.Error(err))
				http.Error(w, "Ошибка формирования ответа", http.StatusInternalServerError)
			}
			return
		}
		logger.Logger.Error("Ошибка добавления URL", zap.Error(err))
		http.Error(w, "Ошибка сохранения URL", http.StatusInternalServerError)
		return
	}

	// Строим полный короткий URL
	shortURL := h.shortener.BuildShortURL(id)

	// Формируем ответ
	response := models.ShortenResponse{
		Result: shortURL,
	}

	// Отправляем результат
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	encoder := json.NewEncoder(w)
	if err := encoder.Encode(response); err != nil {
		logger.Logger.Error("Ошибка кодирования JSON", zap.Error(err))
		http.Error(w, "Ошибка формирования ответа", http.StatusInternalServerError)
		return
	}

	logger.Logger.Info("URL сокращен (JSON API)",
		zap.String("original_url", request.URL),
		zap.String("short_url", shortURL),
		zap.Bool("existing", exists),
	)
}
