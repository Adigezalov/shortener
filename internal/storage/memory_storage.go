package storage

import "sync"

// MemoryStorage реализует хранилище URL в памяти
type MemoryStorage struct {
	urls    map[string]string // id -> original_url
	urlToID map[string]string // original_url -> id
	mu      sync.RWMutex
}

// NewMemoryStorage создает новое хранилище URL в памяти
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		urls:    make(map[string]string),
		urlToID: make(map[string]string),
	}
}

// Add добавляет новый URL в хранилище
func (s *MemoryStorage) Add(id string, url string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Проверяем, есть ли уже такой URL
	if existingID, found := s.urlToID[url]; found {
		return existingID, true
	}

	// Добавляем новый URL
	s.urls[id] = url
	s.urlToID[url] = id

	return id, false
}

// Get возвращает оригинальный URL по идентификатору
func (s *MemoryStorage) Get(id string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	url, ok := s.urls[id]
	return url, ok
}

// FindByOriginalURL ищет ID по оригинальному URL
func (s *MemoryStorage) FindByOriginalURL(url string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	id, ok := s.urlToID[url]
	return id, ok
}

// Close ничего не делает для хранилища в памяти
func (s *MemoryStorage) Close() error {
	return nil
}
