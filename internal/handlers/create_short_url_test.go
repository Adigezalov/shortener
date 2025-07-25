package handlers

import (
	"bytes"
	"context"
	"github.com/Adigezalov/shortener/internal/database"
	"github.com/Adigezalov/shortener/internal/logger"
	"github.com/Adigezalov/shortener/internal/middleware"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandler_CreateShortURL(t *testing.T) {
	// Инициализируем тестовый логгер
	testLogger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("Не удалось создать тестовый логгер: %v", err)
	}
	logger.Logger = testLogger
	defer logger.Logger.Sync()

	tests := []struct {
		name           string
		inputURL       string
		mockSetup      func(*MockURLStorage, *MockURLShortener)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:     "Успешное_создание_нового_короткого_URL",
			inputURL: "https://example.com",
			mockSetup: func(ms *MockURLStorage, msh *MockURLShortener) {
				msh.On("Shorten", "https://example.com").Return("abc123")
				ms.On("AddWithUser", "abc123", "https://example.com", "test-user").Return("abc123", false, nil)
				msh.On("BuildShortURL", "abc123").Return("http://short.url/abc123")
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   "http://short.url/abc123",
		},
		{
			name:     "URL_уже_существует_в_базе",
			inputURL: "https://example.com",
			mockSetup: func(ms *MockURLStorage, msh *MockURLShortener) {
				msh.On("Shorten", "https://example.com").Return("abc123")
				ms.On("AddWithUser", "abc123", "https://example.com", "test-user").Return("abc123", true, database.ErrURLConflict)
				msh.On("BuildShortURL", "abc123").Return("http://short.url/abc123")
			},
			expectedStatus: http.StatusConflict,
			expectedBody:   "http://short.url/abc123",
		},
		{
			name:           "Пустой_URL",
			inputURL:       "",
			mockSetup:      func(ms *MockURLStorage, msh *MockURLShortener) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "URL не может быть пустым\n",
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

			// Создаем тестовый запрос
			req := httptest.NewRequest("POST", "/", bytes.NewBufferString(tt.inputURL))
			
			// Добавляем userID в контекст для тестов, которые не проверяют пустой URL
			if tt.inputURL != "" {
				ctx := context.WithValue(req.Context(), middleware.UserIDKey, "test-user")
				req = req.WithContext(ctx)
			}
			
			w := httptest.NewRecorder()

			// Вызываем тестируемый обработчик
			handler.CreateShortURL(w, req)

			// Проверяем результаты
			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, tt.expectedBody, w.Body.String())

			// Проверяем, что все ожидаемые вызовы моков были выполнены
			mockStorage.AssertExpectations(t)
			mockShortener.AssertExpectations(t)
		})
	}
}
