// Package config предоставляет конфигурацию приложения.
//
// Пакет поддерживает загрузку конфигурации из переменных окружения
// и аргументов командной строки с приоритетом аргументов командной строки.
package config

import (
	"encoding/json"
	"flag"
	"fmt"
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
	DefaultCertFile      = "cert.pem"              // Файл сертификата для HTTPS
	DefaultKeyFile       = "key.pem"               // Файл приватного ключа для HTTPS
	DefaultConfigFile    = ""                      // Файл конфигурации JSON (пустой = не используется)
)

// JSONConfig представляет структуру JSON файла конфигурации.
// Все поля опциональны и используются только если заданы в файле.
type JSONConfig struct {
	ServerAddress    *string `json:"server_address,omitempty"`    // Адрес HTTP сервера
	BaseURL          *string `json:"base_url,omitempty"`          // Базовый URL для коротких ссылок
	FileStoragePath  *string `json:"file_storage_path,omitempty"` // Путь к файлу хранения URL
	DatabaseDSN      *string `json:"database_dsn,omitempty"`      // DSN базы данных
	ProfilingEnabled *bool   `json:"profiling_enabled,omitempty"` // Включить профилирование
	ProfilingPort    *string `json:"profiling_port,omitempty"`    // Порт для pprof endpoints
	ProfilesDir      *string `json:"profiles_dir,omitempty"`      // Директория для профилей
	EnableHTTPS      *bool   `json:"enable_https,omitempty"`      // Включить HTTPS сервер
	CertFile         *string `json:"cert_file,omitempty"`         // Путь к файлу сертификата
	KeyFile          *string `json:"key_file,omitempty"`          // Путь к файлу приватного ключа
	TrustedSubnet    *string `json:"trusted_subnet,omitempty"`    // Доверенная подсеть CIDR
}

// Config содержит все конфигурационные параметры приложения.
//
// Конфигурация загружается в следующем порядке приоритета:
//  1. Аргументы командной строки (высший приоритет)
//  2. Переменные окружения
//  3. JSON файл конфигурации
//  4. Значения по умолчанию (низший приоритет)
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

	// EnableHTTPS включает или выключает HTTPS сервер.
	// При включении сервер запускается с помощью ListenAndServeTLS.
	// Переменная окружения: ENABLE_HTTPS (true/false)
	// Флаг: -s
	EnableHTTPS bool

	// CertFile определяет путь к файлу сертификата для HTTPS.
	// Переменная окружения: CERT_FILE
	// Флаг: -cert
	CertFile string

	// KeyFile определяет путь к файлу приватного ключа для HTTPS.
	// Переменная окружения: KEY_FILE
	// Флаг: -key
	KeyFile string

	// ConfigFile определяет путь к JSON файлу конфигурации.
	// Переменная окружения: CONFIG
	// Флаг: -c, -config
	ConfigFile string

	// TrustedSubnet определяет доверенную подсеть CIDR для доступа к внутренним эндпоинтам.
	// Переменная окружения: TRUSTED_SUBNET
	// Флаг: -t
	TrustedSubnet string
}

