package grpcserver

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/Adigezalov/shortener/internal/auth"
	"github.com/Adigezalov/shortener/internal/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

// LoggingInterceptor перехватчик для логирования gRPC запросов.
func LoggingInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()

		// Получаем IP клиента
		clientIP := ""
		if p, ok := peer.FromContext(ctx); ok {
			clientIP = p.Addr.String()
		}

		logger.Logger.Info("gRPC запрос начат",
			zap.String("method", info.FullMethod),
			zap.String("client_ip", clientIP))

		// Вызываем handler
		resp, err := handler(ctx, req)

		// Логируем результат
		duration := time.Since(start)
		if err != nil {
			st, _ := status.FromError(err)
			logger.Logger.Error("gRPC запрос завершен с ошибкой",
				zap.String("method", info.FullMethod),
				zap.Duration("duration", duration),
				zap.String("error", st.Message()),
				zap.String("code", st.Code().String()))
		} else {
			logger.Logger.Info("gRPC запрос завершен успешно",
				zap.String("method", info.FullMethod),
				zap.Duration("duration", duration))
		}

		return resp, err
	}
}

// AuthInterceptor перехватчик для аутентификации пользователей.
func AuthInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Получаем метаданные из контекста
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			// Создаем метаданные, если их нет
			md = metadata.New(nil)
		}

		// Проверяем, есть ли токен авторизации
		tokens := md.Get("authorization")
		var userID string

		if len(tokens) > 0 {
			// Извлекаем токен (формат: Bearer <signed_user_id>)
			signedUserID := tokens[0]
			if len(signedUserID) > 7 && signedUserID[:7] == "Bearer " {
				signedUserID = signedUserID[7:]
			}

			// Верифицируем подписанный user ID
			var err error
			userID, err = auth.VerifyUserID(signedUserID)
			if err != nil {
				logger.Logger.Warn("gRPC: невалидный токен",
					zap.String("method", info.FullMethod),
					zap.Error(err))
				// Генерируем новый user ID для анонимного пользователя
				userID = auth.GenerateUserID()
				signedUserID = auth.SignUserID(userID)
				// Добавляем новый токен в исходящие метаданные
				header := metadata.Pairs("authorization", "Bearer "+signedUserID)
				grpc.SendHeader(ctx, header)
			}
		} else {
			// Если токена нет, генерируем новый для анонимного пользователя
			userID = auth.GenerateUserID()
			signedUserID := auth.SignUserID(userID)

			// Добавляем новый токен в исходящие метаданные
			header := metadata.Pairs("authorization", "Bearer "+signedUserID)
			grpc.SendHeader(ctx, header)
		}

		// Добавляем user ID в метаданные контекста
		md.Set("user-id", userID)
		newCtx := metadata.NewIncomingContext(ctx, md)

		// Вызываем handler с обновленным контекстом
		return handler(newCtx, req)
	}
}

// IPAuthInterceptor перехватчик для проверки доверенной подсети.
// Применяется только к методам, требующим проверки IP (например, GetStats).
func IPAuthInterceptor(trustedSubnet string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Проверяем IP только для метода GetStats
		if info.FullMethod != "/shortener.ShortenerService/GetStats" {
			return handler(ctx, req)
		}

		// Если доверенная подсеть не настроена, возвращаем ошибку
		if trustedSubnet == "" {
			return nil, status.Error(codes.PermissionDenied, "доступ запрещен")
		}

		// Получаем IP клиента
		var clientIP string
		if p, ok := peer.FromContext(ctx); ok {
			host, _, err := net.SplitHostPort(p.Addr.String())
			if err != nil {
				clientIP = p.Addr.String()
			} else {
				clientIP = host
			}
		}

		// Проверяем, находится ли IP в доверенной подсети
		if !isIPInSubnet(clientIP, trustedSubnet) {
			logger.Logger.Warn("gRPC: доступ запрещен - IP не в доверенной подсети",
				zap.String("client_ip", clientIP),
				zap.String("trusted_subnet", trustedSubnet))
			return nil, status.Error(codes.PermissionDenied, "доступ запрещен")
		}

		return handler(ctx, req)
	}
}

// isIPInSubnet проверяет, находится ли IP в указанной подсети CIDR.
func isIPInSubnet(ipStr, cidr string) bool {
	// Если CIDR пустой, возвращаем false
	if cidr == "" {
		return false
	}

	// Парсим CIDR
	_, subnet, err := net.ParseCIDR(cidr)
	if err != nil {
		logger.Logger.Error("gRPC: ошибка парсинга CIDR",
			zap.String("cidr", cidr),
			zap.Error(err))
		return false
	}

	// Парсим IP адрес
	ip := net.ParseIP(ipStr)
	if ip == nil {
		logger.Logger.Error("gRPC: ошибка парсинга IP",
			zap.String("ip", ipStr))
		return false
	}

	// Проверяем, находится ли IP в подсети
	return subnet.Contains(ip)
}

// RecoveryInterceptor перехватчик для восстановления после паники.
func RecoveryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				logger.Logger.Error("gRPC: паника в обработчике",
					zap.String("method", info.FullMethod),
					zap.Any("panic", r))
				err = status.Error(codes.Internal, fmt.Sprintf("внутренняя ошибка сервера: %v", r))
			}
		}()
		return handler(ctx, req)
	}
}
