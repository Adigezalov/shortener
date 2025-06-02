package handlers

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

// MockDB мок для базы данных
type MockDB struct {
	pingErr error
}

func (m *MockDB) Ping() error {
	return m.pingErr
}

func (m *MockDB) Close() error {
	return nil
}

func TestHandler_PingDB(t *testing.T) {
	tests := []struct {
		name           string
		pingErr        error
		expectedStatus int
	}{
		{
			name:           "успешное_подключение",
			pingErr:        nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "ошибка_подключения",
			pingErr:        errors.New("connection error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем мок базы данных
			mockDB := &MockDB{pingErr: tt.pingErr}

			// Создаем обработчик с моком
			h := &Handler{
				db: mockDB,
			}

			// Создаем тестовый запрос
			req := httptest.NewRequest(http.MethodGet, "/ping", nil)
			w := httptest.NewRecorder()

			// Выполняем запрос
			h.PingDB(w, req)

			// Проверяем результат
			result := w.Result()
			defer result.Body.Close()

			// Проверяем статус ответа
			assert.Equal(t, tt.expectedStatus, result.StatusCode)
		})
	}
}
