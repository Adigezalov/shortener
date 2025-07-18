package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/Adigezalov/shortener/internal/logger"
	"github.com/Adigezalov/shortener/internal/middleware"
	"github.com/Adigezalov/shortener/internal/models"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandler_ShortenBatch(t *testing.T) {
	// Инициализируем тестовый логгер
	testLogger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("Не удалось создать тестовый логгер: %v", err)
	}
	logger.Logger = testLogger
	defer logger.Logger.Sync()

	tests := []struct {
		name           string
		request        []models.BatchShortenRequest
		contentType    string
		mockSetup      func(*MockURLStorage, *MockURLShortener)
		expectedStatus int
		expectedResult []models.BatchShortenResponse
	}{
		{
			name: "Успешное_создание_нескольких_коротких_URL",
			request: []models.BatchShortenRequest{
				{
					CorrelationID: "1",
					OriginalURL:   "https://example1.com",
				},
				{
					CorrelationID: "2",
					OriginalURL:   "https://example2.com",
				},
			},
			contentType: "application/json",
			mockSetup: func(ms *MockURLStorage, msh *MockURLShortener) {
				// Первый URL
				msh.On("Shorten", "https://example1.com").Return("abc123")
				ms.On("AddWithUser", "abc123", "https://example1.com", "test-user").Return("abc123", false, nil)
				msh.On("BuildShortURL", "abc123").Return("http://short.url/abc123")

				// Второй URL
				msh.On("Shorten", "https://example2.com").Return("def456")
				ms.On("AddWithUser", "def456", "https://example2.com", "test-user").Return("def456", false, nil)
				msh.On("BuildShortURL", "def456").Return("http://short.url/def456")
			},
			expectedStatus: http.StatusCreated,
			expectedResult: []models.BatchShortenResponse{
				{
					CorrelationID: "1",
					ShortURL:      "http://short.url/abc123",
				},
				{
					CorrelationID: "2",
					ShortURL:      "http://short.url/def456",
				},
			},
		},
		{
			name: "Один_URL_уже_существует",
			request: []models.BatchShortenRequest{
				{
					CorrelationID: "1",
					OriginalURL:   "https://example1.com",
				},
				{
					CorrelationID: "2",
					OriginalURL:   "https://example2.com",
				},
			},
			contentType: "application/json",
			mockSetup: func(ms *MockURLStorage, msh *MockURLShortener) {
				// Первый URL (существующий)
				msh.On("Shorten", "https://example1.com").Return("existing123")
				ms.On("AddWithUser", "existing123", "https://example1.com", "test-user").Return("existing123", true, nil)
				msh.On("BuildShortURL", "existing123").Return("http://short.url/existing123")

				// Второй URL (новый)
				msh.On("Shorten", "https://example2.com").Return("def456")
				ms.On("AddWithUser", "def456", "https://example2.com", "test-user").Return("def456", false, nil)
				msh.On("BuildShortURL", "def456").Return("http://short.url/def456")
			},
			expectedStatus: http.StatusCreated,
			expectedResult: []models.BatchShortenResponse{
				{
					CorrelationID: "1",
					ShortURL:      "http://short.url/existing123",
				},
				{
					CorrelationID: "2",
					ShortURL:      "http://short.url/def456",
				},
			},
		},
		{
			name:           "Пустой_список_URL",
			request:        []models.BatchShortenRequest{},
			contentType:    "application/json",
			mockSetup:      func(ms *MockURLStorage, msh *MockURLShortener) {},
			expectedStatus: http.StatusBadRequest,
			expectedResult: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем моки
			mockStorage := new(MockURLStorage)
			mockShortener := new(MockURLShortener)

			// Настраиваем моки
			tt.mockSetup(mockStorage, mockShortener)

			// Создаем обработчик с моками
			handler := &Handler{
				storage:   mockStorage,
				shortener: mockShortener,
			}

			// Создаем тело запроса
			body, _ := json.Marshal(tt.request)

			// Создаем тестовый запрос
			req := httptest.NewRequest("POST", "/api/shorten/batch", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", tt.contentType)
			
			// Добавляем userID в контекст для тестов, которые требуют аутентификации
			if tt.expectedStatus == http.StatusCreated {
				ctx := context.WithValue(req.Context(), middleware.UserIDKey, "test-user")
				req = req.WithContext(ctx)
			}
			
			w := httptest.NewRecorder()

			// Вызываем тестируемый обработчик
			handler.ShortenBatch(w, req)

			// Проверяем статус код
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Если ожидается успешный ответ, проверяем результат
			if tt.expectedStatus == http.StatusCreated {
				var response []models.BatchShortenResponse
				err := json.NewDecoder(w.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, response)
			}

			// Проверяем, что все ожидаемые вызовы моков были выполнены
			mockStorage.AssertExpectations(t)
			mockShortener.AssertExpectations(t)
		})
	}
}
