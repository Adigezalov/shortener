package app

import (
	"encoding/json"
	"errors"
	"github.com/Adigezalov/shortener/internal/service"
	"github.com/go-chi/chi/v5"
	"io"
	"log"
	"net/http"
	"strings"
)

type Handlers struct {
	service URLService
}

type shortenRequest struct {
	URL string `json:"url"`
}

type shortenResponse struct {
	Result string `json:"result"`
}

func NewHandlers(service URLService) *Handlers {
	return &Handlers{service: service}
}

func (h *Handlers) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/", h.handleShorten)
	r.Post("/api/shorten", h.handleJSONShorten)
	r.Get("/{id}", h.handleNormal)

	return r
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

		if _, err := w.Write([]byte(shortURL)); err != nil {
			log.Printf("Failed to write response: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		return
	}

	// Создаем новый короткий URL
	shortURL, err := h.service.ShortenURL(originalURL)

	if err != nil {
		if errors.Is(err, service.ErrInvalidURL) || errors.Is(err, service.ErrEmptyURL) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		log.Printf("Failed to shorten URL '%s': %v", originalURL, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)

	if _, err := w.Write([]byte(shortURL)); err != nil {
		log.Printf("Failed to write response: %v (shortURL: %s)", err, shortURL)

		return
	}
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

func (h *Handlers) handleJSONShorten(w http.ResponseWriter, r *http.Request) {
	// Устанавливаем базовый URL для сервиса
	h.service.SetBaseURL(r)

	// Проверяем метод
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Проверяем Content-Type
	if contentType := r.Header.Get("Content-Type"); contentType != "application/json" {
		http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
		return
	}

	// Читаем и парсим тело запроса
	var req shortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request: invalid JSON", http.StatusBadRequest)
		return
	}

	originalURL := strings.TrimSpace(req.URL)
	if originalURL == "" {
		http.Error(w, "URL cannot be empty", http.StatusBadRequest)
		return
	}

	// Пытаемся получить короткий URL, если он уже существует
	if shortURL, exists := h.service.GetShortURLIfExists(originalURL); exists {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)

		resp := shortenResponse{Result: shortURL}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Printf("Failed to encode response: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	// Создаем новый короткий URL
	shortURL, err := h.service.ShortenURL(originalURL)
	if err != nil {
		if errors.Is(err, service.ErrInvalidURL) || errors.Is(err, service.ErrEmptyURL) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		log.Printf("Failed to shorten URL '%s': %v", originalURL, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Формируем и отправляем ответ
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	resp := shortenResponse{Result: shortURL}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("Failed to encode response: %v (shortURL: %s)", err, shortURL)
		return
	}
}
