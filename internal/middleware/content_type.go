package middleware

import (
	"net/http"
	"strings"
)

// ContentTypeMiddleware проверяет, что запрос имеет правильный Content-Type
func ContentTypeMiddleware(contentType string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Пропускаем проверку для GET запросов, так как они обычно не имеют тела
			if r.Method != http.MethodGet {
				// Проверяем Content-Type
				ct := r.Header.Get("Content-Type")
				if !strings.Contains(ct, contentType) {
					http.Error(w, "Invalid Content-Type", http.StatusBadRequest)
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// JSONContentTypeMiddleware проверяет, что запрос имеет Content-Type application/json
func JSONContentTypeMiddleware() func(http.Handler) http.Handler {
	return ContentTypeMiddleware("application/json")
}

// TextPlainContentTypeMiddleware проверяет, что запрос имеет Content-Type text/plain
func TextPlainContentTypeMiddleware() func(http.Handler) http.Handler {
	return ContentTypeMiddleware("text/plain")
}
