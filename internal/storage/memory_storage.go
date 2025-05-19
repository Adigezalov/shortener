package storage

// NewMemoryStorage создает новое in-memory хранилище
func NewMemoryStorage(filePath string) *MemoryStorage {
	storage := &MemoryStorage{
		urlStore:     make(map[string]string),
		reverseStore: make(map[string]string),
	}

	// Загружаем данные из файла, если он существует
	records, err := LoadFromFile(filePath)
	if err != nil {
		return nil
	}

	// Заполняем хранилище данными из файла
	for _, record := range records {
		storage.urlStore[record.UUID] = record.OriginalURL
		storage.reverseStore[record.OriginalURL] = record.UUID
	}

	return storage
}

// Save сохраняет соответствие shortID -> originalURL
func (s *MemoryStorage) Save(shortID, originalURL string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.urlStore[shortID] = originalURL
	s.reverseStore[originalURL] = shortID

	return nil
}

// Get возвращает оригинальный URL по shortID
func (s *MemoryStorage) Get(shortID string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	originalURL, exists := s.urlStore[shortID]

	return originalURL, exists
}

// Exists проверяет существование originalURL в хранилище
func (s *MemoryStorage) Exists(originalURL string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	shortID, exists := s.reverseStore[originalURL]

	return shortID, exists
}
