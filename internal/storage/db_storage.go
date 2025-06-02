package storage

import (
	"github.com/Adigezalov/shortener/internal/database"
	"github.com/Adigezalov/shortener/internal/logger"
	"go.uber.org/zap"
)

// DBStorage реализует хранилище URL в PostgreSQL
type DBStorage struct {
	db database.DBInterface
}

// NewDBStorage создает новое хранилище URL в PostgreSQL
func NewDBStorage(db database.DBInterface) *DBStorage {
	return &DBStorage{
		db: db,
	}
}

// Add добавляет новый URL в хранилище
func (s *DBStorage) Add(id string, url string) (string, bool) {
	shortID, exists, err := s.db.AddURL(id, url)
	if err != nil {
		logger.Logger.Error("Ошибка добавления URL в базу данных",
			zap.String("id", id),
			zap.String("url", url),
			zap.Error(err))
		return "", false
	}
	return shortID, exists
}

// Get возвращает оригинальный URL по идентификатору
func (s *DBStorage) Get(id string) (string, bool) {
	url, exists, err := s.db.GetURL(id)
	if err != nil {
		return "", false
	}
	return url, exists
}

// FindByOriginalURL ищет ID по оригинальному URL
func (s *DBStorage) FindByOriginalURL(url string) (string, bool) {
	id, exists, err := s.db.FindByOriginalURL(url)
	if err != nil {
		return "", false
	}
	return id, exists
}

// Close закрывает соединение с базой данных
func (s *DBStorage) Close() error {
	return s.db.Close()
}
