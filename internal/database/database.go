package database

import (
	"database/sql"
	"embed"
	"errors"
	"github.com/jackc/pgerrcode"
	"github.com/lib/pq"
	"strings"
)

//go:embed schema.sql
var schemaFS embed.FS

// ErrURLConflict ошибка при попытке добавить существующий URL
var ErrURLConflict = errors.New("url already exists")

// DB представляет обертку над sql.DB с дополнительной функциональностью
type DB struct {
	*sql.DB
}

// New создает новое подключение к базе данных и инициализирует схему
func New(dsn string) (*DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	// Создаем экземпляр DB
	database := &DB{DB: db}

	// Инициализируем схему базы данных
	if err = database.initSchema(); err != nil {
		return nil, err
	}

	return database, nil
}

// initSchema инициализирует схему базы данных
func (db *DB) initSchema() error {
	// Читаем SQL-файл
	schemaSQL, err := schemaFS.ReadFile("schema.sql")
	if err != nil {
		return err
	}

	// Разделяем на отдельные команды
	commands := strings.Split(string(schemaSQL), ";")

	// Начинаем транзакцию
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	// Отложенный откат транзакции в случае ошибки
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Выполняем каждую команду в транзакции
	for _, command := range commands {
		command = strings.TrimSpace(command)
		if command == "" {
			continue
		}
		if _, err := tx.Exec(command); err != nil {
			return err
		}
	}

	// Подтверждаем транзакцию
	return tx.Commit()
}

// AddURL добавляет новый URL в базу данных
func (db *DB) AddURL(shortID, originalURL string) (string, bool, error) {
	// Пытаемся добавить новый URL
	_, err := db.Exec(`
		INSERT INTO urls (short_id, original_url)
		VALUES ($1, $2)`,
		shortID, originalURL)

	if err != nil {
		// Проверяем, является ли ошибка нарушением уникальности
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code == pgerrcode.UniqueViolation {
				// Если это нарушение уникальности original_url, получаем существующий short_id
				if strings.Contains(pqErr.Constraint, "original_url") {
					existingID, exists, err := db.FindByOriginalURL(originalURL)
					if err != nil {
						return "", false, err
					}
					if exists {
						return existingID, true, ErrURLConflict
					}
				}
			}
		}
		return "", false, err
	}

	return shortID, false, nil
}

// GetURL возвращает оригинальный URL по короткому идентификатору
func (db *DB) GetURL(shortID string) (string, bool, error) {
	var originalURL string
	err := db.QueryRow(`
		SELECT original_url
		FROM urls
		WHERE short_id = $1`,
		shortID).Scan(&originalURL)

	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}

	return originalURL, true, nil
}

// FindByOriginalURL ищет короткий идентификатор по оригинальному URL
func (db *DB) FindByOriginalURL(originalURL string) (string, bool, error) {
	var shortID string
	err := db.QueryRow(`
		SELECT short_id
		FROM urls
		WHERE original_url = $1`,
		originalURL).Scan(&shortID)

	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}

	return shortID, true, nil
}

// Ping проверяет подключение к базе данных
func (db *DB) Ping() error {
	return db.DB.Ping()
}

// Close закрывает соединение с базой данных
func (db *DB) Close() error {
	return db.DB.Close()
}
