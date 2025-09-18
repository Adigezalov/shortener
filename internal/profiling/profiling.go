package profiling

import (
	"context"
	"net/http"
	_ "net/http/pprof"

	"github.com/Adigezalov/shortener/internal/config"
	"github.com/Adigezalov/shortener/internal/logger"
	"go.uber.org/zap"
)

// Server структура для управления профилированием
type Server struct {
	server *http.Server
	config *config.Config
}

// NewServer создает новый сервер профилирования
func NewServer(cfg *config.Config) *Server {
	if !cfg.ProfilingEnabled {
		return nil
	}

	mux := http.NewServeMux()

	// pprof endpoints автоматически регистрируются при импорте net/http/pprof
	// Добавляем их явно для ясности
	mux.HandleFunc("/debug/pprof/", http.DefaultServeMux.ServeHTTP)

	server := &http.Server{
		Addr:    cfg.ProfilingPort,
		Handler: mux,
	}

	return &Server{
		server: server,
		config: cfg,
	}
}

// Start запускает сервер профилирования
func (s *Server) Start() error {
	if s == nil || s.server == nil {
		logger.Logger.Info("Profiling disabled")
		return nil
	}

	logger.Logger.Info("Starting profiling server",
		zap.String("address", s.server.Addr))

	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Logger.Error("Profiling server error", zap.Error(err))
		}
	}()

	return nil
}

// Stop останавливает сервер профилирования
func (s *Server) Stop(ctx context.Context) error {
	if s == nil || s.server == nil {
		return nil
	}

	logger.Logger.Info("Stopping profiling server")
	return s.server.Shutdown(ctx)
}

// IsEnabled возвращает true если профилирование включено
func (s *Server) IsEnabled() bool {
	return s != nil && s.server != nil
}

// GetAddress возвращает адрес сервера профилирования
func (s *Server) GetAddress() string {
	if s == nil || s.server == nil {
		return ""
	}
	return s.server.Addr
}
