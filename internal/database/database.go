package database

import (
	"database/sql"
	"embed"
	_ "github.com/lib/pq"
	"strings"
)

//go:embed schema.sql
var schemaFS embed.FS

// DBInterface описывает интерфейс для работы с базой данных
type DBInterface interface {
	Ping() error
	Close() error
	AddURL(shortID, originalURL string) (string, bool, error)
	GetURL(shortID string) (string, bool, error)
	FindByOriginalURL(originalURL string) (string, bool, error)
}

// DB представляет обертку над sql.DB с дополнительной функциональностью
type DB struct {
	*sql.DB
}

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

// New создает новое подключение к базе данных и инициализирует схему
func New(dsn string) (*DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	// Проверяем подключение
	if err = db.Ping(); err != nil {
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

	// Выполняем каждую команду
	for _, command := range commands {
		command = strings.TrimSpace(command)
		if command == "" {
			continue
		}
		if _, err := db.Exec(command); err != nil {
			return err
		}
	}

	return nil
}

// AddURL добавляет новый URL в базу данных
func (db *DB) AddURL(shortID, originalURL string) (string, bool, error) {
	// Проверяем, существует ли уже такой URL
	if existingID, exists, err := db.FindByOriginalURL(originalURL); err == nil && exists {
		return existingID, true, nil
	}

	// Добавляем новый URL
	_, err := db.Exec(`
		INSERT INTO urls (short_id, original_url)
		VALUES ($1, $2)
		ON CONFLICT (short_id) DO NOTHING`,
		shortID, originalURL)

	if err != nil {
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

	if err == sql.ErrNoRows {
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

	if err == sql.ErrNoRows {
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
