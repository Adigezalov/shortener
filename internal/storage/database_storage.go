package storage

import (
	"database/sql"
	"github.com/Adigezalov/shortener/internal/database"
	"github.com/Adigezalov/shortener/internal/models"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
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
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == pgerrcode.UniqueViolation {
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

// AddWithUser добавляет новый URL в хранилище с привязкой к пользователю
func (s *DatabaseStorage) AddWithUser(id string, url string, userID string) (string, bool, error) {
	// Проверяем, существует ли уже такой URL
	if existingID, exists := s.FindByOriginalURL(url); exists {
		return existingID, true, database.ErrURLConflict
	}

	// Добавляем новый URL с привязкой к пользователю
	_, err := s.db.Exec(`
		INSERT INTO urls (short_id, original_url, user_id)
		VALUES ($1, $2, $3)
	`, id, url, userID)

	if err != nil {
		// Проверяем, является ли ошибка нарушением уникальности
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == pgerrcode.UniqueViolation {
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
	var isDeleted bool
	err := s.db.QueryRow(`
		SELECT original_url, COALESCE(is_deleted, false)
		FROM urls
		WHERE short_id = $1
	`, id).Scan(&url, &isDeleted)

	if err == sql.ErrNoRows {
		return "", false
	}

	if err != nil {
		return "", false
	}

	// Если URL помечен как удаленный, возвращаем false
	if isDeleted {
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

// GetUserURLs возвращает все URL пользователя (исключая удаленные)
func (s *DatabaseStorage) GetUserURLs(userID string) ([]models.UserURL, error) {
	rows, err := s.db.Query(`
		SELECT short_id, original_url
		FROM urls
		WHERE user_id = $1 AND COALESCE(is_deleted, false) = false
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []models.UserURL
	for rows.Next() {
		var shortID, originalURL string
		if err := rows.Scan(&shortID, &originalURL); err != nil {
			return nil, err
		}
		result = append(result, models.UserURL{
			ShortURL:    shortID,
			OriginalURL: originalURL,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

// DeleteUserURLs помечает URL как удаленные для указанного пользователя
func (s *DatabaseStorage) DeleteUserURLs(userID string, shortURLs []string) error {
	if len(shortURLs) == 0 {
		return nil
	}

	// Помечаем URL как удаленные вместо физического удаления
	query := `
		UPDATE urls 
		SET is_deleted = true
		WHERE user_id = $1 AND short_id = ANY($2)
	`

	_, err := s.db.Exec(query, userID, shortURLs)
	return err
}

// IsDeleted проверяет, помечен ли URL как удаленный
func (s *DatabaseStorage) IsDeleted(shortURL string) (bool, error) {
	var isDeleted bool
	err := s.db.QueryRow(`
		SELECT COALESCE(is_deleted, false)
		FROM urls
		WHERE short_id = $1
	`, shortURL).Scan(&isDeleted)

	if err == sql.ErrNoRows {
		return false, nil // URL не найден, считаем не удаленным
	}

	if err != nil {
		return false, err
	}

	return isDeleted, nil
}

// Close закрывает соединение с базой данных
func (s *DatabaseStorage) Close() error {
	return nil // DB закрывается на уровне приложения
}
