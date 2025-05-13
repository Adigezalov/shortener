package app

import (
	"github.com/Adigezalov/shortener/internal/config"
	"github.com/Adigezalov/shortener/internal/service"
	"github.com/Adigezalov/shortener/internal/storage"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type Server struct {
	httpServer *http.Server
	service    URLService
	router     *chi.Mux
	config     config.Config
	Logger     *zap.Logger
}

func NewServer(cfg config.Config) *Server {
	// Инициализация логгера
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}

	// Инициализация зависимостей
	storage := storage.NewMemoryStorage()
	service := service.NewURLService(storage, cfg.BaseURL)

	// Создаем chi роутер
	router := chi.NewRouter()

	// Добавляем middleware для логирования
	router.Use(loggingMiddleware(logger))

	// Инициализируем обработчики
	h := NewHandlers(service)

	// Настраиваем маршруты
	router.Mount("/", h.Routes())

	return &Server{
		httpServer: &http.Server{
			Addr:    cfg.ServerAddress,
			Handler: router,
		},
		service: service,
		router:  router,
		config:  cfg,
		Logger:  logger,
	}
}

func (s *Server) ListenAndServe() error {
	s.Logger.Info("Server starting",
		zap.String("address", s.config.ServerAddress),
		zap.String("baseURL", s.config.BaseURL),
	)
	return s.httpServer.ListenAndServe()
}

// loggingMiddleware создает middleware для логирования запросов и ответов
func loggingMiddleware(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Создаем обертку для ResponseWriter, чтобы получить статус и размер ответа
			wrapped := wrapResponseWriter(w)

			// Продолжаем выполнение цепочки обработчиков
			next.ServeHTTP(wrapped, r)

			// Логируем информацию о запросе и ответе
			logger.Info("request completed",
				zap.String("uri", r.RequestURI),
				zap.String("method", r.Method),
				zap.Int("status", wrapped.status),
				zap.Int("size", wrapped.size),
				zap.Duration("duration", time.Since(start)),
			)
		})
	}
}

// responseWriterWrapper оборачивает http.ResponseWriter для получения статуса и размера ответа
type responseWriterWrapper struct {
	http.ResponseWriter
	status int
	size   int
}

func wrapResponseWriter(w http.ResponseWriter) *responseWriterWrapper {
	return &responseWriterWrapper{ResponseWriter: w}
}

func (w *responseWriterWrapper) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *responseWriterWrapper) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	size, err := w.ResponseWriter.Write(b)
	w.size += size
	return size, err
}
