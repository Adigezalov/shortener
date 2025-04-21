package service

import (
	"github.com/Adigezalov/shortener/pkg/utils"
	"net/http"
	"testing"
)

func TestURLService_SetBaseURL(t *testing.T) {
	// Подготовка тестового HTTP запроса
	req := &http.Request{
		Host: "example.com",
		Header: http.Header{
			"X-Forwarded-Proto": []string{"https"},
		},
	}

	// Получаем ожидаемый базовый URL через utils.GetBaseURL
	expectedBaseURL := utils.GetBaseURL(req)

	// Создаём экземпляр URLService
	service := &URLService{
		storage: nil, // storage не используется в этом тесте
	}

	// Вызываем метод SetBaseURL
	service.SetBaseURL(req)

	// Проверяем, что baseURL установлен корректно
	if service.baseURL != expectedBaseURL {
		t.Errorf("expected baseURL %q, got %q", expectedBaseURL, service.baseURL)
	}
}
