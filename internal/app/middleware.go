package app

import (
	"bytes"
	"compress/gzip"
	"go.uber.org/zap"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// gzipWriter wraps http.ResponseWriter to provide gzip compression
type gzipWriter struct {
	http.ResponseWriter
	writer *gzip.Writer
}

// Write writes compressed data to the underlying writer
func (g *gzipWriter) Write(data []byte) (int, error) {
	return g.writer.Write(data)
}

// Close closes the gzip writer
func (g *gzipWriter) Close() error {
	return g.writer.Close()
}

var gzipPool = sync.Pool{
	New: func() interface{} {
		return gzip.NewWriter(nil)
	},
}

// gzipMiddleware compresses responses for clients that support gzip
func gzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if client accepts gzip encoding
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		// Check if content type should be compressed
		contentType := r.Header.Get("Content-Type")
		if !strings.Contains(contentType, "application/json") &&
			!strings.Contains(contentType, "text/html") &&
			!strings.Contains(contentType, "text/plain") {
			next.ServeHTTP(w, r)
			return
		}

		// Get gzip writer from pool
		gz := gzipPool.Get().(*gzip.Writer)
		defer gzipPool.Put(gz)
		gz.Reset(w)

		// Create gzip response writer
		gzWriter := &gzipWriter{
			ResponseWriter: w,
			writer:         gz,
		}
		defer gzWriter.Close()

		// Set headers
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Add("Vary", "Accept-Encoding")

		next.ServeHTTP(gzWriter, r)
	})
}

// ungzipMiddleware decompresses gzipped requests
func ungzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if request is gzipped
		if !strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		// Create gzip reader
		gz, err := gzip.NewReader(r.Body)
		if err != nil {
			http.Error(w, "Failed to create gzip reader", http.StatusBadRequest)
			return
		}
		defer gz.Close()

		// Read decompressed data
		var buf bytes.Buffer
		if _, err := io.Copy(&buf, gz); err != nil {
			http.Error(w, "Failed to decompress data", http.StatusBadRequest)
			return
		}

		// Replace body with decompressed data
		r.Body = io.NopCloser(&buf)
		r.ContentLength = int64(buf.Len())
		r.Header.Del("Content-Encoding")

		next.ServeHTTP(w, r)
	})
}

func loggingMiddleware(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Создаем обертку для ResponseWriter, чтобы получить статус и размер ответа
			wrapped := wrapResponseWriter(w)

			// Продолжаем выполнение цепочки обработчиков
			next.ServeHTTP(wrapped, r)

			// Логируем информацию о запросе и ответе
			logger.Info("request completed",
				zap.String("uri", r.RequestURI),
				zap.String("method", r.Method),
				zap.Int("status", wrapped.status),
				zap.Int("size", wrapped.size),
				zap.Duration("duration", time.Since(start)),
			)
		})
	}
}
