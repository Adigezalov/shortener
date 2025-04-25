package app

import (
	"bytes"
	"github.com/Adigezalov/shortener/internal/config"
	"github.com/Adigezalov/shortener/internal/service"
	"github.com/Adigezalov/shortener/internal/storage"
	"github.com/go-chi/chi/v5"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestServer(t *testing.T) {
	// Создаем тестовый сервер
	_storage := storage.NewMemoryStorage()
	_service := service.NewURLService(_storage, "http://localhost:8080")
	handler := NewHandlers(_service)

	router := chi.NewRouter()
	router.Mount("/", handler.Routes())
	testServer := httptest.NewServer(router)
	defer testServer.Close()

	t.Run("server initialization", func(t *testing.T) {
		cfg := config.Config{
			ServerAddress: ":8080",
			BaseURL:       "http://localhost:8080",
		}

		server := NewServer(cfg)
		if server.httpServer.Addr != ":8080" {
			t.Errorf("expected server address :8080, got %s", server.httpServer.Addr)
		}
		if server.service == nil {
			t.Error("expected service to be initialized")
		}
	})

	t.Run("shorten and redirect flow", func(t *testing.T) {
		// Шаг 1: Создаем короткую ссылку
		body := bytes.NewBufferString("http://example.com")
		req, err := http.NewRequest(http.MethodPost, testServer.URL+"/", body)
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "text/plain")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Errorf("failed to close response body: %v", err)
			}
		}()

		if resp.StatusCode != http.StatusCreated {
			t.Errorf("expected status 201, got %d", resp.StatusCode)
		}

		shortURL, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}

		// Шаг 2: Извлекаем ID из короткого URL
		shortID := strings.TrimPrefix(string(shortURL), testServer.URL+"/")

		// Шаг 3: Проверяем редирект
		req, err = http.NewRequest(http.MethodGet, testServer.URL+"/"+shortID, nil)
		if err != nil {
			t.Fatal(err)
		}

		client := &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse // Не следовать редиректу
			},
		}

		resp, err = client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Errorf("failed to close response body: %v", err)
			}
		}()

		if resp.StatusCode != http.StatusTemporaryRedirect {
			t.Errorf("expected status 307, got %d", resp.StatusCode)
		}

		location := resp.Header.Get("Location")
		if location != "http://example.com" {
			t.Errorf("expected location http://example.com, got %s", location)
		}
	})

	t.Run("invalid requests", func(t *testing.T) {
		tests := []struct {
			name           string
			method         string
			path           string
			contentType    string
			body           string
			expectedStatus int
		}{
			{
				name:           "POST with invalid content type",
				method:         http.MethodPost,
				path:           "/",
				contentType:    "application/json",
				body:           "http://example.com",
				expectedStatus: http.StatusUnsupportedMediaType,
			},
			{
				name:           "GET to root",
				method:         http.MethodGet,
				path:           "/",
				expectedStatus: http.StatusMethodNotAllowed,
			},
			{
				name:           "PUT method not allowed",
				method:         http.MethodPut,
				path:           "/",
				expectedStatus: http.StatusMethodNotAllowed,
			},
			{
				name:           "non-existent short URL",
				method:         http.MethodGet,
				path:           "/nonexistent",
				expectedStatus: http.StatusNotFound,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				var bodyBuf *bytes.Buffer
				if tt.body != "" {
					bodyBuf = bytes.NewBufferString(tt.body)
				} else {
					bodyBuf = bytes.NewBufferString("")
				}

				req, err := http.NewRequest(tt.method, testServer.URL+tt.path, bodyBuf)
				if err != nil {
					t.Fatal(err)
				}

				if tt.contentType != "" {
					req.Header.Set("Content-Type", tt.contentType)
				}

				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					t.Fatal(err)
				}
				defer func() {
					if err := resp.Body.Close(); err != nil {
						t.Errorf("failed to close response body: %v", err)
					}
				}()

				if resp.StatusCode != tt.expectedStatus {
					t.Errorf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
				}
			})
		}
	})
}
