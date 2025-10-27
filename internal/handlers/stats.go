package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Adigezalov/shortener/internal/logger"
	"go.uber.org/zap"
)

// StatsResponse представляет ответ статистики
type StatsResponse struct {
	URLs  int `json:"urls"`  // Количество сокращённых URL в сервисе
	Users int `json:"users"` // Количество пользователей в сервисе
}

// GetStats возвращает статистику сервиса
func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	// Получаем статистику из хранилища
	stats, err := h.storage.Stats()
	if err != nil {
		logger.Logger.Error("Ошибка получения статистики", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Формируем ответ
	statsResp := StatsResponse{
		URLs:  stats.URLs,
		Users: stats.Users,
	}

	// Устанавливаем заголовок Content-Type
	w.Header().Set("Content-Type", "application/json")

	// Кодируем и отправляем ответ
	if err := json.NewEncoder(w).Encode(statsResp); err != nil {
		logger.Logger.Error("Ошибка кодирования ответа статистики", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	logger.Logger.Info("Статистика успешно получена",
		zap.Int("urls", statsResp.URLs),
		zap.Int("users", statsResp.Users))
}
