package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

const (
	address       = ":8080"
	shortIDLength = 8
)

var (
	urlStore = make(map[string]string)
	mu       sync.RWMutex
)

// generateShortID создаёт случайный короткий ID длиной shortIDLength символов
func generateShortID() (string, error) {
	b := make([]byte, shortIDLength)

	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(b)[:shortIDLength], nil
}

// getBaseURL определяет базовый URL
func getBaseURL(r *http.Request) string {
	scheme := "http"

	if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}

	return fmt.Sprintf("%s://%s/", scheme, r.Host)
}

// handleShorten обрабатывает POST-запрос для сокращения URL
func handleShorten(w http.ResponseWriter, r *http.Request) {
	// Проверяем метод
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)

		return
	}

	// Проверяем Content-Type
	if contentType := r.Header.Get("Content-Type"); contentType != "text/plain" {
		http.Error(w, "Content-Type must be text/plain", http.StatusUnsupportedMediaType)

		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil || len(body) == 0 {
		http.Error(w, "Bad request: empty body or read error", http.StatusBadRequest)

		return
	}

	originalURL := strings.TrimSpace(string(body))

	// Проверяем, что URL валиден
	if _, err := url.ParseRequestURI(originalURL); err != nil {
		http.Error(w, "Invalid URL", http.StatusBadRequest)

		return
	}

	// Проверяем, не сокращали ли уже этот URL
	mu.RLock()
	for id, link := range urlStore {
		if link == originalURL {
			shortURL := getBaseURL(r) + id

			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusCreated)

			_, _ = w.Write([]byte(shortURL))

			mu.RUnlock()

			return
		}
	}
	mu.RUnlock()

	// Генерируем новый короткий ID
	shortID, err := generateShortID()

	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Printf("Failed to generate short ID: %v", err)

		return
	}

	// Сохраняем в хранилище
	mu.Lock()
	urlStore[shortID] = originalURL
	mu.Unlock()

	shortURL := getBaseURL(r) + shortID

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write([]byte(shortURL))
}

// handleNormal обрабатывает GET-запрос по короткому URL
func handleNormal(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)

		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/")

	if id == "" {
		http.Error(w, "Not found", http.StatusNotFound)

		return
	}

	mu.RLock()
	originalURL, ok := urlStore[id]
	mu.RUnlock()

	if !ok {
		http.Error(w, "Not found", http.StatusNotFound)

		return
	}

	w.Header().Set("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)

}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.URL.Path == "/" && r.Method == http.MethodPost:
		handleShorten(w, r)
	case r.URL.Path != "/" && r.Method == http.MethodGet:
		handleNormal(w, r)
	default:
		http.Error(w, "Not found", http.StatusNotFound)
	}
}

func main() {
	http.HandleFunc("/", rootHandler)

	fmt.Printf("Server starting on %s", address)

	if err := http.ListenAndServe(address, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
