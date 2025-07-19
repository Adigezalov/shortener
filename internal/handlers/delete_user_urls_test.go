package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Adigezalov/shortener/internal/logger"
	"github.com/Adigezalov/shortener/internal/middleware"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestHandler_DeleteUserURLs(t *testing.T) {
	// Инициализируем тестовый логгер
	testLogger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("Не удалось создать тестовый логгер: %v", err)
	}
	logger.Logger = testLogger
	defer logger.Logger.Sync()

	tests := []struct {
		name           string
		userID         string
		requestBody    []string
		mockSetup      func(*MockURLStorage)
		expectedStatus int
	}{
		{
			name:        "успешное_удаление_URL",
			userID:      "user123",
			requestBody: []string{"abc123", "def456"},
			mockSetup: func(ms *MockURLStorage) {
				// Настраиваем мок для асинхронной операции
				ms.On("DeleteUserURLs", "user123", []string{"abc123", "def456"}).Return(nil)
			},
			expectedStatus: http.StatusAccepted,
		},
		{
			name:           "пустой_список_URL",
			userID:         "user123",
			requestBody:    []string{},
			mockSetup:      func(ms *MockURLStorage) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "отсутствие_ID_пользователя_в_контексте",
			userID:         "",
			requestBody:    []string{"abc123"},
			mockSetup:      func(ms *MockURLStorage) {},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем мок хранилища
			mockStorage := &MockURLStorage{}

			// Настраиваем мок
			tt.mockSetup(mockStorage)

			// Создаем хендлер
			handler := New(mockStorage, nil, nil)

			// Создаем тело запроса
			body, _ := json.Marshal(tt.requestBody)

			// Создаем запрос
			req := httptest.NewRequest(http.MethodDelete, "/api/user/urls", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			// Добавляем userID в контекст, если он есть
			if tt.userID != "" {
				ctx := context.WithValue(req.Context(), middleware.UserIDKey, tt.userID)
				req = req.WithContext(ctx)
			}

			// Создаем ResponseRecorder
			w := httptest.NewRecorder()

			// Вызываем хендлер
			handler.DeleteUserURLs(w, req)

			// Проверяем статус код
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Для успешного удаления не проверяем мок, так как операция асинхронная
			if tt.expectedStatus != http.StatusAccepted {
				mockStorage.AssertExpectations(t)
			}
		})
	}
}
