package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Adigezalov/shortener/internal/config"
	"github.com/Adigezalov/shortener/internal/database"
	"github.com/Adigezalov/shortener/internal/handlers"
	"github.com/Adigezalov/shortener/internal/logger"
	customMiddleware "github.com/Adigezalov/shortener/internal/middleware"
	"github.com/Adigezalov/shortener/internal/profiling"
	"github.com/Adigezalov/shortener/internal/shortener"
	"github.com/Adigezalov/shortener/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

// Глобальные переменные для информации о сборке.
// Значения устанавливаются во время сборки с помощью флагов -ldflags.
var (
	buildVersion = "N/A" // Версия сборки
	buildDate    = "N/A" // Дата сборки
	buildCommit  = "N/A" // Коммит Git
)

func main() {
	// Выводим информацию о сборке
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)

	// Инициализируем логгер
	if err := logger.Initialize(); err != nil {
		log.Fatalf("Ошибка инициализации логгера: %v", err)
	}
	defer logger.Sync()

	// Загружаем конфигурацию
	cfg := config.NewConfig()

	// Инициализируем сервер профилирования
	profilingServer := profiling.NewServer(cfg)
	if profilingServer != nil {
		if err := profilingServer.Start(); err != nil {
			logger.Logger.Error("Ошибка запуска сервера профилирования", zap.Error(err))
		}
		defer func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			profilingServer.Stop(ctx)
		}()
	}

	// Инициализируем хранилище URL с помощью фабрики
	store, err := storage.Factory(cfg.DatabaseDSN, cfg.FileStoragePath)
	if err != nil {
		logger.Logger.Fatal("Ошибка инициализации хранилища", zap.Error(err))
	}
	defer store.Close()

	// Инициализируем подключение к базе данных для хендлера /ping
	var dbInterface handlers.Pinger
	if cfg.DatabaseDSN != "" {
		db, err := database.New(cfg.DatabaseDSN)
		if err != nil {
			logger.Logger.Fatal("Ошибка подключения к базе данных", zap.Error(err))
		}
		dbInterface = db
		defer db.Close()
	} else {
		// Если база данных не настроена, передаем nil
		dbInterface = nil
	}

	// Инициализируем сервис сокращения URL
	shortenerService := shortener.New(cfg.BaseURL)

	// Инициализируем обработчик HTTP запросов
	handler := handlers.New(store, shortenerService, dbInterface)

	// Создаем новый роутер chi
	r := chi.NewRouter()

	// Добавляем глобальные middleware
	r.Use(middleware.CleanPath)
	r.Use(customMiddleware.LoggingRecoverer)
	r.Use(customMiddleware.WithRequestID)
	r.Use(customMiddleware.RequestLogger)
	r.Use(customMiddleware.GzipMiddleware)
	r.Use(customMiddleware.AuthMiddleware) // Добавляем middleware аутентификации

	// Определяем маршруты
	r.Get("/ping", handler.PingDB)
	r.With(customMiddleware.TextPlainContentTypeMiddleware()).Post("/", handler.CreateShortURL)
	r.With(customMiddleware.JSONContentTypeMiddleware()).Post("/api/shorten", handler.ShortenURL)
	r.With(customMiddleware.JSONContentTypeMiddleware()).Post("/api/shorten/batch", handler.ShortenBatch)
	r.Get("/{id}", handler.RedirectToURL)

	// Маршруты, требующие аутентификации
	r.Route("/api/user", func(r chi.Router) {
		r.Use(customMiddleware.RequireAuth)
		r.Get("/urls", handler.GetUserURLs)
		r.Delete("/urls", handler.DeleteUserURLs)
	})

	// Настраиваем HTTP-сервер
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
			zap.String("database_dsn", cfg.DatabaseDSN),
			zap.Bool("profiling_enabled", cfg.ProfilingEnabled),
			zap.String("profiling_port", cfg.ProfilingPort),
			zap.Bool("https_enabled", cfg.EnableHTTPS),
			zap.String("cert_file", cfg.CertFile),
			zap.String("key_file", cfg.KeyFile),
		)

		var err error
		if cfg.EnableHTTPS {
			err = srv.ListenAndServeTLS(cfg.CertFile, cfg.KeyFile)
		} else {
			err = srv.ListenAndServe()
		}

		if err != nil && err != http.ErrServerClosed {
			logger.Logger.Fatal("Ошибка запуска сервера", zap.Error(err))
		}
	}()

	// Ожидаем сигнал завершения
	<-stop

	logger.Logger.Info("Завершение работы сервера...")

	// Создаем контекст с таймаутом для корректного завершения
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Завершаем работу сервера
	if err := srv.Shutdown(ctx); err != nil {
		logger.Logger.Error("Ошибка при завершении работы сервера", zap.Error(err))
	}
}
