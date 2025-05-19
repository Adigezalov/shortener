package storage

import "sync"

type URLGetter interface {
	Get(shortID string) (string, bool)
}

type URLSaver interface {
	Save(shortID, originalURL string) error
	Exists(originalURL string) (string, bool)
}

// URLReadWriter объединяет интерфейсы для полного доступа
type URLReadWriter interface {
	URLGetter
	URLSaver
}

// MemoryStorage реализует Storage используя map в памяти
type MemoryStorage struct {
	urlStore     map[string]string // shortID -> originalURL
	reverseStore map[string]string // originalURL -> shortID
	mu           sync.RWMutex
}

// URLRecord описывает структуру сокращённой ссылки
type URLRecord struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}
