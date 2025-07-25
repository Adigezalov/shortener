package handlers

import (
	"github.com/Adigezalov/shortener/internal/logger"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"net/http"
)

// RedirectToURL обрабатывает GET запрос на перенаправление по короткому URL
func (h *Handler) RedirectToURL(w http.ResponseWriter, r *http.Request) {
	// Получаем ID из параметров запроса
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "ID не может быть пустым", http.StatusBadRequest)
		return
	}

	// Проверяем, не удален ли URL
	isDeleted, err := h.storage.IsDeleted(id)
	if err != nil {
		logger.Logger.Error("Ошибка проверки статуса удаления URL",
			zap.String("id", id),
			zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if isDeleted {
		logger.Logger.Info("Попытка доступа к удаленному URL",
			zap.String("id", id))
		http.Error(w, "Gone", http.StatusGone)
		return
	}

	// Ищем оригинальный URL в хранилище
	originalURL, found := h.storage.Get(id)
	if !found {
		http.Error(w, "URL не найден", http.StatusNotFound)
		return
	}

	// Перенаправляем на оригинальный URL
	w.Header().Set("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)

	logger.Logger.Info("Перенаправление по короткому URL",
		zap.String("id", id),
		zap.String("original_url", originalURL),
	)
}
