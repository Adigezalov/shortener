package app

import (
	"bytes"
	"compress/gzip"
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
	// Создаем тестовый сервер с middleware
	_storage := storage.NewMemoryStorage()
	_service := service.NewURLService(_storage, "http://localhost:8080")
	handler := NewHandlers(_service)

	router := chi.NewRouter()
	// Добавляем middleware как в реальном сервере
	router.Use(
		ungzipMiddleware,
		gzipMiddleware,
	)
	router.Mount("/", handler.Routes())
	testServer := httptest.NewServer(router)
	defer testServer.Close()

	// Вспомогательная функция для создания сжатого запроса
	createGzippedRequest := func(t *testing.T, method, url, body string) *http.Request {
		t.Helper()
		var buf bytes.Buffer
		gz := gzip.NewWriter(&buf)
		if _, err := gz.Write([]byte(body)); err != nil {
			t.Fatal(err)
		}
		if err := gz.Close(); err != nil {
			t.Fatal(err)
		}

		req, err := http.NewRequest(method, url, &buf)
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Encoding", "gzip")
		return req
	}

	t.Run("server_initialization", func(t *testing.T) {
		cfg := &config.Config{
			ServerAddress: ":8080",
			BaseURL:       "http://localhost:8080",
		}

		server := NewServer(*cfg)
		if server.httpServer.Addr != ":8080" {
			t.Errorf("expected server address :8080, got %s", server.httpServer.Addr)
		}
		if server.service == nil {
			t.Error("expected service to be initialized")
		}
	})

	t.Run("shorten_and_redirect_flow", func(t *testing.T) {
		// Тестируем обычный запрос
		t.Run("plain_request", func(t *testing.T) {
			testShortenAndRedirect(t, testServer, false)
		})

		// Тестируем сжатый запрос
		t.Run("gzipped_request", func(t *testing.T) {
			testShortenAndRedirect(t, testServer, true)
		})
	})

	t.Run("invalid requests", func(t *testing.T) {
		tests := []struct {
			name           string
			method         string
			path           string
			contentType    string
			body           string
			useGzip        bool
			expectedStatus int
		}{
			{
				name:           "POST_with_invalid_content_type",
				method:         http.MethodPost,
				path:           "/",
				contentType:    "application/json",
				body:           "http://example.com",
				expectedStatus: http.StatusUnsupportedMediaType,
			},
			{
				name:           "GET_to_root",
				method:         http.MethodGet,
				path:           "/",
				expectedStatus: http.StatusMethodNotAllowed,
			},
			{
				name:           "PUT_method_not_allowed",
				method:         http.MethodPut,
				path:           "/",
				expectedStatus: http.StatusMethodNotAllowed,
			},
			{
				name:           "non-existent_short_URL",
				method:         http.MethodGet,
				path:           "/nonexistent",
				expectedStatus: http.StatusNotFound,
			},
			{
				name:           "invalid_gzipped_request",
				method:         http.MethodPost,
				path:           "/",
				contentType:    "text/plain",
				body:           "invalid gzip data",
				useGzip:        true,
				expectedStatus: http.StatusBadRequest,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				var req *http.Request
				var err error

				if tt.useGzip {
					req = createGzippedRequest(t, tt.method, testServer.URL+tt.path, tt.body)
				} else {
					bodyBuf := bytes.NewBufferString(tt.body)
					req, err = http.NewRequest(tt.method, testServer.URL+tt.path, bodyBuf)
					if err != nil {
						t.Fatal(err)
					}
				}

				if tt.contentType != "" {
					req.Header.Set("Content-Type", tt.contentType)
				}

				// Устанавливаем Accept-Encoding для тестирования сжатых ответов
				req.Header.Set("Accept-Encoding", "gzip")

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

				// Проверяем сжатие ответа, если ожидается успешный ответ
				if resp.StatusCode < 400 {
					if ce := resp.Header.Get("Content-Encoding"); ce != "gzip" {
						t.Errorf("expected gzipped response, got Content-Encoding: %s", ce)
					}
				}
			})
		}
	})

	t.Run("gzip_response_handling", func(t *testing.T) {
		tests := []struct {
			name             string
			acceptEncoding   string
			expectCompressed bool
		}{
			{
				name:             "accepts_gzip",
				acceptEncoding:   "gzip",
				expectCompressed: true,
			},
			{
				name:             "no_accept_encoding",
				acceptEncoding:   "",
				expectCompressed: false,
			},
			{
				name:             "accepts_other_encoding",
				acceptEncoding:   "deflate",
				expectCompressed: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				body := bytes.NewBufferString("http://example.com")
				req, err := http.NewRequest(http.MethodPost, testServer.URL+"/", body)
				if err != nil {
					t.Fatal(err)
				}
				req.Header.Set("Content-Type", "text/plain")
				if tt.acceptEncoding != "" {
					req.Header.Set("Accept-Encoding", tt.acceptEncoding)
				}

				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					t.Fatal(err)
				}
				defer resp.Body.Close()

				contentEncoding := resp.Header.Get("Content-Encoding")
				if tt.expectCompressed && contentEncoding != "gzip" {
					t.Errorf("expected gzipped response, got Content-Encoding: %s", contentEncoding)
				} else if !tt.expectCompressed && contentEncoding == "gzip" {
					t.Error("unexpected gzipped response")
				}

				// Проверяем, что можем прочитать ответ
				var responseBody []byte
				if contentEncoding == "gzip" {
					gz, err := gzip.NewReader(resp.Body)
					if err != nil {
						t.Fatalf("failed to create gzip reader: %v", err)
					}
					defer gz.Close()
					responseBody, err = io.ReadAll(gz)
					if err != nil {
						t.Fatalf("failed to read gzipped response: %v", err)
					}
				} else {
					responseBody, err = io.ReadAll(resp.Body)
					if err != nil {
						t.Fatalf("failed to read response: %v", err)
					}
				}

				if len(responseBody) == 0 {
					t.Error("empty response body")
				}
			})
		}
	})
}

