package database

// MockDB мок для базы данных, используется в тестах
type MockDB struct{}

func (m *MockDB) Ping() error {
	return nil
}

func (m *MockDB) Close() error {
	return nil
}

func (m *MockDB) AddURL(shortID, originalURL string) (string, bool, error) {
	// Для тестов всегда возвращаем успешное добавление
	return shortID, false, nil
}

func (m *MockDB) GetURL(shortID string) (string, bool, error) {
	// Для тестов возвращаем фиктивный URL
	return "https://example.com", true, nil
}

func (m *MockDB) FindByOriginalURL(originalURL string) (string, bool, error) {
	// Для тестов возвращаем фиктивный ID
	return "abc123", true, nil
}
