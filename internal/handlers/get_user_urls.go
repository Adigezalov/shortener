package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Adigezalov/shortener/internal/logger"
	"github.com/Adigezalov/shortener/internal/middleware"
	"github.com/Adigezalov/shortener/internal/models"
	"go.uber.org/zap"
)

// GetUserURLs возвращает все URL пользователя
func (h *Handler) GetUserURLs(w http.ResponseWriter, r *http.Request) {
	// Получаем ID пользователя из контекста
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Получаем URL пользователя из хранилища
	userURLs, err := h.storage.GetUserURLs(userID)
	if err != nil {
		logger.Logger.Error("Ошибка получения URL пользователя",
			zap.String("user_id", userID),
			zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Если у пользователя нет URL, возвращаем 204 No Content
	if len(userURLs) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Преобразуем URL в полные ссылки
	result := make([]models.UserURL, len(userURLs))
	for i, userURL := range userURLs {
		result[i] = models.UserURL{
			ShortURL:    h.shortener.BuildShortURL(userURL.ShortURL),
			OriginalURL: userURL.OriginalURL,
		}
	}

	// Устанавливаем заголовок Content-Type
	w.Header().Set("Content-Type", "application/json")

	// Кодируем и отправляем ответ
	if err := json.NewEncoder(w).Encode(result); err != nil {
		logger.Logger.Error("Ошибка кодирования ответа",
			zap.String("user_id", userID),
			zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	logger.Logger.Info("Возвращены URL пользователя",
		zap.String("user_id", userID),
		zap.Int("count", len(result)))
}
