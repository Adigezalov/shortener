package handlers

import (
	"github.com/Adigezalov/shortener/internal/service"
)

// HandlerWithService содержит обработчики HTTP запросов, использующие service слой.
type HandlerWithService struct {
	service *service.ShortenerService
}

// NewHandlerWithService создает новый экземпляр обработчика с service слоем.
func NewHandlerWithService(svc *service.ShortenerService) *HandlerWithService {
	return &HandlerWithService{
		service: svc,
	}
}
