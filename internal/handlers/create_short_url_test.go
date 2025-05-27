package handlers

import (
	"bytes"
	"github.com/Adigezalov/shortener/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"net/http"
	"net/http/httptest"
	"testing"
)

// MockStorage - мок для интерфейса хранилища
type MockStorage struct {
	mock.Mock
}

func (m *MockStorage) Add(id, url string) (string, bool) {
	args := m.Called(id, url)
	return args.String(0), args.Bool(1)
}

func (m *MockStorage) Get(id string) (string, bool) {
	args := m.Called(id)
	return args.String(0), args.Bool(1)
}

func (m *MockStorage) FindByOriginalURL(url string) (string, bool) {
	args := m.Called(url)
	return args.String(0), args.Bool(1)
}

// MockShortener - мок для интерфейса сокращения URL
type MockShortener struct {
	mock.Mock
}

func (m *MockShortener) Shorten(url string) string {
	args := m.Called(url)
	return args.String(0)
}

func (m *MockShortener) BuildShortURL(id string) string {
	args := m.Called(id)
	return args.String(0)
}

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
		mockSetup      func(*MockStorage, *MockShortener)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:     "Успешное создание нового короткого URL",
			inputURL: "https://example.com",
			mockSetup: func(ms *MockStorage, msh *MockShortener) {
				ms.On("FindByOriginalURL", "https://example.com").Return("", false)
				msh.On("Shorten", "https://example.com").Return("abc123")
				ms.On("Add", "abc123", "https://example.com").Return("abc123", false)
				msh.On("BuildShortURL", "abc123").Return("http://short.url/abc123")
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   "http://short.url/abc123",
		},
		{
			name:     "URL уже существует в базе",
			inputURL: "https://example.com",
			mockSetup: func(ms *MockStorage, msh *MockShortener) {
				ms.On("FindByOriginalURL", "https://example.com").Return("existing123", true)
				msh.On("BuildShortURL", "existing123").Return("http://short.url/existing123")
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   "http://short.url/existing123",
		},
		{
			name:           "Пустой URL",
			inputURL:       "",
			mockSetup:      func(ms *MockStorage, msh *MockShortener) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "URL не может быть пустым\n",
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

			// Создаем тестовый запрос
			req := httptest.NewRequest("POST", "/", bytes.NewBufferString(tt.inputURL))
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
