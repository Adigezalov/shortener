package database

import (
	"database/sql"
	_ "github.com/lib/pq"
)

// DBInterface описывает интерфейс для работы с базой данных
type DBInterface interface {
	Ping() error
	Close() error
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

// New создает новое подключение к базе данных
func New(dsn string) (*DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	// Проверяем подключение
	if err = db.Ping(); err != nil {
		return nil, err
	}

	return &DB{DB: db}, nil
}

// Ping проверяет подключение к базе данных
func (db *DB) Ping() error {
	return db.DB.Ping()
}

// Close закрывает соединение с базой данных
func (db *DB) Close() error {
	return db.DB.Close()
}
