package config

import (
	"flag"
	"os"
	"strings"
)

// Config содержит конфигурационные параметры приложения
type Config struct {
	// ServerAddress адрес запуска HTTP-сервера
	ServerAddress string
	// BaseURL базовый адрес для сокращенных URL
	BaseURL string
	// FileStoragePath путь к файлу хранения
	FileStoragePath string
}

// NewConfig создает и инициализирует конфигурацию из аргументов командной строки и переменных окружения
func NewConfig() *Config {
	cfg := &Config{}

	// Устанавливаем значения по умолчанию
	serverAddress := "localhost:8080"
	baseURL := "http://localhost:8080"
	fileStoragePath := "./storage.json"

	// Проверяем переменные окружения
	if envServerAddr := os.Getenv("SERVER_ADDRESS"); envServerAddr != "" {
		serverAddress = envServerAddr
	}
	if envBaseURL := os.Getenv("BASE_URL"); envBaseURL != "" {
		baseURL = envBaseURL
	}

	if envFileStoragePath := os.Getenv("FILE_STORAGE_PATH"); envFileStoragePath != "" {
		fileStoragePath = envFileStoragePath
	}

	// Регистрируем флаги командной строки
	flag.StringVar(&cfg.ServerAddress, "a", serverAddress, "адрес запуска HTTP-сервера")
	flag.StringVar(&cfg.BaseURL, "b", baseURL, "базовый адрес для сокращенных URL")
	flag.StringVar(&cfg.FileStoragePath, "f", fileStoragePath, "путь к файлу хранения URL")

	// Разбираем флаги
	flag.Parse()

	// Валидируем и нормализуем конфигурацию
	cfg.normalize()

	return cfg
}

// normalize выполняет нормализацию и валидацию параметров конфигурации
func (c *Config) normalize() {
	// Убеждаемся, что BaseURL не заканчивается слешем
	c.BaseURL = strings.TrimSuffix(c.BaseURL, "/")

	// Если в BaseURL не указан протокол, добавляем http://
	if !strings.HasPrefix(c.BaseURL, "http://") && !strings.HasPrefix(c.BaseURL, "https://") {
		c.BaseURL = "http://" + c.BaseURL
	}
}
