package app

import (
	"github.com/Adigezalov/shortener/internal/service"
	"github.com/Adigezalov/shortener/internal/storage"
	"github.com/go-chi/chi/v5"
	"net/http"
)

const Address = ":8080"

type Server struct {
	httpServer *http.Server
	service    *service.URLService
	router     *chi.Mux
}

func NewServer() *Server {
	// Инициализация зависимостей
	_storage := storage.NewMemoryStorage()
	_service := service.NewURLService(_storage)

	// Создаем chi роутер
	router := chi.NewRouter()

	// Инициализируем обработчики
	h := NewHandlers(_service)

	// Настраиваем маршруты
	router.Mount("/", h.Routes())

	return &Server{
		httpServer: &http.Server{
			Addr:    Address,
			Handler: router,
		},
		service: _service,
		router:  router,
	}
}

func (s *Server) ListenAndServe() error {
	return s.httpServer.ListenAndServe()
}
