package storage

import (
	"database/sql"
	"github.com/Adigezalov/shortener/internal/database"
	"github.com/jackc/pgerrcode"
	"github.com/lib/pq"
)

// DatabaseStorage реализует хранилище URL в PostgreSQL
type DatabaseStorage struct {
	db *database.DB
}

// NewDatabaseStorage создает новое хранилище URL в PostgreSQL
func NewDatabaseStorage(db *database.DB) *DatabaseStorage {
	return &DatabaseStorage{
		db: db,
	}
}

// Add добавляет новый URL в хранилище
func (s *DatabaseStorage) Add(id string, url string) (string, bool, error) {
	// Проверяем, существует ли уже такой URL
	if existingID, exists := s.FindByOriginalURL(url); exists {
		return existingID, true, database.ErrURLConflict
	}

	// Добавляем новый URL
	_, err := s.db.Exec(`
		INSERT INTO urls (short_id, original_url)
		VALUES ($1, $2)
	`, id, url)

	if err != nil {
		// Проверяем, является ли ошибка нарушением уникальности
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == pgerrcode.UniqueViolation {
			// Если произошел конфликт, проверяем по какому полю
			existingID, exists := s.FindByOriginalURL(url)
			if exists {
				// Конфликт по original_url - возвращаем существующий ID
				return existingID, true, nil
			}
			// Конфликт по short_id - возвращаем ошибку конфликта
			return "", false, database.ErrURLConflict
		}
		return "", false, err
	}

	return id, false, nil
}

// Get возвращает оригинальный URL по идентификатору
func (s *DatabaseStorage) Get(id string) (string, bool) {
	var url string
	err := s.db.QueryRow(`
		SELECT original_url
		FROM urls
		WHERE short_id = $1
	`, id).Scan(&url)

	if err == sql.ErrNoRows {
		return "", false
	}

	if err != nil {
		return "", false
	}

	return url, true
}

// FindByOriginalURL ищет ID по оригинальному URL
func (s *DatabaseStorage) FindByOriginalURL(url string) (string, bool) {
	var id string
	err := s.db.QueryRow(`
		SELECT short_id
		FROM urls
		WHERE original_url = $1
	`, url).Scan(&id)

	if err == sql.ErrNoRows {
		return "", false
	}

	if err != nil {
		return "", false
	}

	return id, true
}

// Close закрывает соединение с базой данных
func (s *DatabaseStorage) Close() error {
	return nil // DB закрывается на уровне приложения
}
