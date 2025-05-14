package app

import (
	"bytes"
	"encoding/json"
	"github.com/Adigezalov/shortener/internal/service"
	"github.com/Adigezalov/shortener/internal/storage"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandleShorten(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		host           string
		contentType    string
		body           string
		prepopulate    map[string]string // shortID -> originalURL для предварительного заполнения хранилища
		expectedStatus int
		expectedBody   string
		isError        bool
		checkStorage   bool
	}{
		{
			name:           "successful_shorten_new_URL",
			method:         http.MethodPost,
			host:           "example.com",
			contentType:    "text/plain",
			body:           "http://example.com",
			prepopulate:    nil,
			expectedStatus: http.StatusCreated,
			expectedBody:   "http://example.com/",
			checkStorage:   true,
		},
		{
			name:           "empty_body",
			method:         http.MethodPost,
			contentType:    "text/plain",
			prepopulate:    nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Bad request: empty body or read error\n",
			checkStorage:   false,
			isError:        true,
		},
		{
			name:           "invalid_content_type",
			method:         http.MethodPost,
			contentType:    "application/json",
			body:           "http://example.com",
			prepopulate:    nil,
			expectedStatus: http.StatusUnsupportedMediaType,
			expectedBody:   "Content-Type must be text/plain\n",
			checkStorage:   false,
			isError:        true,
		},
		{
			name:           "wrong_HTTP_method",
			method:         http.MethodGet,
			contentType:    "text/plain",
			body:           "http://example.com",
			prepopulate:    nil,
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   "Method not allowed\n",
			checkStorage:   false,
			isError:        true,
		},
		{
			name:           "invalid_URL",
			method:         http.MethodPost,
			contentType:    "text/plain",
			body:           "invalid-url",
			prepopulate:    nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "invalid URL\n",
			checkStorage:   false,
			isError:        true,
		},
		{
			name:        "URL_already_exists",
			method:      http.MethodPost,
			host:        "localhost:8080",
			contentType: "text/plain",
			body:        "http://existing.com",
			prepopulate: map[string]string{
				"abc12345": "http://existing.com",
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   "http://localhost:8080/abc12345",
			checkStorage:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := storage.NewMemoryStorage()
			for shortID, originalURL := range tt.prepopulate {
				if err := store.Save(shortID, originalURL); err != nil {
					t.Fatalf("failed to prepopulate storage: %v", err)
				}
			}

			urlService := service.NewURLService(store, "http://localhost:8080")
			h := &Handlers{
				service: urlService,
			}

			req := httptest.NewRequest(tt.method, "/", bytes.NewBufferString(tt.body))
			req.Host = tt.host
			req.Header.Set("Content-Type", tt.contentType)
			w := httptest.NewRecorder()

			urlService.SetBaseURL(req)

			h.handleShorten(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			// Проверяем статус код
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			// Читаем тело ответа
			bodyBytes, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("failed to read response body: %v", err)
			}
			body := string(bodyBytes)

			// Для случаев с ошибкой проверяем только точное соответствие тела ответа
			if tt.isError {
				if body != tt.expectedBody {
					t.Errorf("expected body %q, got %q", tt.expectedBody, body)
				}
				return // Дальнейшие проверки не требуются
			}

			// Проверки только для успешных случаев:

			// 1. Проверяем что ответ начинается с ожидаемого базового URL
			if !strings.HasPrefix(body, tt.expectedBody) {
				t.Errorf("expected body to start with %q, got %q", tt.expectedBody, body)
			}

			// 2. Извлекаем shortID (часть после последнего '/')
			shortID := body[strings.LastIndex(body, "/")+1:]

			// 3. Проверяем длину shortID (предполагаем 8 символов)
			if len(shortID) != 8 {
				t.Errorf("expected shortID length 8, got %d (%q)", len(shortID), shortID)
			}

			// 4. Проверяем хранилище если требуется
			if tt.checkStorage {
				originalURL := strings.TrimSpace(tt.body)

				// Проверяем что URL сохранился с правильным shortID
				if storedURL, exists := store.Get(shortID); !exists || storedURL != originalURL {
					t.Errorf("expected URL %q to be saved in storage with shortID %q", originalURL, shortID)
				}

				// Проверяем обратное соответствие
				if storedShortID, exists := store.Exists(originalURL); !exists || storedShortID != shortID {
					t.Errorf("expected shortID %q for URL %q in reverse storage", shortID, originalURL)
				}
			}
		})
	}
}

