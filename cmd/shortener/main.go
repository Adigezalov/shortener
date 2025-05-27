package main

import (
	"context"
	"github.com/Adigezalov/shortener/internal/config"
	"github.com/Adigezalov/shortener/internal/handlers"
	"github.com/Adigezalov/shortener/internal/logger"
	customMiddleware "github.com/Adigezalov/shortener/internal/middleware"
	"github.com/Adigezalov/shortener/internal/shortener"
	"github.com/Adigezalov/shortener/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// Инициализируем логгер
	if err := logger.Initialize(); err != nil {
		log.Fatalf("Ошибка инициализации логгера: %v", err)
	}
	defer logger.Sync()

	// Загружаем конфигурацию
	cfg := config.NewConfig()

	// Инициализируем хранилище URL
	store := storage.New(cfg.FileStoragePath)
	defer store.Close() // Закрываем хранилище при завершении

	// Инициализируем сервис сокращения URL с базовым URL из конфигурации
	shortenerService := shortener.New(cfg.BaseURL)

	// Инициализируем обработчик HTTP запросов
	handler := handlers.New(store, shortenerService)

	// Создаем новый роутер chi
	r := chi.NewRouter()

	// Добавляем middleware
	r.Use(middleware.CleanPath)              // Очистка пути URL
	r.Use(customMiddleware.LoggingRecoverer) // Восстановление после паники с логированием
	r.Use(customMiddleware.WithRequestID)    // Добавление ID запроса
	r.Use(customMiddleware.RequestLogger)    // Логирование запросов и ответов
	r.Use(customMiddleware.GzipMiddleware)   // Обработка gzip сжатия

	// Определяем маршруты
	r.Post("/", handler.CreateShortURL)        // Эндпоинт для создания сокращенного URL (text/plain)
	r.Get("/{id}", handler.RedirectToURL)      // Эндпоинт для перенаправления по идентификатору
	r.Post("/api/shorten", handler.ShortenURL) // Эндпоинт для создания сокращенного URL (JSON)

	// Настраиваем корректное завершение сервера
	srv := &http.Server{
		Addr:    cfg.ServerAddress,
		Handler: r,
	}

	// Канал для получения сигналов OS
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Запускаем сервер в отдельной горутине
	go func() {
		logger.Logger.Info("Сервер запущен",
			zap.String("address", cfg.ServerAddress),
			zap.String("base_url", cfg.BaseURL),
			zap.String("storage_path", cfg.FileStoragePath),
		)

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Logger.Fatal("Ошибка запуска сервера", zap.Error(err))
		}
	}()

	// Ожидаем сигнал завершения
	<-stop

	logger.Logger.Info("Завершение работы сервера, сохранение данных...")

	// Создаем контекст с таймаутом для корректного завершения
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Завершаем работу сервера
	if err := srv.Shutdown(ctx); err != nil {
		logger.Logger.Error("Ошибка при завершении работы сервера", zap.Error(err))
	}
}
