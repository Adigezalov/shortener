package handlers

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/Adigezalov/shortener/internal/logger"
	"github.com/Adigezalov/shortener/internal/middleware"
	"go.uber.org/zap"
)

// DeleteUserURLs асинхронно удаляет URL пользователя
func (h *Handler) DeleteUserURLs(w http.ResponseWriter, r *http.Request) {
	// Получаем ID пользователя из контекста
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		logger.Logger.Error("Не удалось получить ID пользователя из контекста")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Декодируем список URL для удаления
	var shortURLs []string
	if err := json.NewDecoder(r.Body).Decode(&shortURLs); err != nil {
		logger.Logger.Error("Ошибка декодирования запроса на удаление URL",
			zap.String("user_id", userID),
			zap.Error(err))
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Проверяем, что список не пустой
	if len(shortURLs) == 0 {
		logger.Logger.Warn("Получен пустой список URL для удаления",
			zap.String("user_id", userID))
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Запускаем асинхронное удаление URL
	go h.asyncDeleteURLs(userID, shortURLs)

	// Возвращаем статус 202 Accepted
	w.WriteHeader(http.StatusAccepted)

	logger.Logger.Info("Принят запрос на удаление URL",
		zap.String("user_id", userID),
		zap.Int("count", len(shortURLs)))
}

// asyncDeleteURLs асинхронно удаляет URL пользователя с использованием паттерна fanIn
func (h *Handler) asyncDeleteURLs(userID string, shortURLs []string) {
	// Создаем каналы для обработки URL
	inputCh := make(chan string, len(shortURLs))

	// Заполняем входной канал
	for _, shortURL := range shortURLs {
		inputCh <- shortURL
	}
	close(inputCh)

	// Создаем несколько воркеров для обработки URL (паттерн fanOut)
	const numWorkers = 3
	var wg sync.WaitGroup

	// Канал для сбора результатов (паттерн fanIn)
	resultCh := make(chan string, len(shortURLs))

	// Запускаем воркеров
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for shortURL := range inputCh {
				// Обрабатываем каждый URL
				resultCh <- shortURL
			}
		}()
	}

	// Ждем завершения всех воркеров
	go func() {
		wg.Wait()
		close(resultCh)
	}()

	// Собираем все URL для пакетного удаления
	var urlsToDelete []string
	for shortURL := range resultCh {
		urlsToDelete = append(urlsToDelete, shortURL)
	}

	// Выполняем пакетное удаление в хранилище
	if err := h.storage.DeleteUserURLs(userID, urlsToDelete); err != nil {
		logger.Logger.Error("Ошибка удаления URL пользователя",
			zap.String("user_id", userID),
			zap.Strings("short_urls", urlsToDelete),
			zap.Error(err))
		return
	}

	logger.Logger.Info("URL пользователя успешно помечены как удаленные",
		zap.String("user_id", userID),
		zap.Int("count", len(urlsToDelete)))
}