func TestHandleNormal(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		path           string
		prepopulate    map[string]string // shortID -> originalURL
		expectedStatus int
		expectedBody   string
		expectedLoc    string
	}{
		{
			name:           "successful redirect",
			method:         http.MethodGet,
			path:           "/abc123",
			prepopulate:    map[string]string{"abc123": "http://example.com"},
			expectedStatus: http.StatusTemporaryRedirect,
			expectedLoc:    "http://example.com",
		},
		{
			name:           "wrong HTTP method",
			method:         http.MethodPost,
			path:           "/abc123",
			prepopulate:    map[string]string{"abc123": "http://example.com"},
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   "Method not allowed\n",
		},
		{
			name:           "empty path",
			method:         http.MethodGet,
			path:           "/",
			prepopulate:    nil,
			expectedStatus: http.StatusNotFound,
			expectedBody:   "Not found\n",
		},
		{
			name:           "non-existent ID",
			method:         http.MethodGet,
			path:           "/nonexistent",
			prepopulate:    map[string]string{"abc123": "http://example.com"},
			expectedStatus: http.StatusNotFound,
			expectedBody:   "Not found\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := storage.NewMemoryStorage()
			for shortID, originalURL := range tt.prepopulate {
				if err := store.Save(shortID, originalURL); err != nil {
					t.Fatalf("failed to prepopulate storage: %v", err)
				}
			}

			urlService := service.NewURLService(store, "http://localhost:8080")
			h := NewHandlers(urlService)

			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			h.handleNormal(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			// Проверяем статус код
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			// Проверяем тело ответа для ошибок
			if tt.expectedBody != "" {
				bodyBytes, err := io.ReadAll(resp.Body)
				if err != nil {
					t.Fatalf("failed to read response body: %v", err)
				}
				body := string(bodyBytes)
				if body != tt.expectedBody {
					t.Errorf("expected body %q, got %q", tt.expectedBody, body)
				}
			}

			// Проверяем Location header для редиректов
			if tt.expectedLoc != "" {
				loc := resp.Header.Get("Location")
				if loc != tt.expectedLoc {
					t.Errorf("expected Location header %q, got %q", tt.expectedLoc, loc)
				}
			}

			// Проверяем Content-Type header
			if tt.expectedStatus == http.StatusTemporaryRedirect {
				contentType := resp.Header.Get("Content-Type")
				if contentType != "text/plain" {
					t.Errorf("expected Content-Type header 'text/plain', got %q", contentType)
				}
			}
		})
	}
}

