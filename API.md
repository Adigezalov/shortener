# API Документация

Сервис сокращения URL предоставляет REST API для создания, получения и управления короткими ссылками.

## Базовый URL

```
http://localhost:8080
```

## Аутентификация

Сервис использует cookie-based аутентификацию. При первом запросе автоматически создается пользователь и устанавливается cookie `user_id`.

## Эндпоинты

### 1. Создание короткого URL (Text/Plain)

Создает короткий URL из переданного оригинального URL.

**Запрос:**
```http
POST /
Content-Type: text/plain

https://example.com/very/long/url
```

**Ответы:**

- **201 Created** - URL успешно создан
  ```
  http://localhost:8080/abc12345
  ```

- **400 Bad Request** - Пустой или некорректный URL
- **409 Conflict** - URL уже существует (возвращает существующий)
- **500 Internal Server Error** - Внутренняя ошибка

### 2. Создание короткого URL (JSON)

Создает короткий URL через JSON API.

**Запрос:**
```http
POST /api/shorten
Content-Type: application/json

{
  "url": "https://example.com/very/long/url"
}
```

**Ответы:**

- **201 Created** - URL успешно создан
  ```json
  {
    "result": "http://localhost:8080/abc12345"
  }
  ```

- **400 Bad Request** - Некорректный JSON или пустой URL
- **409 Conflict** - URL уже существует
- **415 Unsupported Media Type** - Неправильный Content-Type

### 3. Пакетное создание URL

Создает несколько коротких URL за один запрос.

**Запрос:**
```http
POST /api/shorten/batch
Content-Type: application/json

[
  {
    "correlation_id": "req_1",
    "original_url": "https://example.com/page1"
  },
  {
    "correlation_id": "req_2", 
    "original_url": "https://example.com/page2"
  }
]
```

**Ответы:**

- **201 Created** - URL успешно созданы
  ```json
  [
    {
      "correlation_id": "req_1",
      "short_url": "http://localhost:8080/abc123"
    },
    {
      "correlation_id": "req_2",
      "short_url": "http://localhost:8080/def456"
    }
  ]
  ```

- **400 Bad Request** - Некорректный JSON

### 4. Получение оригинального URL

Перенаправляет на оригинальный URL по короткому идентификатору.

**Запрос:**
```http
GET /{id}
```

**Ответы:**

- **307 Temporary Redirect** - Перенаправление на оригинальный URL
  ```http
  Location: https://example.com/original/url
  ```

- **404 Not Found** - Короткий URL не найден
- **410 Gone** - URL был удален пользователем

### 5. Получение URL пользователя

Возвращает все URL, созданные текущим пользователем.

**Запрос:**
```http
GET /api/user/urls
Cookie: user_id=abc123...
```

**Ответы:**

- **200 OK** - Список URL пользователя
  ```json
  [
    {
      "short_url": "http://localhost:8080/abc123",
      "original_url": "https://example.com/page1"
    },
    {
      "short_url": "http://localhost:8080/def456", 
      "original_url": "https://example.com/page2"
    }
  ]
  ```

- **204 No Content** - У пользователя нет URL
- **401 Unauthorized** - Отсутствует аутентификация

### 6. Удаление URL пользователя

Помечает URL как удаленные (мягкое удаление).

**Запрос:**
```http
DELETE /api/user/urls
Content-Type: application/json
Cookie: user_id=abc123...

["abc123", "def456"]
```

**Ответы:**

- **202 Accepted** - Запрос на удаление принят
- **400 Bad Request** - Некорректный JSON
- **401 Unauthorized** - Отсутствует аутентификация

### 7. Проверка состояния БД

Проверяет доступность базы данных.

**Запрос:**
```http
GET /ping
```

**Ответы:**

- **200 OK** - База данных доступна
- **500 Internal Server Error** - База данных недоступна

### 8. Статистика сервиса (Internal)

Возвращает статистику сервиса: количество URL и пользователей.

**Запрос:**
```http
GET /api/internal/stats
X-Real-IP: 192.168.1.100
```

**Ответы:**

- **200 OK** - Статистика сервиса
  ```json
  {
    "urls": 150,
    "users": 25
  }
  ```

- **403 Forbidden** - IP-адрес клиента не входит в доверенную подсеть
- **500 Internal Server Error** - Внутренняя ошибка сервера

**Требования:**
- Эндпоинт доступен только для IP-адресов из доверенной подсети (настраивается через `trusted_subnet`)
- IP-адрес клиента передается в заголовке `X-Real-IP`
- Если `trusted_subnet` не настроен, доступ к эндпоинту запрещен

