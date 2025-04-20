package app

import (
	"github.com/Adigezalov/shortener/internal/service"
	"io"
	"log"
	"net/http"
	"strings"
)

type Handlers struct {
	service *service.URLService
}

func NewHandlers(service *service.URLService) *Handlers {
	return &Handlers{service: service}
}

func (h *Handlers) RootHandler(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.URL.Path == "/" && r.Method == http.MethodPost:
		h.handleShorten(w, r)
	case r.URL.Path != "/" && r.Method == http.MethodGet:
		h.handleNormal(w, r)
	default:
		http.Error(w, "Not found", http.StatusNotFound)
	}
}

func (h *Handlers) handleShorten(w http.ResponseWriter, r *http.Request) {
	// Устанавливаем базовый URL для сервиса
	h.service.SetBaseURL(r)

	// Проверяем метод
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)

		return
	}

	// Проверяем Content-Type
	if contentType := r.Header.Get("Content-Type"); !strings.HasPrefix(contentType, "text/plain") {
		http.Error(w, "Content-Type must be text/plain", http.StatusUnsupportedMediaType)

		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil || len(body) == 0 {
		http.Error(w, "Bad request: empty body or read error", http.StatusBadRequest)

		return
	}

	originalURL := strings.TrimSpace(string(body))

	// Пытаемся получить короткий URL, если он уже существует
	if shortURL, exists := h.service.GetShortURLIfExists(originalURL); exists {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusCreated)

		_, _ = w.Write([]byte(shortURL))

		return
	}

	// Создаем новый короткий URL
	shortURL, err := h.service.ShortenURL(originalURL)
	if err != nil {
		switch err {
		case service.ErrInvalidURL:
			http.Error(w, "Invalid URL", http.StatusBadRequest)
		case service.ErrEmptyURL:
			http.Error(w, "Empty URL", http.StatusBadRequest)
		default:
			log.Printf("Failed to shorten URL: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)

	_, _ = w.Write([]byte(shortURL))
}

func (h *Handlers) handleNormal(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)

		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/")
	if id == "" {
		http.Error(w, "Not found", http.StatusNotFound)

		return
	}

	originalURL, err := h.service.GetOriginalURL(id)
	if err != nil {
		http.Error(w, "Not found", http.StatusNotFound)

		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
