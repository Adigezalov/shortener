package config

import (
	"flag"
	"os"
	"strings"
)

// Константы для значений по умолчанию
const (
	DefaultServerAddress = ":8080"
	DefaultBaseURL       = "http://localhost:8080"
	DefaultFileStorage   = "storage.json"
	DefaultDatabaseDSN   = ""
	DefaultProfilingPort = ":6060"
	DefaultProfilesDir   = "benchmarks/profiles"
)

// Config содержит конфигурационные параметры приложения
type Config struct {
	// ServerAddress адрес запуска HTTP-сервера
	ServerAddress string
	// BaseURL базовый адрес для сокращенных URL
	BaseURL string
	// FileStoragePath путь к файлу хранения
	FileStoragePath string
	// DatabaseDSN строка подключения к PostgreSQL
	DatabaseDSN string
	// ProfilingEnabled включает/выключает профилирование
	ProfilingEnabled bool
	// ProfilingPort порт для pprof endpoints
	ProfilingPort string
	// ProfilesDir директория для сохранения профилей
	ProfilesDir string
}

// NewConfig создает и инициализирует конфигурацию из аргументов командной строки и переменных окружения
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
