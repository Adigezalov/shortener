package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Adigezalov/shortener/internal/middleware"
	"github.com/Adigezalov/shortener/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestHandler_GetUserURLs(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		userURLs       []models.UserURL
		storageError   error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:   "успешное_получение_URL_пользователя",
			userID: "user123",
			userURLs: []models.UserURL{
				{ShortURL: "abc123", OriginalURL: "https://example.com"},
				{ShortURL: "def456", OriginalURL: "https://google.com"},
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `[{"short_url":"http://localhost:8080/abc123","original_url":"https://example.com"},{"short_url":"http://localhost:8080/def456","original_url":"https://google.com"}]`,
		},
		{
			name:           "пользователь_без_URL",
			userID:         "user456",
			userURLs:       []models.UserURL{},
			expectedStatus: http.StatusNoContent,
			expectedBody:   "",
		},
		{
			name:           "отсутствие_ID_пользователя_в_контексте",
			userID:         "",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Unauthorized\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем мок хранилища
			mockStorage := &MockURLStorage{}
			mockShortener := &MockURLShortener{}

			// Настраиваем мок
			if tt.userID != "" {
				mockStorage.On("GetUserURLs", tt.userID).Return(tt.userURLs, tt.storageError)
				for _, userURL := range tt.userURLs {
					mockShortener.On("BuildShortURL", userURL.ShortURL).Return("http://localhost:8080/" + userURL.ShortURL)
				}
			}

			// Создаем хендлер
			handler := New(mockStorage, mockShortener, nil)

			// Создаем запрос
			req := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)

			// Добавляем userID в контекст, если он есть
			if tt.userID != "" {
				ctx := context.WithValue(req.Context(), middleware.UserIDKey, tt.userID)
				req = req.WithContext(ctx)
			}

			// Создаем ResponseRecorder
			w := httptest.NewRecorder()

			// Вызываем хендлер
			handler.GetUserURLs(w, req)

			// Проверяем статус код
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Проверяем тело ответа
			if tt.expectedStatus == http.StatusOK {
				var response []models.UserURL
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Len(t, response, len(tt.userURLs))

				for i, userURL := range response {
					assert.Equal(t, "http://localhost:8080/"+tt.userURLs[i].ShortURL, userURL.ShortURL)
					assert.Equal(t, tt.userURLs[i].OriginalURL, userURL.OriginalURL)
				}
			} else if tt.expectedStatus == http.StatusNoContent {
				assert.Empty(t, w.Body.String())
			} else {
				assert.Equal(t, tt.expectedBody, w.Body.String())
			}

			// Проверяем, что все ожидания мока выполнены
			mockStorage.AssertExpectations(t)
			if tt.userID != "" && len(tt.userURLs) > 0 {
				mockShortener.AssertExpectations(t)
			}
		})
	}
}
