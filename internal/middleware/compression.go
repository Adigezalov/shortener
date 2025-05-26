package middleware

import (
	"compress/gzip"
	"github.com/Adigezalov/shortener/internal/logger"
	"go.uber.org/zap"
	"io"
	"net/http"
	"strings"
)

// Типы контента, для которых будем применять сжатие
var compressibleTypes = []string{
	"application/json",
	"text/html",
}

// gzipWriter - обертка для http.ResponseWriter, которая применяет gzip-сжатие
type gzipWriter struct {
	http.ResponseWriter
	gzipWriter *gzip.Writer
	statusCode int
	written    bool
	compress   bool
}

// WriteHeader перехватывает статус код и вызывает оригинальный метод
func (w *gzipWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode

	// Если это не редирект и не ошибка, проверим тип контента
	if statusCode < 300 || statusCode >= 400 {
		contentType := w.Header().Get("Content-Type")
		w.compress = shouldCompress(contentType)

		if !w.compress {
			// Если контент не должен сжиматься, удаляем заголовок gzip
			w.Header().Del("Content-Encoding")
		}
	} else {
		// Для редиректов и ошибок не сжимаем
		w.Header().Del("Content-Encoding")
		w.compress = false
	}

	w.ResponseWriter.WriteHeader(statusCode)
	w.written = true
}

// Write записывает данные, возможно с применением сжатия
func (w *gzipWriter) Write(b []byte) (int, error) {
	if !w.written {
		// Если WriteHeader еще не был вызван, вызовем его с кодом 200
		w.WriteHeader(http.StatusOK)
	}

	if w.compress {
		return w.gzipWriter.Write(b)
	}

	return w.ResponseWriter.Write(b)
}

// Close закрывает gzip writer, если он используется
func (w *gzipWriter) Close() error {
	if w.compress {
		return w.gzipWriter.Close()
	}
	return nil
}

// gzipReader структура для хранения тела запроса, распакованного из gzip
type gzipReader struct {
	r  io.ReadCloser
	gz *gzip.Reader
}

// Read читает распакованные данные
func (r *gzipReader) Read(p []byte) (n int, err error) {
	return r.gz.Read(p)
}

// Close закрывает оба reader'а
func (r *gzipReader) Close() error {
	if err := r.gz.Close(); err != nil {
		return err
	}
	return r.r.Close()
}

// shouldCompress проверяет, нужно ли сжимать ответ
func shouldCompress(contentType string) bool {
	for _, t := range compressibleTypes {
		if strings.Contains(contentType, t) {
			return true
		}
	}
	return false
}

// GzipMiddleware middleware для обработки gzip-сжатия
func GzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Обработка запроса, сжатого с помощью gzip
		if r.Header.Get("Content-Encoding") == "gzip" {
			// Создаем gzip reader
			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				logger.Logger.Error("Ошибка создания gzip reader", zap.Error(err))
				http.Error(w, "Ошибка декодирования gzip запроса", http.StatusBadRequest)
				return
			}

			// Заменяем тело запроса на распакованное
			r.Body = &gzipReader{r: r.Body, gz: gz}

			// Логируем, что запрос был распакован
			logger.Logger.Info("Запрос распакован из gzip")
		}

		// Проверяем, поддерживает ли клиент gzip-сжатие в ответах
		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			// Создаем объект для записи сжатых данных
			gz := gzip.NewWriter(w)

			// Создаем обертку для ResponseWriter
			gzipW := &gzipWriter{
				ResponseWriter: w,
				gzipWriter:     gz,
				statusCode:     http.StatusOK,
				compress:       false,
			}

			// Устанавливаем заголовок Content-Encoding
			w.Header().Set("Content-Encoding", "gzip")

			// Вызываем следующий обработчик с оберткой для ResponseWriter
			next.ServeHTTP(gzipW, r)

			// Закрываем gzip writer
			if err := gzipW.Close(); err != nil {
				logger.Logger.Error("Ошибка закрытия gzip writer", zap.Error(err))
			}
		} else {
			// Клиент не поддерживает gzip, отправляем несжатый ответ
			next.ServeHTTP(w, r)
		}
	})
}
