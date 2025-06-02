package handlers

import (
	"net/http"
)

// PingDB проверяет подключение к базе данных
func (h *Handler) PingDB(w http.ResponseWriter, r *http.Request) {
	if err := h.db.Ping(); err != nil {
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
