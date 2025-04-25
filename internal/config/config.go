package config

import (
	"flag"
	"strings"
)

// Config хранит все параметры конфигурации приложения
type Config struct {
	ServerAddress string // Адрес запуска HTTP-сервера (флаг -a)
	BaseURL       string // Базовый адрес для сокращенных URL (флаг -b)
}

// ParseFlags обрабатывает аргументы командной строки и возвращает Config
func ParseFlags() Config {
	var cfg Config

	flag.StringVar(&cfg.ServerAddress, "a", "localhost:8080", "HTTP server address (e.g., 'localhost:8080')")
	flag.StringVar(&cfg.BaseURL, "b", "http://localhost:8080", "Base URL for shortened links (e.g., 'http://localhost:8080')")

	flag.Parse()

	// Убедимся, что BaseURL не заканчивается на "/"
	cfg.BaseURL = strings.TrimRight(cfg.BaseURL, "/")

	return cfg
}
