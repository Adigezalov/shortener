package utils

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// DataGenerator генератор тестовых данных для бенчмарков
type DataGenerator struct {
	rand *rand.Rand
}

// NewDataGenerator создает новый генератор данных
func NewDataGenerator() *DataGenerator {
	return &DataGenerator{
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// GenerateURL генерирует случайный URL
func (dg *DataGenerator) GenerateURL() string {
	domains := []string{
		"example.com", "test.org", "sample.net", "demo.io", 
		"benchmark.dev", "performance.test", "speed.check",
	}
	
	paths := []string{
		"", "/page", "/article", "/post", "/content", "/data",
		"/api/v1/resource", "/users/profile", "/products/item",
	}
	
	domain := domains[dg.rand.Intn(len(domains))]
	path := paths[dg.rand.Intn(len(paths))]
	
	// Иногда добавляем параметры запроса
	if dg.rand.Float32() < 0.3 {
		params := dg.generateQueryParams()
		if params != "" {
			path += "?" + params
		}
	}
	
	return fmt.Sprintf("https://%s%s", domain, path)
}

// GenerateURLs генерирует массив случайных URL
func (dg *DataGenerator) GenerateURLs(count int) []string {
	urls := make([]string, count)
	for i := 0; i < count; i++ {
		urls[i] = dg.GenerateURL()
	}
	return urls
}

// GenerateLongURL генерирует длинный URL для тестирования
func (dg *DataGenerator) GenerateLongURL() string {
	baseURL := dg.GenerateURL()
	
	// Добавляем много параметров для создания длинного URL
	params := make([]string, 10+dg.rand.Intn(20))
	for i := range params {
		key := dg.generateRandomString(5 + dg.rand.Intn(10))
		value := dg.generateRandomString(10 + dg.rand.Intn(50))
		params[i] = fmt.Sprintf("%s=%s", key, value)
	}
	
	return baseURL + "?" + strings.Join(params, "&")
}

// GenerateUserID генерирует случайный ID пользователя
func (dg *DataGenerator) GenerateUserID() string {
	return fmt.Sprintf("user-%d-%s", 
		dg.rand.Intn(10000), 
		dg.generateRandomString(8))
}

// GenerateShortID генерирует случайный короткий ID
func (dg *DataGenerator) GenerateShortID() string {
	return dg.generateRandomString(8)
}

// generateQueryParams генерирует случайные параметры запроса
func (dg *DataGenerator) generateQueryParams() string {
	paramCount := 1 + dg.rand.Intn(5)
	params := make([]string, paramCount)
	
	for i := 0; i < paramCount; i++ {
		key := dg.generateRandomString(3 + dg.rand.Intn(8))
		value := dg.generateRandomString(5 + dg.rand.Intn(15))
		params[i] = fmt.Sprintf("%s=%s", key, value)
	}
	
	return strings.Join(params, "&")
}

// generateRandomString генерирует случайную строку заданной длины
func (dg *DataGenerator) generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[dg.rand.Intn(len(charset))]
	}
	
	return string(result)
}

// GenerateBatchURLs генерирует пакет URL для batch тестирования
func (dg *DataGenerator) GenerateBatchURLs(count int, correlationIDPrefix string) []BatchURL {
	urls := make([]BatchURL, count)
	for i := 0; i < count; i++ {
		urls[i] = BatchURL{
			CorrelationID: fmt.Sprintf("%s-%d", correlationIDPrefix, i),
			OriginalURL:   dg.GenerateURL(),
		}
	}
	return urls
}

// BatchURL представляет URL для batch операций
type BatchURL struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}