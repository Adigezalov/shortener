package handlers

import (
	"encoding/json"
	"github.com/Adigezalov/shortener/internal/logger"
	"github.com/Adigezalov/shortener/internal/models"
	"go.uber.org/zap"
	"net/http"
)

// ShortenBatch обрабатывает POST запрос на пакетное создание сокращенных URL
func (h *Handler) ShortenBatch(w http.ResponseWriter, r *http.Request) {
	// Проверяем заголовок Content-Type
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		http.Error(w, "Неподдерживаемый тип контента", http.StatusUnsupportedMediaType)
		return
	}

	// Читаем запрос
	var request []models.BatchShortenRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&request); err != nil {
		logger.Logger.Error("Ошибка декодирования JSON", zap.Error(err))
		http.Error(w, "Неверный формат JSON", http.StatusBadRequest)
		return
	}

	// Проверяем, что батч не пустой
	if len(request) == 0 {
		http.Error(w, "Пустой список URL", http.StatusBadRequest)
		return
	}

	// Создаем слайс для ответа
	response := make([]models.BatchShortenResponse, 0, len(request))

	// Обрабатываем каждый URL в батче
	for _, item := range request {
		// Проверяем URL
		if item.OriginalURL == "" {
			logger.Logger.Warn("Пустой URL в батче", zap.String("correlation_id", item.CorrelationID))
			continue
		}

		var id string
		var exists bool

		// Сначала проверяем, есть ли такой URL в хранилище
		id, exists = h.storage.FindByOriginalURL(item.OriginalURL)
		if !exists {
			// Если URL не найден, генерируем новый ID
			id = h.shortener.Shorten(item.OriginalURL)
			// Добавляем URL в хранилище
			id, _ = h.storage.Add(id, item.OriginalURL)
		}

		// Строим полный короткий URL
		shortURL := h.shortener.BuildShortURL(id)

		// Добавляем результат в ответ
		response = append(response, models.BatchShortenResponse{
			CorrelationID: item.CorrelationID,
			ShortURL:      shortURL,
		})

		logger.Logger.Info("URL сокращен (Batch API)",
			zap.String("correlation_id", item.CorrelationID),
			zap.String("original_url", item.OriginalURL),
			zap.String("short_url", shortURL),
			zap.Bool("existing", exists),
		)
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
}
