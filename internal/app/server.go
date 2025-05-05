package app

import (
	"github.com/Adigezalov/shortener/internal/config"
	"github.com/Adigezalov/shortener/internal/service"
	"github.com/Adigezalov/shortener/internal/storage"
	"github.com/go-chi/chi/v5"
	"net/http"
)

type Server struct {
	httpServer *http.Server
	service    URLService
	router     *chi.Mux
	config     config.Config
}

func NewServer(cfg config.Config) *Server {
	// Инициализация зависимостей
	storage := storage.NewMemoryStorage()
	service := service.NewURLService(storage, cfg.BaseURL)

	// Создаем chi роутер
	router := chi.NewRouter()

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
	}
}

func (s *Server) ListenAndServe() error {
	return s.httpServer.ListenAndServe()
}
