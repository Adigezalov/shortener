package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	CookieName = "user_id"
	secretKey  = "your-secret-key-here" // В продакшене должен быть из конфига
)

// GenerateUserID создает новый уникальный ID пользователя
func GenerateUserID() string {
	return uuid.New().String()
}

// SignUserID создает подписанную куку с ID пользователя
func SignUserID(userID string) string {
	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write([]byte(userID))
	signature := hex.EncodeToString(h.Sum(nil))
	return fmt.Sprintf("%s.%s", userID, signature)
}

// VerifyUserID проверяет подпись куки и возвращает ID пользователя
func VerifyUserID(signedUserID string) (string, bool) {
	// Разделяем userID и подпись по последней точке
	lastDotIndex := strings.LastIndex(signedUserID, ".")
	if lastDotIndex == -1 {
		return "", false
	}

	userID := signedUserID[:lastDotIndex]
	signature := signedUserID[lastDotIndex+1:]

	if userID == "" || signature == "" {
		return "", false
	}

	// Проверяем подпись
	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write([]byte(userID))
	expectedSignature := hex.EncodeToString(h.Sum(nil))

	if !hmac.Equal([]byte(signature), []byte(expectedSignature)) {
		return "", false
	}

	return userID, true
}

// GetUserIDFromRequest извлекает и проверяет ID пользователя из куки или заголовка Authorization
func GetUserIDFromRequest(r *http.Request) (string, bool) {
	// Сначала пробуем получить из заголовка Authorization
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		// Ожидаем формат "Bearer <signed_user_id>"
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			signedUserID := authHeader[7:]
			userID, valid := VerifyUserID(signedUserID)
			if valid {
				return userID, true
			}
		}
		// Если нет префикса Bearer, пробуем как есть
		userID, valid := VerifyUserID(authHeader)
		if valid {
			return userID, true
		}
	}

	// Если заголовка нет, пробуем получить из куки
	cookie, err := r.Cookie(CookieName)
	if err != nil {
		return "", false
	}

	return VerifyUserID(cookie.Value)
}

// SetUserIDCookie устанавливает куку с ID пользователя
func SetUserIDCookie(w http.ResponseWriter, userID string) {
	signedUserID := SignUserID(userID)
	cookie := &http.Cookie{
		Name:     CookieName,
		Value:    signedUserID,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   int((24 * time.Hour).Seconds()), // 24 часа
	}
	http.SetCookie(w, cookie)
}

// SetAuthorizationHeader устанавливает заголовок Authorization с подписанным ID пользователя
func SetAuthorizationHeader(w http.ResponseWriter, userID string) {
	signedUserID := SignUserID(userID)
	w.Header().Set("Authorization", signedUserID)
}
