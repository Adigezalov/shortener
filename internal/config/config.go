// Package config предоставляет конфигурацию приложения.
//
// Пакет поддерживает загрузку конфигурации из переменных окружения
// и аргументов командной строки с приоритетом аргументов командной строки.
package config

import (
	"flag"
	"os"
	"strings"
)

// Константы для значений по умолчанию
const (
	DefaultServerAddress = ":8080"                 // Адрес HTTP сервера по умолчанию
	DefaultBaseURL       = "http://localhost:8080" // Базовый URL для коротких ссылок
	DefaultFileStorage   = "storage.json"          // Файл для хранения URL
	DefaultDatabaseDSN   = ""                      // DSN базы данных (пустой = не используется)
	DefaultProfilingPort = ":6060"                 // Порт для pprof endpoints
	DefaultProfilesDir   = "benchmarks/profiles"   // Директория для профилей производительности
)

// Config содержит все конфигурационные параметры приложения.
//
// Конфигурация загружается в следующем порядке приоритета:
//  1. Аргументы командной строки (высший приоритет)
//  2. Переменные окружения
//  3. Значения по умолчанию (низший приоритет)
//
// Пример использования:
//
//	cfg := config.NewConfig()
//	server := &http.Server{
//		Addr: cfg.ServerAddress,
//	}
type Config struct {
	// ServerAddress определяет адрес и порт для запуска HTTP-сервера.
	// Формат: ":8080" или "localhost:8080"
	// Переменная окружения: SERVER_ADDRESS
	// Флаг: -a
	ServerAddress string

	// BaseURL определяет базовый адрес для формирования коротких URL.
	// Должен включать протокол и домен: "https://example.com"
	// Переменная окружения: BASE_URL
	// Флаг: -b
	BaseURL string

	// FileStoragePath определяет путь к файлу для хранения URL.
	// Используется только если DatabaseDSN не задан.
	// Переменная окружения: FILE_STORAGE_PATH
	// Флаг: -f
	FileStoragePath string

	// DatabaseDSN содержит строку подключения к PostgreSQL.
	// Формат: "postgres://user:password@host:port/dbname?sslmode=disable"
	// Если пустой, используется файловое хранилище.
	// Переменная окружения: DATABASE_DSN
	// Флаг: -d
	DatabaseDSN string

	// ProfilingEnabled включает или выключает pprof профилирование.
	// При включении запускается дополнительный HTTP сервер на ProfilingPort.
	// Переменная окружения: PROFILING_ENABLED (true/false)
	// Флаг: -profiling
	ProfilingEnabled bool

	// ProfilingPort определяет порт для pprof endpoints.
	// Формат: ":6060"
	// Переменная окружения: PROFILING_PORT
	// Флаг: -profiling-port
	ProfilingPort string

	// ProfilesDir определяет директорию для сохранения профилей производительности.
	// Переменная окружения: PROFILES_DIR
	// Флаг: -profiles-dir
	ProfilesDir string
}

// NewConfig создает и инициализирует конфигурацию из переменных окружения и аргументов командной строки.
//
// Функция автоматически парсит флаги командной строки и применяет нормализацию
// к полученным значениям. Аргументы командной строки имеют приоритет над
// переменными окружения.
//
// Возвращает готовую к использованию конфигурацию.
//
// Пример использования:
//
//	cfg := config.NewConfig()
//	fmt.Printf("Server will start on %s\n", cfg.ServerAddress)
//	fmt.Printf("Base URL: %s\n", cfg.BaseURL)
func NewConfig() *Config {
	cfg := &Config{}

	// Устанавливаем значения по умолчанию
	serverAddress := DefaultServerAddress
	baseURL := DefaultBaseURL
	fileStoragePath := DefaultFileStorage
	databaseDSN := DefaultDatabaseDSN
	profilingEnabled := false
	profilingPort := DefaultProfilingPort
	profilesDir := DefaultProfilesDir

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
	if envDatabaseDSN := os.Getenv("DATABASE_DSN"); envDatabaseDSN != "" {
		databaseDSN = envDatabaseDSN
	}
	if envProfilingEnabled := os.Getenv("PROFILING_ENABLED"); envProfilingEnabled == "true" {
		profilingEnabled = true
	}
	if envProfilingPort := os.Getenv("PROFILING_PORT"); envProfilingPort != "" {
		profilingPort = envProfilingPort
	}
	if envProfilesDir := os.Getenv("PROFILES_DIR"); envProfilesDir != "" {
		profilesDir = envProfilesDir
	}

	// Регистрируем флаги командной строки
	flag.StringVar(&cfg.ServerAddress, "a", serverAddress, "адрес запуска HTTP-сервера")
	flag.StringVar(&cfg.BaseURL, "b", baseURL, "базовый адрес для сокращенных URL")
	flag.StringVar(&cfg.FileStoragePath, "f", fileStoragePath, "путь к файлу хранения URL")
	flag.StringVar(&cfg.DatabaseDSN, "d", databaseDSN, "строка подключения к PostgreSQL")
	flag.BoolVar(&cfg.ProfilingEnabled, "profiling", profilingEnabled, "включить профилирование")
	flag.StringVar(&cfg.ProfilingPort, "profiling-port", profilingPort, "порт для pprof endpoints")
	flag.StringVar(&cfg.ProfilesDir, "profiles-dir", profilesDir, "директория для сохранения профилей")

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
