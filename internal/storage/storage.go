package storage

import (
	"sync"
)

// Storage определяет интерфейс для работы с хранилищем URL
type Storage interface {
	// Save сохраняет соответствие короткого ID и оригинального URL
	Save(shortID, originalURL string) error

	// Get возвращает оригинальный URL по короткому ID и флаг наличия записи
	Get(shortID string) (string, bool)

	// Exists проверяет, существует ли уже данный оригинальный URL в хранилище
	// Возвращает короткий ID если URL уже существует
	Exists(originalURL string) (string, bool)
}

// MemoryStorage реализует Storage используя map в памяти
type MemoryStorage struct {
	urlStore     map[string]string // shortID -> originalURL
	reverseStore map[string]string // originalURL -> shortID
	mu           sync.RWMutex
}

// NewMemoryStorage создает новое in-memory хранилище
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		urlStore:     make(map[string]string),
		reverseStore: make(map[string]string),
	}
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
