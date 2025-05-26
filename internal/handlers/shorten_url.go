package handlers

import (
	"encoding/json"
	"github.com/Adigezalov/shortener/internal/logger"
	"github.com/Adigezalov/shortener/internal/models"
	"go.uber.org/zap"
	"net/http"
)

// ShortenURL обрабатывает POST запрос на создание сокращенного URL (JSON API)
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
		logger.Logger.Error("Ошибка декодирования JSON", zap.Error(err))
		http.Error(w, "Неверный формат JSON", http.StatusBadRequest)
		return
	}

	// Проверяем URL
	if request.URL == "" {
		http.Error(w, "URL не может быть пустым", http.StatusBadRequest)
		return
	}

	// Создаем или получаем существующий короткий ID
	var id string
	var exists bool

	// Сначала проверяем, есть ли такой URL в хранилище
	id, exists = h.storage.FindByOriginalURL(request.URL)
	if !exists {
		// Если URL не найден, генерируем новый ID
		id = h.shortener.Shorten(request.URL)
		// Добавляем URL в хранилище
		id, _ = h.storage.Add(id, request.URL)
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