func TestHandleJSONShorten(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		host           string
		contentType    string
		body           string
		prepopulate    map[string]string // shortID -> originalURL для предварительного заполнения хранилища
		expectedStatus int
		expectedBody   string
		isError        bool
		checkStorage   bool
	}{
		{
			name:           "successful_shorten_new_URL",
			method:         http.MethodPost,
			host:           "example.com",
			contentType:    "application/json",
			body:           `{"url":"http://example.com"}`,
			prepopulate:    nil,
			expectedStatus: http.StatusCreated,
			expectedBody:   `{"result":"http://example.com/`,
			checkStorage:   true,
		},
		{
			name:           "empty_body",
			method:         http.MethodPost,
			contentType:    "application/json",
			body:           "",
			prepopulate:    nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `Bad request: invalid JSON`,
			isError:        true,
		},
		{
			name:           "invalid_content_type",
			method:         http.MethodPost,
			contentType:    "text/plain",
			body:           `{"url":"http://example.com"}`,
			prepopulate:    nil,
			expectedStatus: http.StatusUnsupportedMediaType,
			expectedBody:   `Content-Type must be application/json`,
			isError:        true,
		},
		{
			name:           "wrong_HTTP_method",
			method:         http.MethodGet,
			contentType:    "application/json",
			body:           `{"url":"http://example.com"}`,
			prepopulate:    nil,
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   `Method not allowed`,
			isError:        true,
		},
		{
			name:           "invalid_URL",
			method:         http.MethodPost,
			contentType:    "application/json",
			body:           `{"url":"invalid-url"}`,
			prepopulate:    nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `invalid URL`,
			isError:        true,
		},
		{
			name:           "empty_URL",
			method:         http.MethodPost,
			contentType:    "application/json",
			body:           `{"url":""}`,
			prepopulate:    nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `URL cannot be empty`,
			isError:        true,
		},
		{
			name:           "invalid_JSON",
			method:         http.MethodPost,
			contentType:    "application/json",
			body:           `{"url":}`,
			prepopulate:    nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `Bad request: invalid JSON`,
			isError:        true,
		},
		{
			name:           "missing_url_field",
			method:         http.MethodPost,
			contentType:    "application/json",
			body:           `{"not_url":"http://example.com"}`,
			prepopulate:    nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `URL cannot be empty`,
			isError:        true,
		},
		{
			name:        "URL_already_exists",
			method:      http.MethodPost,
			host:        "localhost:8080",
			contentType: "application/json",
			body:        `{"url":"http://existing.com"}`,
			prepopulate: map[string]string{
				"abc12345": "http://existing.com",
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   `{"result":"http://localhost:8080/abc12345"}`,
			checkStorage:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := storage.NewMemoryStorage()
			for shortID, originalURL := range tt.prepopulate {
				if err := store.Save(shortID, originalURL); err != nil {
					t.Fatalf("failed to prepopulate storage: %v", err)
				}
			}

			urlService := service.NewURLService(store, "http://localhost:8080")
			h := &Handlers{
				service: urlService,
			}

			req := httptest.NewRequest(tt.method, "/api/shorten", bytes.NewBufferString(tt.body))
			req.Host = tt.host
			req.Header.Set("Content-Type", tt.contentType)
			w := httptest.NewRecorder()

			urlService.SetBaseURL(req)

			h.handleJSONShorten(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			// Проверяем статус код
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			// Читаем тело ответа
			bodyBytes, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("failed to read response body: %v", err)
			}
			body := string(bodyBytes)

			// Для случаев с ошибкой проверяем только наличие ожидаемого текста в теле ответа
			if tt.isError {
				if !strings.Contains(body, tt.expectedBody) {
					t.Errorf("expected body to contain %q, got %q", tt.expectedBody, body)
				}
				return // Дальнейшие проверки не требуются
			}

			// Проверяем Content-Type для успешных ответов
			contentType := resp.Header.Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("expected Content-Type 'application/json', got %q", contentType)
			}

			// Проверяем что ответ начинается с ожидаемого JSON
			if !strings.HasPrefix(body, tt.expectedBody) {
				t.Errorf("expected body to start with %q, got %q", tt.expectedBody, body)
			}

			// Для успешных ответов парсим JSON и проверяем структуру
			var respJSON shortenResponse
			if err := json.Unmarshal(bodyBytes, &respJSON); err != nil {
				t.Fatalf("failed to unmarshal response JSON: %v", err)
			}

			// Проверяем что result начинается с ожидаемого базового URL
			if !strings.HasPrefix(respJSON.Result, "http://"+tt.host+"/") &&
				!strings.HasPrefix(respJSON.Result, tt.expectedBody) {
				t.Errorf("expected result to start with base URL, got %q", respJSON.Result)
			}

			// Проверяем хранилище если требуется
			if tt.checkStorage {
				// Извлекаем URL из тела запроса
				var reqJSON shortenRequest
				if err := json.Unmarshal([]byte(tt.body), &reqJSON); err != nil {
					t.Fatalf("failed to unmarshal test case body: %v", err)
				}
				originalURL := strings.TrimSpace(reqJSON.URL)

				// Извлекаем shortID из ответа
				shortID := respJSON.Result[strings.LastIndex(respJSON.Result, "/")+1:]

				// Проверяем что URL сохранился с правильным shortID
				if storedURL, exists := store.Get(shortID); !exists || storedURL != originalURL {
					t.Errorf("expected URL %q to be saved in storage with shortID %q", originalURL, shortID)
				}

				// Проверяем обратное соответствие
				if storedShortID, exists := store.Exists(originalURL); !exists || storedShortID != shortID {
					t.Errorf("expected shortID %q for URL %q in reverse storage", shortID, originalURL)
				}
			}
		})
	}
}
