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
		return NewDatabaseStorage(db), nil
	}

	// Если нет DSN, создаем хранилище в памяти с опциональным сохранением в файл
	return NewMemoryStorage(filePath), nil
}
