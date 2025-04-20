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
	storage := storage.NewMemoryStorage()
	service := service.NewURLService(storage)

	mux := http.NewServeMux()
	mux.HandleFunc("/", NewHandlers(service).RootHandler)

	return &Server{
		httpServer: &http.Server{
			Addr:    Address,
			Handler: mux,
		},
		service: service,
	}
}

func (s *Server) ListenAndServe() error {
	return s.httpServer.ListenAndServe()
}
