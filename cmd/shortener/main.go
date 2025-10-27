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
	}

	// Инициализируем хранилище URL с помощью фабрики
	store, err := storage.Factory(cfg.DatabaseDSN, cfg.FileStoragePath)
	if err != nil {
		logger.Logger.Fatal("Ошибка инициализации хранилища", zap.Error(err))
	}

	// Инициализируем подключение к базе данных для хендлера /ping
	var dbInterface handlers.Pinger
	if cfg.DatabaseDSN != "" {
		db, err := database.New(cfg.DatabaseDSN)
		if err != nil {
			logger.Logger.Fatal("Ошибка подключения к базе данных", zap.Error(err))
		}
		dbInterface = db
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

	// Маршрут для внутренней статистики с проверкой IP
	r.Get("/api/internal/stats", customMiddleware.IPAuthMiddleware(cfg.TrustedSubnet)(http.HandlerFunc(handler.GetStats)).ServeHTTP)

	// Настраиваем HTTP-сервер
	srv := &http.Server{
		Addr:    cfg.ServerAddress,
		Handler: r,
	}

	// Канал для получения сигналов OS
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)

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
			zap.String("trusted_subnet", cfg.TrustedSubnet),
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
	sig := <-stop

	logger.Logger.Info("Получен сигнал завершения работы",
		zap.String("signal", sig.String()))

	// Создаем контекст с таймаутом для корректного завершения
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	logger.Logger.Info("Начинаем корректное завершение работы сервера...")

	// Завершаем работу основного HTTP сервера
	// Это позволит завершить обработку всех активных запросов
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Logger.Error("Ошибка при завершении работы HTTP сервера", zap.Error(err))
	} else {
		logger.Logger.Info("HTTP сервер корректно завершил работу")
	}

	// Завершаем работу сервера профилирования, если он запущен
	if profilingServer != nil {
		profilingCtx, profilingCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer profilingCancel()

		if err := profilingServer.Stop(profilingCtx); err != nil {
			logger.Logger.Error("Ошибка при завершении работы сервера профилирования", zap.Error(err))
		} else {
			logger.Logger.Info("Сервер профилирования корректно завершил работу")
		}
	}

	// Закрываем хранилище для сохранения всех данных
	logger.Logger.Info("Сохраняем данные в хранилище...")
	if err := store.Close(); err != nil {
		logger.Logger.Error("Ошибка при закрытии хранилища", zap.Error(err))
	} else {
		logger.Logger.Info("Хранилище корректно закрыто, все данные сохранены")
	}

	// Закрываем подключение к базе данных, если оно есть
	if dbInterface != nil {
		if closer, ok := dbInterface.(interface{ Close() error }); ok {
			if err := closer.Close(); err != nil {
				logger.Logger.Error("Ошибка при закрытии подключения к базе данных", zap.Error(err))
			} else {
				logger.Logger.Info("Подключение к базе данных корректно закрыто")
			}
		}
	}

	logger.Logger.Info("Сервер корректно завершил работу")
}