## Коды ошибок

| Код | Описание |
|-----|----------|
| 200 | OK - Запрос выполнен успешно |
| 201 | Created - Ресурс создан |
| 202 | Accepted - Запрос принят к обработке |
| 204 | No Content - Нет содержимого |
| 307 | Temporary Redirect - Временное перенаправление |
| 400 | Bad Request - Некорректный запрос |
| 401 | Unauthorized - Требуется аутентификация |
| 403 | Forbidden - Доступ запрещен (IP не в доверенной подсети) |
| 404 | Not Found - Ресурс не найден |
| 409 | Conflict - Конфликт (URL уже существует) |
| 410 | Gone - Ресурс удален |
| 415 | Unsupported Media Type - Неподдерживаемый тип контента |
| 500 | Internal Server Error - Внутренняя ошибка сервера |

## Примеры использования

### Создание и использование короткого URL

```bash
# 1. Создаем короткий URL
curl -X POST http://localhost:8080/ \
  -H "Content-Type: text/plain" \
  -d "https://example.com/very/long/url"

# Ответ: http://localhost:8080/abc12345

# 2. Используем короткий URL
curl -I http://localhost:8080/abc12345

# Ответ: 307 Temporary Redirect
# Location: https://example.com/very/long/url
```

### Пакетное создание URL

```bash
curl -X POST http://localhost:8080/api/shorten/batch \
  -H "Content-Type: application/json" \
  -d '[
    {
      "correlation_id": "1",
      "original_url": "https://example.com/page1"
    },
    {
      "correlation_id": "2",
      "original_url": "https://example.com/page2"
    }
  ]'
```

### Получение URL пользователя

```bash
# Сначала создаем URL (получаем cookie)
curl -c cookies.txt -X POST http://localhost:8080/ \
  -H "Content-Type: text/plain" \
  -d "https://example.com/test"

# Затем получаем список URL
curl -b cookies.txt http://localhost:8080/api/user/urls
```

### Получение статистики сервиса

```bash
# Запускаем сервер с доверенной подсетью
TRUSTED_SUBNET=127.0.0.1/32 ./shortener

# Получаем статистику (IP из доверенной подсети)
curl -H "X-Real-IP: 127.0.0.1" http://localhost:8080/api/internal/stats

# Ответ: {"urls": 150, "users": 25}

# Попытка доступа с запрещенного IP
curl -H "X-Real-IP: 192.168.1.1" http://localhost:8080/api/internal/stats
# Ответ: 403 Forbidden
```

## Middleware

Сервис использует следующие middleware:

- **Logging** - Логирование всех запросов
- **Recovery** - Восстановление после паники
- **Compression** - Gzip сжатие ответов
- **Authentication** - Автоматическая аутентификация пользователей
- **Request ID** - Генерация уникального ID для каждого запроса
- **IP Auth** - Проверка IP-адреса для внутренних эндпоинтов

## Конфигурация

Сервис настраивается через переменные окружения или флаги командной строки:

| Параметр | Переменная окружения | Флаг | По умолчанию | Описание |
|----------|---------------------|------|--------------|----------|
| Адрес сервера | `SERVER_ADDRESS` | `-a` | `:8080` | Адрес и порт HTTP сервера |
| Базовый URL | `BASE_URL` | `-b` | `http://localhost:8080` | Базовый URL для коротких ссылок |
| Файл хранения | `FILE_STORAGE_PATH` | `-f` | `storage.json` | Путь к файлу хранения |
| База данных | `DATABASE_DSN` | `-d` | - | Строка подключения к PostgreSQL |
| Доверенная подсеть | `TRUSTED_SUBNET` | `-t` | - | CIDR подсети для доступа к внутренним эндпоинтам |

## Хранение данных

Сервис поддерживает три типа хранения:

1. **PostgreSQL** - Основное хранилище (если указан `DATABASE_DSN`)
2. **Файловое хранилище** - JSON файл (если БД не настроена)
3. **In-Memory** - Хранение в памяти (для тестирования)

## Профилирование

При включении профилирования (`PROFILING_ENABLED=true`) доступны pprof endpoints:

- `http://localhost:6060/debug/pprof/` - Список профилей
- `http://localhost:6060/debug/pprof/heap` - Профиль памяти
- `http://localhost:6060/debug/pprof/profile` - Профиль CPU
- `http://localhost:6060/debug/pprof/goroutine` - Профиль горутин