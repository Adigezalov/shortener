package models

// ShortenRequest представляет запрос на сокращение URL
type ShortenRequest struct {
	URL string `json:"url"`
}

// ShortenResponse представляет ответ с сокращенным URL
type ShortenResponse struct {
	Result string `json:"result"`
}

// URLRecord представляет запись URL для сохранения
type URLRecord struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}
