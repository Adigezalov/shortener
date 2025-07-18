package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
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
	// Разделяем userID и подпись
	var userID, signature string
	n, err := fmt.Sscanf(signedUserID, "%s.%s", &userID, &signature)
	if err != nil || n != 2 {
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

// GetUserIDFromRequest извлекает и проверяет ID пользователя из куки
func GetUserIDFromRequest(r *http.Request) (string, bool) {
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