func testShortenAndRedirect(t *testing.T, testServer *httptest.Server, useGzip bool) {
	// Шаг 1: Создаем короткую ссылку
	var req *http.Request
	var err error

	if useGzip {
		req = createGzippedRequest(t, http.MethodPost, testServer.URL+"/", "http://example.com")
	} else {
		body := bytes.NewBufferString("http://example.com")
		req, err = http.NewRequest(http.MethodPost, testServer.URL+"/", body)
		if err != nil {
			t.Fatal(err)
		}
	}
	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("Accept-Encoding", "gzip")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("expected status 201, got %d", resp.StatusCode)
	}

	// Проверяем сжатие ответа
	if ce := resp.Header.Get("Content-Encoding"); ce != "gzip" {
		t.Errorf("expected gzipped response, got Content-Encoding: %s", ce)
	}

	// Распаковываем ответ если нужно
	var shortURL []byte
	gz, err := gzip.NewReader(resp.Body)
	if err != nil {
		t.Fatalf("failed to create gzip reader: %v", err)
	}
	shortURL, err = io.ReadAll(gz)
	gz.Close()
	if err != nil {
		t.Fatalf("failed to read gzipped response: %v", err)
	}

	// Шаг 2: Извлекаем ID из короткого URL
	shortID := strings.TrimPrefix(string(shortURL), testServer.URL+"/")

	// Шаг 3: Проверяем редирект
	req, err = http.NewRequest(http.MethodGet, testServer.URL+"/"+shortID, nil)
	if err != nil {
		t.Fatalf("Failed to create GET request for short URL %s: %v", shortID, err)
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
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusTemporaryRedirect {
		t.Errorf("expected status 307, got %d", resp.StatusCode)
	}

	location := resp.Header.Get("Location")
	if location != "http://example.com" {
		t.Errorf("expected location http://example.com, got %s", location)
	}
}

// Вспомогательная функция для создания сжатого запроса
func createGzippedRequest(t *testing.T, method, url, body string) *http.Request {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if _, err := gz.Write([]byte(body)); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest(method, url, &buf)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Content-Type", "text/plain")
	return req
}
