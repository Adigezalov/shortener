package handlers

import (
	"bytes"
	"encoding/json"
	"github.com/Adigezalov/shortener/internal/logger"
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
		mockSetup      func(*MockStorage, *MockShortener)
		expectedStatus int
		expectedResult []models.BatchShortenResponse
	}{
		{
			name: "Успешное создание нескольких коротких URL",
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
			mockSetup: func(ms *MockStorage, msh *MockShortener) {
				// Первый URL
				msh.On("Shorten", "https://example1.com").Return("abc123")
				ms.On("Add", "abc123", "https://example1.com").Return("abc123", false, nil)
				msh.On("BuildShortURL", "abc123").Return("http://short.url/abc123")

				// Второй URL
				msh.On("Shorten", "https://example2.com").Return("def456")
				ms.On("Add", "def456", "https://example2.com").Return("def456", false, nil)
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
			name: "Один URL уже существует",
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
			mockSetup: func(ms *MockStorage, msh *MockShortener) {
				// Первый URL (существующий)
				msh.On("Shorten", "https://example1.com").Return("existing123")
				ms.On("Add", "existing123", "https://example1.com").Return("existing123", true, nil)
				msh.On("BuildShortURL", "existing123").Return("http://short.url/existing123")

				// Второй URL (новый)
				msh.On("Shorten", "https://example2.com").Return("def456")
				ms.On("Add", "def456", "https://example2.com").Return("def456", false, nil)
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
			name:           "Пустой список URL",
			request:        []models.BatchShortenRequest{},
			contentType:    "application/json",
			mockSetup:      func(ms *MockStorage, msh *MockShortener) {},
			expectedStatus: http.StatusBadRequest,
			expectedResult: nil,
		},
		{
			name: "Неверный Content-Type",
			request: []models.BatchShortenRequest{
				{
					CorrelationID: "1",
					OriginalURL:   "https://example.com",
				},
			},
			contentType:    "text/plain",
			mockSetup:      func(ms *MockStorage, msh *MockShortener) {},
			expectedStatus: http.StatusUnsupportedMediaType,
			expectedResult: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем моки
			mockStorage := new(MockStorage)
			mockShortener := new(MockShortener)

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
