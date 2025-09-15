package examples

import (
	"fmt"
	"os"

	"github.com/Adigezalov/shortener/internal/config"
)

// ExampleNewConfig демонстрирует создание конфигурации с значениями по умолчанию.
func Example_configDefault() {
	// Очищаем переменные окружения для чистого теста
	os.Unsetenv("SERVER_ADDRESS")
	os.Unsetenv("BASE_URL")
	os.Unsetenv("DATABASE_DSN")

	cfg := config.NewConfig()

	fmt.Printf("Server Address: %s\n", cfg.ServerAddress)
	fmt.Printf("Base URL: %s\n", cfg.BaseURL)
	fmt.Printf("File Storage: %s\n", cfg.FileStoragePath)
	fmt.Printf("Database DSN: %s\n", cfg.DatabaseDSN)
	fmt.Printf("Profiling Enabled: %t\n", cfg.ProfilingEnabled)

	// Output:
	// Server Address: :8080
	// Base URL: http://localhost:8080
	// File Storage: storage.json
	// Database DSN:
	// Profiling Enabled: false
}

// ExampleNewConfigEnvironmentVariables демонстрирует загрузку конфигурации из переменных окружения.
func Example_configEnvironmentVariables() {
	// Устанавливаем переменные окружения
	os.Setenv("SERVER_ADDRESS", ":9090")
	os.Setenv("BASE_URL", "https://short.ly")
	os.Setenv("DATABASE_DSN", "postgres://user:pass@localhost/shortener")
	os.Setenv("PROFILING_ENABLED", "true")
	os.Setenv("PROFILING_PORT", ":6061")

	cfg := config.NewConfig()

	fmt.Printf("Server Address: %s\n", cfg.ServerAddress)
	fmt.Printf("Base URL: %s\n", cfg.BaseURL)
	fmt.Printf("Database DSN: %s\n", cfg.DatabaseDSN)
	fmt.Printf("Profiling Enabled: %t\n", cfg.ProfilingEnabled)
	fmt.Printf("Profiling Port: %s\n", cfg.ProfilingPort)

	// Очищаем переменные окружения
	os.Unsetenv("SERVER_ADDRESS")
	os.Unsetenv("BASE_URL")
	os.Unsetenv("DATABASE_DSN")
	os.Unsetenv("PROFILING_ENABLED")
	os.Unsetenv("PROFILING_PORT")

	// Output:
	// Server Address: :9090
	// Base URL: https://short.ly
	// Database DSN: postgres://user:pass@localhost/shortener
	// Profiling Enabled: true
	// Profiling Port: :6061
}

// ExampleConfigURLNormalization демонстрирует нормализацию базового URL.
func Example_configURLNormalization() {
	// Тест 1: URL с завершающим слешем
	os.Setenv("BASE_URL", "https://example.com/")
	cfg1 := config.NewConfig()
	fmt.Printf("URL с слешем: %s -> %s\n", "https://example.com/", cfg1.BaseURL)

	// Тест 2: URL без протокола
	os.Setenv("BASE_URL", "example.com")
	cfg2 := config.NewConfig()
	fmt.Printf("URL без протокола: %s -> %s\n", "example.com", cfg2.BaseURL)

	// Тест 3: URL с портом
	os.Setenv("BASE_URL", "localhost:8080")
	cfg3 := config.NewConfig()
	fmt.Printf("URL с портом: %s -> %s\n", "localhost:8080", cfg3.BaseURL)

	// Очищаем переменные окружения
	os.Unsetenv("BASE_URL")

	// Output:
	// URL с слешем: https://example.com/ -> https://example.com
	// URL без протокола: example.com -> http://example.com
	// URL с портом: localhost:8080 -> http://localhost:8080
}
