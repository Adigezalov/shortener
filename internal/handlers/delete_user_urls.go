package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Adigezalov/shortener/internal/logger"
	"github.com/Adigezalov/shortener/internal/middleware"
	"go.uber.org/zap"
)

// DeleteUserURLs асинхронно удаляет URL пользователя
func (h *Handler) DeleteUserURLs(w http.ResponseWriter, r *http.Request) {
	// Получаем ID пользователя из контекста
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Декодируем список URL для удаления
	var shortURLs []string
	if err := json.NewDecoder(r.Body).Decode(&shortURLs); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Проверяем, что список не пустой
	if len(shortURLs) == 0 {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Запускаем асинхронное удаление URL
	go h.asyncDeleteURLs(userID, shortURLs)

	// Возвращаем статус 202 Accepted
	w.WriteHeader(http.StatusAccepted)

	logger.Logger.Info("Принят запрос на удаление URL",
		zap.String("user_id", userID),
		zap.Int("count", len(shortURLs)))
}

// asyncDeleteURLs асинхронно удаляет URL пользователя с использованием паттерна fanIn
func (h *Handler) asyncDeleteURLs(userID string, shortURLs []string) {
	// Выполняем пакетное удаление в хранилище
	if err := h.storage.DeleteUserURLs(userID, shortURLs); err != nil {
		logger.Logger.Error("Ошибка удаления URL пользователя",
			zap.String("user_id", userID),
			zap.Strings("short_urls", shortURLs),
			zap.Error(err),
		)
		return
	}

	logger.Logger.Info("URL пользователя успешно удалены",
		zap.String("user_id", userID),
		zap.Int("count", len(shortURLs)))
}
