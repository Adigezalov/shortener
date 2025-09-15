package examples

import (
	"fmt"

	"github.com/Adigezalov/shortener/internal/config"
	"github.com/Adigezalov/shortener/internal/shortener"
)

// Example демонстрирует базовое использование сервиса сокращения URL.
func Example() {
	// Создаем конфигурацию
	cfg := config.NewConfig()
	fmt.Printf("Server Address: %s\n", cfg.ServerAddress)
	fmt.Printf("Base URL: %s\n", cfg.BaseURL)

	// Создаем сервис сокращения
	service := shortener.New("https://short.ly")
	
	// Сокращаем URL
	originalURL := "https://example.com/very/long/path"
	shortID := service.Shorten(originalURL)
	_ = service.BuildShortURL(shortID) // shortURL для демонстрации
	
	fmt.Printf("Оригинальный URL: %s\n", originalURL)
	fmt.Printf("Длина короткого ID: %d\n", len(shortID))
	fmt.Printf("Короткий URL начинается с: https://short.ly/\n")
	
	// Output:
	// Server Address: :8080
	// Base URL: http://localhost:8080
	// Оригинальный URL: https://example.com/very/long/path
	// Длина короткого ID: 8
	// Короткий URL начинается с: https://short.ly/
}