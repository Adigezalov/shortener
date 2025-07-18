package handlers

import (
	"github.com/Adigezalov/shortener/internal/models"
	"github.com/stretchr/testify/mock"
)

// MockURLStorage - мок для интерфейса хранилища
type MockURLStorage struct {
	mock.Mock
}

func (m *MockURLStorage) Add(id, url string) (string, bool, error) {
	args := m.Called(id, url)
	return args.String(0), args.Bool(1), args.Error(2)
}

func (m *MockURLStorage) AddWithUser(id, url, userID string) (string, bool, error) {
	args := m.Called(id, url, userID)
	return args.String(0), args.Bool(1), args.Error(2)
}

func (m *MockURLStorage) Get(id string) (string, bool) {
	args := m.Called(id)
	return args.String(0), args.Bool(1)
}

func (m *MockURLStorage) FindByOriginalURL(url string) (string, bool) {
	args := m.Called(url)
	return args.String(0), args.Bool(1)
}

func (m *MockURLStorage) GetUserURLs(userID string) ([]models.UserURL, error) {
	args := m.Called(userID)
	return args.Get(0).([]models.UserURL), args.Error(1)
}

func (m *MockURLStorage) Close() error {
	args := m.Called()
	return args.Error(0)
}

// MockURLShortener - мок для интерфейса сокращения URL
type MockURLShortener struct {
	mock.Mock
}

func (m *MockURLShortener) Shorten(url string) string {
	args := m.Called(url)
	return args.String(0)
}

func (m *MockURLShortener) BuildShortURL(id string) string {
	args := m.Called(id)
	return args.String(0)
}

// MockPinger - мок для интерфейса проверки подключения к базе данных
type MockPinger struct {
	mock.Mock
}

func (m *MockPinger) Ping() error {
	args := m.Called()
	return args.Error(0)
}
