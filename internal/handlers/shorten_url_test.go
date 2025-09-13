package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Adigezalov/shortener/internal/database"
	"github.com/Adigezalov/shortener/internal/logger"
	"github.com/Adigezalov/shortener/internal/middleware"
	"github.com/Adigezalov/shortener/internal/models"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
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
		mockSetup      func(*MockURLStorage, *MockURLShortener)
		expectedStatus int
		expectedResult string
	}{
		{
			name: "Успешное_создание_нового_короткого_URL",
			request: models.ShortenRequest{
				URL: "https://example.com",
			},
			contentType: "application/json",
			mockSetup: func(ms *MockURLStorage, msh *MockURLShortener) {
				msh.On("Shorten", "https://example.com").Return("abc123")
				ms.On("AddWithUser", "abc123", "https://example.com", "test-user").Return("abc123", false, nil)
				msh.On("BuildShortURL", "abc123").Return("http://short.url/abc123")
			},
			expectedStatus: http.StatusCreated,
			expectedResult: "http://short.url/abc123",
		},
		{
			name: "URL_уже_существует_в_базе",
			request: models.ShortenRequest{
				URL: "https://example.com",
			},
			contentType: "application/json",
			mockSetup: func(ms *MockURLStorage, msh *MockURLShortener) {
				msh.On("Shorten", "https://example.com").Return("abc123")
				ms.On("AddWithUser", "abc123", "https://example.com", "test-user").Return("abc123", true, database.ErrURLConflict)
				msh.On("BuildShortURL", "abc123").Return("http://short.url/abc123")
			},
			expectedStatus: http.StatusConflict,
			expectedResult: "http://short.url/abc123",
		},
		{
			name: "Пустой_URL",
			request: models.ShortenRequest{
				URL: "",
			},
			contentType:    "application/json",
			mockSetup:      func(ms *MockURLStorage, msh *MockURLShortener) {},
			expectedStatus: http.StatusBadRequest,
			expectedResult: "",
		},
		{
			name: "Неверный_Content-Type",
			request: models.ShortenRequest{
				URL: "https://example.com",
			},
			contentType:    "text/plain",
			mockSetup:      func(ms *MockURLStorage, msh *MockURLShortener) {},
			expectedStatus: http.StatusUnsupportedMediaType,
			expectedResult: "",
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
			req := httptest.NewRequest("POST", "/api/shorten", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", tt.contentType)

			// Добавляем userID в контекст для тестов, которые требуют аутентификации
			if tt.expectedStatus == http.StatusCreated || tt.expectedStatus == http.StatusConflict {
				ctx := context.WithValue(req.Context(), middleware.UserIDKey, "test-user")
				req = req.WithContext(ctx)
			}

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
