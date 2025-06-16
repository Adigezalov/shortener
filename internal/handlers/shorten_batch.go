package handlers

import (
	"encoding/json"
	"github.com/Adigezalov/shortener/internal/logger"
	"github.com/Adigezalov/shortener/internal/models"
	"go.uber.org/zap"
	"net/http"

	"github.com/Adigezalov/shortener/internal/database"
)

// ShortenBatch обрабатывает POST запрос на пакетное создание сокращенных URL
func (h *Handler) ShortenBatch(w http.ResponseWriter, r *http.Request) {
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

		// Генерируем новый ID и пытаемся добавить URL
		id := h.shortener.Shorten(item.OriginalURL)
		id, exists, err := h.storage.Add(id, item.OriginalURL)
		if err != nil && err != database.ErrURLConflict {
			logger.Logger.Error("Ошибка добавления URL",
				zap.String("correlation_id", item.CorrelationID),
				zap.Error(err))
			continue
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
