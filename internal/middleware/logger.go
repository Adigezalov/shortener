package middleware

import (
	"context"
	"crypto/rand"
	"fmt"
	"github.com/Adigezalov/shortener/internal/logger"
	"go.uber.org/zap"
	"net/http"
	"runtime/debug"
	"time"
)

// contextKey пользовательский тип для ключей контекста
type contextKey string

// Определяем константы для ключей контекста
const requestIDKey contextKey = "requestID"

// responseWriter является оберткой для http.ResponseWriter для отслеживания статуса и размера ответа
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int
}

// WriteHeader перехватывает оригинальный метод для сохранения кода статуса
func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

// Write перехватывает оригинальный метод для подсчета размера ответа
func (rw *responseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	return size, err
}

// RequestLogger middleware для логирования запросов и ответов
func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Время начала обработки запроса
		start := time.Now()

		// Создаем обертку для ResponseWriter
		rw := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK, // По умолчанию 200 OK
		}

		// Передаем запрос следующему обработчику
		next.ServeHTTP(rw, r)

		// Вычисляем длительность обработки запроса
		duration := time.Since(start)

		// Логируем информацию о запросе и ответе
		logger.Logger.Info("HTTP request",
			zap.String("method", r.Method),
			zap.String("uri", r.RequestURI),
			zap.Int("status", rw.statusCode),
			zap.Int("response_size", rw.size),
			zap.Duration("duration", duration),
		)
	})
}

// LoggingRecoverer middleware для восстановления после паники с логированием
func LoggingRecoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// Получаем стек вызовов
				stack := debug.Stack()

				// Логируем панику
				logger.Logger.Error("Паника в обработчике HTTP",
					zap.Any("error", err),
					zap.ByteString("stack", stack),
					zap.String("method", r.Method),
					zap.String("uri", r.RequestURI),
				)

				// Отвечаем клиенту с ошибкой
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()

		// Выполняем следующий обработчик
		next.ServeHTTP(w, r)
	})
}

// WithRequestID добавляет уникальный идентификатор запроса в контекст и логи
func WithRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Генерируем уникальный ID запроса
		requestID := generateRequestID()

		// Добавляем ID запроса в заголовок ответа
		w.Header().Set("X-Request-ID", requestID)

		// Создаем новый контекст с ID запроса
		ctx := context.WithValue(r.Context(), requestIDKey, requestID)

		// Логируем с ID запроса
		logger.Logger.Info("Request started",
			zap.String("request_id", requestID),
			zap.String("method", r.Method),
			zap.String("uri", r.RequestURI),
		)

		// Вызываем следующий обработчик с обновленным контекстом
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetRequestID возвращает ID запроса из контекста
func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey).(string); ok {
		return id
	}
	return "unknown"
}

// generateRequestID создает уникальный идентификатор запроса
func generateRequestID() string {
	b := make([]byte, 8)
	_, err := rand.Read(b)
	if err != nil {
		return "unknown"
	}
	return fmt.Sprintf("%x", b)
}
