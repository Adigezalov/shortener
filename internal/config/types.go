package config

type Config struct {
	ServerAddress   string // Адрес запуска HTTP-сервера (флаг -a)
	BaseURL         string // Базовый адрес для сокращенных URL (флаг -b)
	FileStoragePath string // путь до файла, куда сохраняются данные в формате JSON (флаг -f)
}