// loadJSONConfig загружает конфигурацию из JSON файла.
// Возвращает nil если файл не существует или пустой путь.
// Возвращает ошибку если файл существует но не может быть прочитан или распарсен.
func loadJSONConfig(configPath string) (*JSONConfig, error) {
	if configPath == "" {
		return nil, nil
	}

	// Проверяем существование файла
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, nil
	}

	// Читаем файл
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("не удалось прочитать файл конфигурации %s: %w", configPath, err)
	}

	// Парсим JSON
	var jsonConfig JSONConfig
	if err := json.Unmarshal(data, &jsonConfig); err != nil {
		return nil, fmt.Errorf("не удалось распарсить JSON файл конфигурации %s: %w", configPath, err)
	}

	return &jsonConfig, nil
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

	// Шаг 1: Устанавливаем значения по умолчанию
	cfg.ServerAddress = DefaultServerAddress
	cfg.BaseURL = DefaultBaseURL
	cfg.FileStoragePath = DefaultFileStorage
	cfg.DatabaseDSN = DefaultDatabaseDSN
	cfg.ProfilingEnabled = false
	cfg.ProfilingPort = DefaultProfilingPort
	cfg.ProfilesDir = DefaultProfilesDir
	cfg.EnableHTTPS = false
	cfg.CertFile = DefaultCertFile
	cfg.KeyFile = DefaultKeyFile
	cfg.ConfigFile = DefaultConfigFile
	cfg.TrustedSubnet = ""

	// Шаг 2: Применяем переменные окружения (включая путь к конфигурационному файлу)
	if envServerAddr := os.Getenv("SERVER_ADDRESS"); envServerAddr != "" {
		cfg.ServerAddress = envServerAddr
	}
	if envBaseURL := os.Getenv("BASE_URL"); envBaseURL != "" {
		cfg.BaseURL = envBaseURL
	}
	if envFileStoragePath := os.Getenv("FILE_STORAGE_PATH"); envFileStoragePath != "" {
		cfg.FileStoragePath = envFileStoragePath
	}
	if envDatabaseDSN := os.Getenv("DATABASE_DSN"); envDatabaseDSN != "" {
		cfg.DatabaseDSN = envDatabaseDSN
	}
	if envProfilingEnabled := os.Getenv("PROFILING_ENABLED"); envProfilingEnabled == "true" {
		cfg.ProfilingEnabled = true
	} else if envProfilingEnabled == "false" {
		cfg.ProfilingEnabled = false
	}
	if envProfilingPort := os.Getenv("PROFILING_PORT"); envProfilingPort != "" {
		cfg.ProfilingPort = envProfilingPort
	}
	if envProfilesDir := os.Getenv("PROFILES_DIR"); envProfilesDir != "" {
		cfg.ProfilesDir = envProfilesDir
	}
	if envEnableHTTPS := os.Getenv("ENABLE_HTTPS"); envEnableHTTPS == "true" {
		cfg.EnableHTTPS = true
	} else if envEnableHTTPS == "false" {
		cfg.EnableHTTPS = false
	}
	if envCertFile := os.Getenv("CERT_FILE"); envCertFile != "" {
		cfg.CertFile = envCertFile
	}
	if envKeyFile := os.Getenv("KEY_FILE"); envKeyFile != "" {
		cfg.KeyFile = envKeyFile
	}
	if envConfigFile := os.Getenv("CONFIG"); envConfigFile != "" {
		cfg.ConfigFile = envConfigFile
	}
	if envTrustedSubnet := os.Getenv("TRUSTED_SUBNET"); envTrustedSubnet != "" {
		cfg.TrustedSubnet = envTrustedSubnet
	}

	// Шаг 3: Регистрируем флаги командной строки
	flag.StringVar(&cfg.ServerAddress, "a", cfg.ServerAddress, "адрес запуска HTTP-сервера")
	flag.StringVar(&cfg.BaseURL, "b", cfg.BaseURL, "базовый адрес для сокращенных URL")
	flag.StringVar(&cfg.FileStoragePath, "f", cfg.FileStoragePath, "путь к файлу хранения URL")
	flag.StringVar(&cfg.DatabaseDSN, "d", cfg.DatabaseDSN, "строка подключения к PostgreSQL")
	flag.BoolVar(&cfg.ProfilingEnabled, "profiling", cfg.ProfilingEnabled, "включить профилирование")
	flag.StringVar(&cfg.ProfilingPort, "profiling-port", cfg.ProfilingPort, "порт для pprof endpoints")
	flag.StringVar(&cfg.ProfilesDir, "profiles-dir", cfg.ProfilesDir, "директория для сохранения профилей")
	flag.BoolVar(&cfg.EnableHTTPS, "s", cfg.EnableHTTPS, "включить HTTPS сервер")
	flag.StringVar(&cfg.CertFile, "cert", cfg.CertFile, "путь к файлу сертификата для HTTPS")
	flag.StringVar(&cfg.KeyFile, "key", cfg.KeyFile, "путь к файлу приватного ключа для HTTPS")
	flag.StringVar(&cfg.ConfigFile, "c", cfg.ConfigFile, "путь к JSON файлу конфигурации")
	flag.StringVar(&cfg.ConfigFile, "config", cfg.ConfigFile, "путь к JSON файлу конфигурации")
	flag.StringVar(&cfg.TrustedSubnet, "t", cfg.TrustedSubnet, "доверенная подсеть CIDR")

	// Шаг 4: Парсим флаги командной строки
	flag.Parse()

	// Шаг 5: Загружаем JSON конфигурацию и применяем её значения
	// (только если они не были переопределены флагами или переменными окружения)
	jsonConfig, err := loadJSONConfig(cfg.ConfigFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка загрузки конфигурации: %v\n", err)
		os.Exit(1)
	}

	if jsonConfig != nil {
		// Применяем значения из JSON только если они не были установлены через флаги или переменные окружения
		if jsonConfig.ServerAddress != nil && !isFlagSet("a") && os.Getenv("SERVER_ADDRESS") == "" {
			cfg.ServerAddress = *jsonConfig.ServerAddress
		}
		if jsonConfig.BaseURL != nil && !isFlagSet("b") && os.Getenv("BASE_URL") == "" {
			cfg.BaseURL = *jsonConfig.BaseURL
		}
		if jsonConfig.FileStoragePath != nil && !isFlagSet("f") && os.Getenv("FILE_STORAGE_PATH") == "" {
			cfg.FileStoragePath = *jsonConfig.FileStoragePath
		}
		if jsonConfig.DatabaseDSN != nil && !isFlagSet("d") && os.Getenv("DATABASE_DSN") == "" {
			cfg.DatabaseDSN = *jsonConfig.DatabaseDSN
		}
		if jsonConfig.ProfilingEnabled != nil && !isFlagSet("profiling") && os.Getenv("PROFILING_ENABLED") == "" {
			cfg.ProfilingEnabled = *jsonConfig.ProfilingEnabled
		}
		if jsonConfig.ProfilingPort != nil && !isFlagSet("profiling-port") && os.Getenv("PROFILING_PORT") == "" {
			cfg.ProfilingPort = *jsonConfig.ProfilingPort
		}
		if jsonConfig.ProfilesDir != nil && !isFlagSet("profiles-dir") && os.Getenv("PROFILES_DIR") == "" {
			cfg.ProfilesDir = *jsonConfig.ProfilesDir
		}
		if jsonConfig.EnableHTTPS != nil && !isFlagSet("s") && os.Getenv("ENABLE_HTTPS") == "" {
			cfg.EnableHTTPS = *jsonConfig.EnableHTTPS
		}
		if jsonConfig.CertFile != nil && !isFlagSet("cert") && os.Getenv("CERT_FILE") == "" {
			cfg.CertFile = *jsonConfig.CertFile
		}
		if jsonConfig.KeyFile != nil && !isFlagSet("key") && os.Getenv("KEY_FILE") == "" {
			cfg.KeyFile = *jsonConfig.KeyFile
		}
		if jsonConfig.TrustedSubnet != nil && !isFlagSet("t") && os.Getenv("TRUSTED_SUBNET") == "" {
			cfg.TrustedSubnet = *jsonConfig.TrustedSubnet
		}
	}

	// Валидируем и нормализуем конфигурацию
	cfg.normalize()

	return cfg
}

// isFlagSet проверяет, был ли установлен флаг командной строки
func isFlagSet(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

// normalize выполняет нормализацию и валидацию параметров конфигурации
func (c *Config) normalize() {
	// Убеждаемся, что BaseURL не заканчивается слешем
	c.BaseURL = strings.TrimSuffix(c.BaseURL, "/")

	// Если в BaseURL не указан протокол, добавляем соответствующий протокол
	if !strings.HasPrefix(c.BaseURL, "http://") && !strings.HasPrefix(c.BaseURL, "https://") {
		if c.EnableHTTPS {
			c.BaseURL = "https://" + c.BaseURL
		} else {
			c.BaseURL = "http://" + c.BaseURL
		}
	}

	// Если включен HTTPS, но BaseURL использует HTTP, обновляем протокол
	if c.EnableHTTPS && strings.HasPrefix(c.BaseURL, "http://") {
		c.BaseURL = strings.Replace(c.BaseURL, "http://", "https://", 1)
	}
}
