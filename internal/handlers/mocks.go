package handlers

import (
	"github.com/stretchr/testify/mock"
)

// MockStorage - мок для интерфейса хранилища
type MockStorage struct {
	mock.Mock
}

func (m *MockStorage) Add(id, url string) (string, bool, error) {
	args := m.Called(id, url)
	return args.String(0), args.Bool(1), args.Error(2)
}

func (m *MockStorage) Get(id string) (string, bool) {
	args := m.Called(id)
	return args.String(0), args.Bool(1)
}

func (m *MockStorage) FindByOriginalURL(url string) (string, bool) {
	args := m.Called(url)
	return args.String(0), args.Bool(1)
}

func (m *MockStorage) Close() error {
	args := m.Called()
	return args.Error(0)
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

// MockPinger - мок для интерфейса проверки подключения к базе данных
type MockPinger struct {
	mock.Mock
}

func (m *MockPinger) Ping() error {
	args := m.Called()
	return args.Error(0)
}
