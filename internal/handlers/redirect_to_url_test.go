package handlers

import (
	"github.com/Adigezalov/shortener/internal/logger"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandler_RedirectToURL(t *testing.T) {
	// Инициализируем тестовый логгер
	testLogger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("Не удалось создать тестовый логгер: %v", err)
	}
	logger.Logger = testLogger
	defer logger.Logger.Sync()

	tests := []struct {
		name           string
		urlID          string
		mockSetup      func(*MockStorage)
		expectedStatus int
		expectedURL    string
	}{
		{
			name:  "Успешное перенаправление",
			urlID: "abc123",
			mockSetup: func(ms *MockStorage) {
				ms.On("Get", "abc123").Return("https://example.com", true)
			},
			expectedStatus: http.StatusTemporaryRedirect,
			expectedURL:    "https://example.com",
		},
		{
			name:  "URL не найден",
			urlID: "notfound",
			mockSetup: func(ms *MockStorage) {
				ms.On("Get", "notfound").Return("", false)
			},
			expectedStatus: http.StatusNotFound,
			expectedURL:    "",
		},
		{
			name:           "Некорректный ID",
			urlID:          "//", // Некорректный ID, который вызовет 404 в Chi router
			mockSetup:      func(ms *MockStorage) {},
			expectedStatus: http.StatusNotFound,
			expectedURL:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем мок хранилища
			mockStorage := new(MockStorage)

			// Настраиваем мок
			tt.mockSetup(mockStorage)

			// Создаем обработчик с моком
			handler := &Handler{
				storage: mockStorage,
			}

			// Создаем роутер Chi для тестирования с параметрами URL
			r := chi.NewRouter()
			r.Get("/{id}", handler.RedirectToURL)

			// Создаем тестовый запрос
			req := httptest.NewRequest("GET", "/"+tt.urlID, nil)
			w := httptest.NewRecorder()

			// Выполняем запрос
			r.ServeHTTP(w, req)

			// Проверяем результаты
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedURL != "" {
				assert.Equal(t, tt.expectedURL, w.Header().Get("Location"))
			}

			// Проверяем, что все ожидаемые вызовы мока были выполнены
			mockStorage.AssertExpectations(t)
		})
	}
}
