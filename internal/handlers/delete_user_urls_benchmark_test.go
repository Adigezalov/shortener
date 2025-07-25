package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Adigezalov/shortener/internal/logger"
	"github.com/Adigezalov/shortener/internal/middleware"
	"github.com/Adigezalov/shortener/internal/storage"
	"go.uber.org/zap"
)

func BenchmarkHandler_DeleteUserURLs(b *testing.B) {
	// Инициализируем тестовый логгер
	testLogger, err := zap.NewDevelopment()
	if err != nil {
		b.Fatalf("Не удалось создать тестовый логгер: %v", err)
	}
	logger.Logger = testLogger
	defer logger.Logger.Sync()

	// Создаем хранилище в памяти
	store := storage.NewMemoryStorage("")
	defer store.Close()

	// Создаем хендлер
	handler := New(store, nil, nil)

	userID := "bench-user"

	// Подготавливаем данные для бенчмарка
	shortURLs := make([]string, 100)
	for i := 0; i < 100; i++ {
		shortID := fmt.Sprintf("bench%d", i)
		shortURLs[i] = shortID
		url := fmt.Sprintf("http://example%d.com", i)
		_, _, err := store.AddWithUser(shortID, url, userID)
		if err != nil {
			b.Fatalf("Ошибка добавления URL: %v", err)
		}
	}

	// Подготавливаем запрос на удаление
	deleteURLs := shortURLs[:50] // Удаляем половину URL
	body, _ := json.Marshal(deleteURLs)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodDelete, "/api/user/urls", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		// Добавляем userID в контекст
		ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
		req = req.WithContext(ctx)

		// Создаем ResponseRecorder
		w := httptest.NewRecorder()

		// Вызываем хендлер
		handler.DeleteUserURLs(w, req)

		if w.Code != http.StatusAccepted {
			b.Fatalf("Ожидался статус %d, получен %d", http.StatusAccepted, w.Code)
		}
	}
}
