package storage

import (
	"github.com/Adigezalov/shortener/internal/database"
)

// Factory создает хранилище URL в зависимости от конфигурации
func Factory(dbDSN, filePath string) (URLStorage, error) {
	// Пробуем создать хранилище в PostgreSQL
	if dbDSN != "" {
		db, err := database.New(dbDSN)
		if err != nil {
			return nil, err
		}
		return NewDBStorage(db), nil
	}

	// Если нет DSN, но есть путь к файлу, используем файловое хранилище
	if filePath != "" {
		return NewFileStorage(filePath), nil
	}

	// По умолчанию используем хранилище в памяти
	return NewMemoryStorage(), nil
}
