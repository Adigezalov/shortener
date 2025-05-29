package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger глобальный экземпляр логгера
var Logger *zap.Logger

// Initialize инициализирует логгер
func Initialize() error {
	// Настраиваем конфигурацию логгера
	config := zap.NewProductionConfig()
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.Level = zap.NewAtomicLevelAt(zap.InfoLevel) // Устанавливаем уровень Info

	// Создаем логгер
	var err error
	Logger, err = config.Build()
	if err != nil {
		return err
	}

	// Заменяем глобальный логгер zap
	zap.ReplaceGlobals(Logger)

	return nil
}

// Sync выполняет синхронизацию логгера перед завершением программы
func Sync() {
	if Logger != nil {
		// Игнорируем ошибки синхронизации, так как это часто
		// может происходить при завершении программы
		_ = Logger.Sync()
	}
}
