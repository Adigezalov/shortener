package middleware

import (
	"context"
	"net/http"

	"github.com/Adigezalov/shortener/internal/auth"
)

type authContextKey string

const UserIDKey authContextKey = "userID"

// AuthMiddleware проверяет аутентификацию пользователя и устанавливает куку если нужно
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, valid := auth.GetUserIDFromRequest(r)
		
		if !valid {
			// Генерируем новый ID пользователя и устанавливаем куку
			userID = auth.GenerateUserID()
			auth.SetUserIDCookie(w, userID)
		}

		// Добавляем userID в контекст запроса
		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireAuth middleware требует наличия валидной аутентификации
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, valid := auth.GetUserIDFromRequest(r)
		
		if !valid {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Добавляем userID в контекст запроса
		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserIDFromContext извлекает ID пользователя из контекста
func GetUserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(UserIDKey).(string)
	return userID, ok
}