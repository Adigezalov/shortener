package handlers

import (
	"bytes"
	"encoding/json"
	"github.com/Adigezalov/shortener/internal/database"
	"github.com/Adigezalov/shortener/internal/logger"
	"github.com/Adigezalov/shortener/internal/models"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandler_ShortenURL(t *testing.T) {
	// Инициализируем тестовый логгер
	testLogger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("Не удалось создать тестовый логгер: %v", err)
	}
	logger.Logger = testLogger
	defer logger.Logger.Sync()

	tests := []struct {
		name           string
		request        models.ShortenRequest
		contentType    string
		mockSetup      func(*MockStorage, *MockShortener)
		expectedStatus int
		expectedResult string
	}{
		{
			name: "Успешное создание нового короткого URL",
			request: models.ShortenRequest{
				URL: "https://example.com",
			},
			contentType: "application/json",
			mockSetup: func(ms *MockStorage, msh *MockShortener) {
				msh.On("Shorten", "https://example.com").Return("abc123")
				ms.On("Add", "abc123", "https://example.com").Return("abc123", false, nil)
				msh.On("BuildShortURL", "abc123").Return("http://short.url/abc123")
			},
			expectedStatus: http.StatusCreated,
			expectedResult: "http://short.url/abc123",
		},
		{
			name: "URL уже существует в базе",
			request: models.ShortenRequest{
				URL: "https://example.com",
			},
			contentType: "application/json",
			mockSetup: func(ms *MockStorage, msh *MockShortener) {
				msh.On("Shorten", "https://example.com").Return("abc123")
				ms.On("Add", "abc123", "https://example.com").Return("abc123", true, database.ErrURLConflict)
				msh.On("BuildShortURL", "abc123").Return("http://short.url/abc123")
			},
			expectedStatus: http.StatusConflict,
			expectedResult: "http://short.url/abc123",
		},
		{
			name: "Пустой URL",
			request: models.ShortenRequest{
				URL: "",
			},
			contentType:    "application/json",
			mockSetup:      func(ms *MockStorage, msh *MockShortener) {},
			expectedStatus: http.StatusBadRequest,
			expectedResult: "",
		},
		{
			name: "Неверный Content-Type",
			request: models.ShortenRequest{
				URL: "https://example.com",
			},
			contentType:    "text/plain",
			mockSetup:      func(ms *MockStorage, msh *MockShortener) {},
			expectedStatus: http.StatusUnsupportedMediaType,
			expectedResult: "",
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
			req := httptest.NewRequest("POST", "/api/shorten", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", tt.contentType)
			w := httptest.NewRecorder()

			// Вызываем тестируемый обработчик
			handler.ShortenURL(w, req)

			// Проверяем статус код
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Если ожидается успешный ответ или конфликт, проверяем результат
			if tt.expectedStatus == http.StatusCreated || tt.expectedStatus == http.StatusConflict {
				var response models.ShortenResponse
				err := json.NewDecoder(w.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, response.Result)
			}

			// Проверяем, что все ожидаемые вызовы моков были выполнены
			mockStorage.AssertExpectations(t)
			mockShortener.AssertExpectations(t)
		})
	}
}
