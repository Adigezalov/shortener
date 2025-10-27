package middleware

import (
	"net"
	"net/http"
	"strings"

	"github.com/Adigezalov/shortener/internal/logger"
	"go.uber.org/zap"
)

// IPAuthMiddleware проверяет, что IP-адрес клиента входит в доверенную подсеть
func IPAuthMiddleware(trustedSubnet string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Если trusted_subnet пустой, доступ запрещен
			if trustedSubnet == "" {
				logger.Logger.Warn("Доступ к защищенному эндпоинту запрещен: trusted_subnet не настроена")
				w.WriteHeader(http.StatusForbidden)
				return
			}

			// Получаем IP адрес из заголовка X-Real-IP
			ipStr := r.Header.Get("X-Real-IP")
			if ipStr == "" {
				logger.Logger.Warn("Доступ к защищенному эндпоинту запрещен: заголовок X-Real-IP отсутствует")
				w.WriteHeader(http.StatusForbidden)
				return
			}

			// Парсим IP адрес
			ip := net.ParseIP(ipStr)
			if ip == nil {
				logger.Logger.Warn("Доступ к защищенному эндпоинту запрещен: неверный IP адрес",
					zap.String("ip", ipStr))
				w.WriteHeader(http.StatusForbidden)
				return
			}

			// Проверяем, входит ли IP в доверенную подсеть
			_, ipNet, err := net.ParseCIDR(trustedSubnet)
			if err != nil {
				logger.Logger.Error("Ошибка парсинга CIDR подсети",
					zap.String("trusted_subnet", trustedSubnet),
					zap.Error(err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			if !ipNet.Contains(ip) {
				logger.Logger.Warn("Доступ к защищенному эндпоинту запрещен: IP не в доверенной подсети",
					zap.String("ip", ipStr),
					zap.String("trusted_subnet", trustedSubnet))
				w.WriteHeader(http.StatusForbidden)
				return
			}

			logger.Logger.Info("Доступ разрешен к защищенному эндпоинту",
				zap.String("ip", ipStr),
				zap.String("trusted_subnet", trustedSubnet))

			// Если IP входит в подсеть, продолжаем обработку запроса
			next.ServeHTTP(w, r)
		})
	}
}

// getRealIP извлекает реальный IP адрес из запроса
func getRealIP(r *http.Request) string {
	// Сначала проверяем X-Real-IP
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return strings.TrimSpace(ip)
	}

	// Затем проверяем X-Forwarded-For
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		// X-Forwarded-For может содержать несколько IP через запятую
		ips := strings.Split(ip, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// В конце используем RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}

	return ip
}
