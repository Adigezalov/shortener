package app

import (
	"github.com/Adigezalov/shortener/internal/config"
	"github.com/Adigezalov/shortener/internal/service"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"net/http"
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
	service := service.NewURLService(cfg.BaseURL, cfg.FileStoragePath)

	// Создаем chi роутер
	router := chi.NewRouter()

	router.Use(
		ungzipMiddleware,
		loggingMiddleware(logger),
		gzipMiddleware,
	)

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
