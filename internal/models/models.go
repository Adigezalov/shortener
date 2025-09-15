// Package models содержит модели данных для сервиса сокращения URL.
//
// Пакет определяет структуры для HTTP API запросов и ответов,
// а также внутренние модели для работы с хранилищем данных.
package models

// ShortenRequest представляет запрос на сокращение URL через JSON API.
//
// Используется в эндпоинте POST /api/shorten для получения URL,
// который необходимо сократить.
//
// Пример JSON:
//
//	{
//	  "url": "https://example.com/very/long/url"
//	}
type ShortenRequest struct {
	URL string `json:"url"` // URL для сокращения
}

// ShortenResponse представляет ответ с сокращенным URL через JSON API.
//
// Возвращается эндпоинтом POST /api/shorten в случае успешного
// создания короткого URL.
//
// Пример JSON:
//
//	{
//	  "result": "http://localhost:8080/abc123"
//	}
type ShortenResponse struct {
	Result string `json:"result"` // Сокращенный URL
}

// BatchShortenRequest представляет элемент запроса на пакетное сокращение URL.
//
// Используется в эндпоинте POST /api/shorten/batch для массового
// создания коротких URL. Каждый элемент содержит correlation_id
// для связи запроса с ответом.
//
// Пример JSON элемента:
//
//	{
//	  "correlation_id": "user_request_1",
//	  "original_url": "https://example.com/page1"
//	}
type BatchShortenRequest struct {
	CorrelationID string `json:"correlation_id"` // Идентификатор для связи запроса с ответом
	OriginalURL   string `json:"original_url"`   // Оригинальный URL для сокращения
}

// BatchShortenResponse представляет элемент ответа на пакетное сокращение URL.
//
// Возвращается эндпоинтом POST /api/shorten/batch для каждого
// успешно обработанного URL из запроса.
//
// Пример JSON элемента:
//
//	{
//	  "correlation_id": "user_request_1",
//	  "short_url": "http://localhost:8080/abc123"
//	}
type BatchShortenResponse struct {
	CorrelationID string `json:"correlation_id"` // Идентификатор из соответствующего запроса
	ShortURL      string `json:"short_url"`      // Созданный короткий URL
}

// URLRecord представляет запись URL для сохранения в файловом хранилище.
//
// Используется для сериализации данных в JSON формат при сохранении
// в файл. Содержит всю необходимую информацию для восстановления URL.
type URLRecord struct {
	UUID        string `json:"uuid"`         // Уникальный идентификатор пользователя
	ShortURL    string `json:"short_url"`    // Короткий идентификатор URL
	OriginalURL string `json:"original_url"` // Оригинальный URL
}

// UserURL представляет URL пользователя для API ответов.
//
// Используется в эндпоинте GET /api/user/urls для возврата
// списка всех URL, созданных пользователем.
//
// Пример JSON элемента:
//
//	{
//	  "short_url": "http://localhost:8080/abc123",
//	  "original_url": "https://example.com/page1"
//	}
type UserURL struct {
	ShortURL    string `json:"short_url"`    // Короткий URL
	OriginalURL string `json:"original_url"` // Оригинальный URL
}

// URLStorage представляет запись URL в базе данных.
//
// Используется для маппинга данных из PostgreSQL таблицы urls.
// Содержит информацию о пользователе, URL и статусе удаления.
type URLStorage struct {
	UUID        string `db:"user_id"`      // ID пользователя
	ShortURL    string `db:"short_url"`    // Короткий идентификатор
	OriginalURL string `db:"original_url"` // Оригинальный URL
	DeletedFlag bool   `db:"is_deleted"`   // Флаг удаления (мягкое удаление)
}
