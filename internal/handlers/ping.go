package handlers

import (
	"net/http"
)

// PingDB проверяет подключение к базе данных
func (h *Handler) PingDB(w http.ResponseWriter, r *http.Request) {
	// Если база данных не настроена, возвращаем OK
	if h.db == nil {
		w.WriteHeader(http.StatusOK)
		return
	}

	if err := h.db.Ping(); err != nil {
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
