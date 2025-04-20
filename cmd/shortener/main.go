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
func getBaseURL(req *http.Request) string {
	scheme := "http"

	if req.TLS != nil || req.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}

	return fmt.Sprintf("%s://%s/", scheme, req.Host)
}

// handleShorten обрабатывает POST-запрос для сокращения URL
func handleShorten(res http.ResponseWriter, req *http.Request) {
	// Проверяем метод
	if req.Method != http.MethodPost {
		http.Error(res, "Method not allowed", http.StatusMethodNotAllowed)

		return
	}

	// Проверяем Content-Type
	if contentType := req.Header.Get("Content-Type"); !strings.HasPrefix(contentType, "text/plain") {
		http.Error(res, "Content-Type must be text/plain", http.StatusUnsupportedMediaType)
		return
	}

	body, err := io.ReadAll(req.Body)
	if err != nil || len(body) == 0 {
		http.Error(res, "Bad request: empty body or read error", http.StatusBadRequest)

		return
	}

	originalURL := strings.TrimSpace(string(body))

	// Проверяем, что URL валиден
	if _, err := url.ParseRequestURI(originalURL); err != nil {
		http.Error(res, "Invalid URL", http.StatusBadRequest)

		return
	}

	// Проверяем, не сокращали ли уже этот URL
	mu.RLock()
	for id, link := range urlStore {
		if link == originalURL {
			shortURL := getBaseURL(req) + id

			res.Header().Set("Content-Type", "text/plain")
			res.WriteHeader(http.StatusCreated)

			_, _ = res.Write([]byte(shortURL))

			mu.RUnlock()

			return
		}
	}
	mu.RUnlock()

	// Генерируем новый короткий ID
	shortID, err := generateShortID()

	if err != nil {
		http.Error(res, "Internal server error", http.StatusInternalServerError)
		log.Printf("Failed to generate short ID: %v", err)

		return
	}

	// Сохраняем в хранилище
	mu.Lock()
	urlStore[shortID] = originalURL
	mu.Unlock()

	shortURL := getBaseURL(req) + shortID

	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(http.StatusCreated)
	_, _ = res.Write([]byte(shortURL))
}

// handleNormal обрабатывает GET-запрос по короткому URL
func handleNormal(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(res, "Method not allowed", http.StatusMethodNotAllowed)

		return
	}

	id := strings.TrimPrefix(req.URL.Path, "/")

	if id == "" {
		http.Error(res, "Not found", http.StatusBadRequest)

		return
	}

	mu.RLock()
	originalURL, ok := urlStore[id]
	mu.RUnlock()

	if !ok {
		http.Error(res, "Not found", http.StatusBadRequest)

		return
	}

	res.Header().Set("Content-Type", "text/plain")
	res.Header().Set("Location", originalURL)
	res.WriteHeader(http.StatusTemporaryRedirect)

}

func rootHandler(res http.ResponseWriter, req *http.Request) {
	switch {
	case req.URL.Path == "/" && req.Method == http.MethodPost:
		handleShorten(res, req)
	case req.URL.Path != "/" && req.Method == http.MethodGet:
		handleNormal(res, req)
	default:
		http.Error(res, "Not found", http.StatusNotFound)
	}
}

func main() {
	http.HandleFunc("/", rootHandler)

	fmt.Printf("Server starting on %s", address)

	if err := http.ListenAndServe(address, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
