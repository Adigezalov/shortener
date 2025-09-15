# Документация проекта

## Обзор

Сервис сокращения URL предоставляет REST API для создания коротких ссылок из длинных URL. Проект написан на Go и поддерживает различные типы хранения данных.

## Архитектура

### Основные пакеты

- **handlers** - HTTP обработчики для всех эндпоинтов API
- **models** - Модели данных для запросов и ответов
- **config** - Конфигурация приложения
- **shortener** - Сервис генерации коротких идентификаторов
- **storage** - Слой хранения данных (память/файл/PostgreSQL)
- **middleware** - HTTP middleware (логирование, сжатие, аутентификация)

### Интерфейсы

#### URLStorage
Основной интерфейс для хранения URL:
```go
type URLStorage interface {
    Add(id string, url string) (string, bool, error)
    AddWithUser(id string, url string, userID string) (string, bool, error)
    Get(id string) (string, bool)
    FindByOriginalURL(url string) (string, bool)
    GetUserURLs(userID string) ([]models.UserURL, error)
    DeleteUserURLs(userID string, shortURLs []string) error
    IsDeleted(shortURL string) (bool, error)
    Close() error
}
```

#### URLShortener
Интерфейс для сокращения URL:
```go
type URLShortener interface {
    Shorten(url string) string
    BuildShortURL(id string) string
}
```

## Использование

### Базовый пример

```go
package main

import (
    "github.com/Adigezalov/shortener/internal/config"
    "github.com/Adigezalov/shortener/internal/shortener"
    "github.com/Adigezalov/shortener/internal/storage"
)

func main() {
    // Загружаем конфигурацию
    cfg := config.NewConfig()
    
    // Создаем хранилище
    store := storage.NewMemoryStorage("")
    defer store.Close()
    
    // Создаем сервис сокращения
    service := shortener.New(cfg.BaseURL)
    
    // Сокращаем URL
    shortID := service.Shorten("https://example.com/very/long/url")
    store.Add(shortID, "https://example.com/very/long/url")
    
    // Получаем полный короткий URL
    shortURL := service.BuildShortURL(shortID)
    // shortURL = "http://localhost:8080/abc12345"
}
```

### Конфигурация

Приложение настраивается через переменные окружения или флаги:

```bash
# Переменные окружения
export SERVER_ADDRESS=":8080"
export BASE_URL="https://short.ly"
export DATABASE_DSN="postgres://user:pass@localhost/db"

# Или флаги командной строки
./shortener -a :8080 -b https://short.ly -d "postgres://..."
```

### API Endpoints

#### POST / (Text/Plain)
Создание короткого URL через text/plain:
```bash
curl -X POST http://localhost:8080/ \
  -H "Content-Type: text/plain" \
  -d "https://example.com/long/url"
```

#### POST /api/shorten (JSON)
Создание короткого URL через JSON API:
```bash
curl -X POST http://localhost:8080/api/shorten \
  -H "Content-Type: application/json" \
  -d '{"url": "https://example.com/long/url"}'
```

#### POST /api/shorten/batch (JSON)
Пакетное создание коротких URL:
```bash
curl -X POST http://localhost:8080/api/shorten/batch \
  -H "Content-Type: application/json" \
  -d '[
    {"correlation_id": "1", "original_url": "https://example.com/1"},
    {"correlation_id": "2", "original_url": "https://example.com/2"}
  ]'
```

#### GET /{id}
Редирект на оригинальный URL:
```bash
curl -I http://localhost:8080/abc12345
# HTTP/1.1 307 Temporary Redirect
# Location: https://example.com/long/url
```

## Хранение данных

### Типы хранилищ

1. **PostgreSQL** - Основное хранилище для production
2. **Файловое** - JSON файл для development
3. **In-Memory** - Хранение в памяти для тестирования

### Выбор хранилища

Хранилище выбирается автоматически:
- Если указан `DATABASE_DSN` - используется PostgreSQL
- Иначе используется файловое хранилище или память

## Профилирование

### Включение профилирования

```bash
export PROFILING_ENABLED=true
export PROFILING_PORT=:6060
./shortener
```

### Доступные профили

- `http://localhost:6060/debug/pprof/` - Список профилей
- `http://localhost:6060/debug/pprof/heap` - Профиль памяти
- `http://localhost:6060/debug/pprof/profile` - Профиль CPU
- `http://localhost:6060/debug/pprof/goroutine` - Профиль горутин

### Сбор профилей

```bash
# Автоматический сбор всех профилей
make collect-profiles

# Сравнение профилей до и после оптимизации
make compare-profiles
```

## Разработка

### Команды разработчика

```bash
make build              # Сборка приложения
make test               # Запуск тестов
make fmt                # Форматирование кода
make lint               # Линтинг кода
make check              # Комплексная проверка
make doc                # Генерация документации
make profile            # Профилирование производительности
```

### Структура проекта

```
├── cmd/shortener/          # Точка входа приложения
├── internal/               # Внутренние пакеты
│   ├── config/            # Конфигурация
│   ├── handlers/          # HTTP обработчики
│   ├── middleware/        # HTTP middleware
│   ├── models/            # Модели данных
│   ├── shortener/         # Сервис сокращения URL
│   ├── storage/           # Слой хранения
│   └── profiling/         # Профилирование
├── examples/              # Примеры использования
├── benchmarks/            # Бенчмарки и профили
└── scripts/               # Скрипты автоматизации
```

### Тестирование

```bash
# Все тесты
go test ./...

# Тесты с покрытием
go test -cover ./...

# Бенчмарки
go test -bench=. ./benchmarks/...

# Примеры
go test ./examples/
```

## Примеры

Полные примеры использования API доступны в директории `examples/`:

```bash
# Запуск примеров
go test ./examples/ -v

# Просмотр документации с примерами
make doc
```

## Документация API

Подробная документация API доступна в файле [API.md](API.md).

## Производительность

Результаты профилирования и оптимизации доступны в [benchmarks/profiles/optimization_report.md](benchmarks/profiles/optimization_report.md).

## Лицензия

Проект разработан в рамках практического трека Яндекс.Практикума.