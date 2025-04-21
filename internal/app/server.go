package app

import (
	"github.com/Adigezalov/shortener/internal/service"
	"github.com/Adigezalov/shortener/internal/storage"
	"net/http"
)

const Address = ":8080"

type Server struct {
	httpServer *http.Server
	service    *service.URLService
}

func NewServer() *Server {
	_storage := storage.NewMemoryStorage()
	_service := service.NewURLService(_storage)

	mux := http.NewServeMux()
	mux.HandleFunc("/", NewHandlers(_service).RootHandler)

	return &Server{
		httpServer: &http.Server{
			Addr:    Address,
			Handler: mux,
		},
		service: _service,
	}
}

func (s *Server) ListenAndServe() error {
	return s.httpServer.ListenAndServe()
}
