package config

import (
	"flag"
	"os"
	"strings"
)

// ParseFlags обрабатывает аргументы командной строки и возвращает Config
func ParseFlags() *Config {
	cfg := &Config{}

	// Устанавливаем значения по умолчанию
	defaultServerAddr := "localhost:8080"
	defaultBaseURL := "http://localhost:8080"
	defaultFileStoragePath := "tmp/storage"

	// Сначала проверяем переменные окружения
	if envRunAddr := os.Getenv("SERVER_ADDRESS"); envRunAddr != "" {
		defaultServerAddr = envRunAddr
	}

	if envBaseURL := os.Getenv("BASE_URL"); envBaseURL != "" {
		defaultBaseURL = strings.TrimRight(envBaseURL, "/")
	}

	if envFileStoragePath := os.Getenv(" FILE_STORAGE_PATH "); envFileStoragePath != "" {
		defaultFileStoragePath = envFileStoragePath
	}

	// Затем парсим флаги, которые могут переопределить значения
	flag.StringVar(&cfg.ServerAddress, "a", defaultServerAddr, "HTTP server address")
	flag.StringVar(&cfg.BaseURL, "b", defaultBaseURL, "Base URL for shortened links")
	flag.StringVar(&cfg.FileStoragePath, "f", defaultFileStoragePath, "File storage path")

	flag.Parse()

	// Убедимся, что BaseURL не заканчивается на "/"
	cfg.BaseURL = strings.TrimRight(cfg.BaseURL, "/")

	return cfg
}
