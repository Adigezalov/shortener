// Package grpcserver предоставляет gRPC сервер для сервиса сокращения URL.
package grpcserver

import (
	"context"
	"strings"

	"github.com/Adigezalov/shortener/internal/logger"
	"github.com/Adigezalov/shortener/internal/service"
	pb "github.com/Adigezalov/shortener/pkg/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Server реализует gRPC сервер для ShortenerService.
type Server struct {
	pb.UnimplementedShortenerServiceServer
	service *service.ShortenerService
}

// NewServer создает новый экземпляр gRPC сервера.
func NewServer(svc *service.ShortenerService) *Server {
	return &Server{
		service: svc,
	}
}

// getUserIDFromContext извлекает user ID из метаданных контекста.
func getUserIDFromContext(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "метаданные отсутствуют")
	}

	userIDs := md.Get("user-id")
	if len(userIDs) == 0 {
		return "", status.Error(codes.Unauthenticated, "user-id отсутствует в метаданных")
	}

	return userIDs[0], nil
}

// getClientIPFromContext извлекает IP адрес клиента из метаданных контекста.
func getClientIPFromContext(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}

	ips := md.Get("x-real-ip")
	if len(ips) > 0 {
		return ips[0]
	}

	return ""
}

// CreateShortURL создает короткий URL из текста.
func (s *Server) CreateShortURL(ctx context.Context, req *pb.CreateShortURLRequest) (*pb.CreateShortURLResponse, error) {
	logger.Logger.Info("gRPC: CreateShortURL вызван",
		zap.String("url", req.Url))

	// Получаем user ID из контекста
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		logger.Logger.Error("gRPC: ошибка получения user ID", zap.Error(err))
		return nil, err
	}

	// Вызываем бизнес-логику
	result := s.service.CreateShortURL(req.Url, userID)
	if result.Error != nil {
		if result.Error == service.ErrEmptyURL {
			return nil, status.Error(codes.InvalidArgument, "URL не может быть пустым")
		}
		logger.Logger.Error("gRPC: ошибка создания короткого URL", zap.Error(result.Error))
		return nil, status.Error(codes.Internal, "ошибка сохранения URL")
	}

	logger.Logger.Info("gRPC: короткий URL создан",
		zap.String("short_url", result.ShortURL),
		zap.Bool("exists", result.Exists))

	return &pb.CreateShortURLResponse{
		ShortUrl: result.ShortURL,
		Conflict: result.Exists,
	}, nil
}

// ShortenURL сокращает URL (JSON API аналог).
func (s *Server) ShortenURL(ctx context.Context, req *pb.ShortenURLRequest) (*pb.ShortenURLResponse, error) {
	logger.Logger.Info("gRPC: ShortenURL вызван",
		zap.String("url", req.Url))

	// Получаем user ID из контекста
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		logger.Logger.Error("gRPC: ошибка получения user ID", zap.Error(err))
		return nil, err
	}

	// Вызываем бизнес-логику
	result := s.service.CreateShortURL(req.Url, userID)
	if result.Error != nil {
		if result.Error == service.ErrEmptyURL {
			return nil, status.Error(codes.InvalidArgument, "URL не может быть пустым")
		}
		logger.Logger.Error("gRPC: ошибка сокращения URL", zap.Error(result.Error))
		return nil, status.Error(codes.Internal, "ошибка сохранения URL")
	}

	logger.Logger.Info("gRPC: URL сокращен",
		zap.String("result", result.ShortURL),
		zap.Bool("conflict", result.Exists))

	return &pb.ShortenURLResponse{
		Result:   result.ShortURL,
		Conflict: result.Exists,
	}, nil
}

// ShortenBatch выполняет пакетное сокращение URL.
func (s *Server) ShortenBatch(ctx context.Context, req *pb.ShortenBatchRequest) (*pb.ShortenBatchResponse, error) {
	logger.Logger.Info("gRPC: ShortenBatch вызван",
		zap.Int("items_count", len(req.Items)))

	if len(req.Items) == 0 {
		return nil, status.Error(codes.InvalidArgument, "список URL не может быть пустым")
	}

	// Получаем user ID из контекста
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		logger.Logger.Error("gRPC: ошибка получения user ID", zap.Error(err))
		return nil, err
	}

	// Преобразуем proto запрос в service запрос
	items := make([]service.BatchItem, 0, len(req.Items))
	for _, item := range req.Items {
		items = append(items, service.BatchItem{
			CorrelationID: item.CorrelationId,
			OriginalURL:   item.OriginalUrl,
		})
	}

	// Вызываем бизнес-логику
	results := s.service.CreateShortURLBatch(items, userID)

	// Преобразуем результаты в proto ответ
	pbResults := make([]*pb.BatchShortenResultItem, 0, len(results))
	for _, result := range results {
		pbResults = append(pbResults, &pb.BatchShortenResultItem{
			CorrelationId: result.CorrelationID,
			ShortUrl:      result.ShortURL,
		})
	}

	logger.Logger.Info("gRPC: пакет URL сокращен",
		zap.Int("results_count", len(pbResults)))

	return &pb.ShortenBatchResponse{
		Items: pbResults,
	}, nil
}

// GetOriginalURL получает оригинальный URL по короткому ID.
func (s *Server) GetOriginalURL(ctx context.Context, req *pb.GetOriginalURLRequest) (*pb.GetOriginalURLResponse, error) {
	logger.Logger.Info("gRPC: GetOriginalURL вызван",
		zap.String("id", req.Id))

	// Извлекаем ID из полного URL, если передан полный URL
	id := req.Id
	if strings.Contains(id, "/") {
		parts := strings.Split(id, "/")
		id = parts[len(parts)-1]
	}

	// Вызываем бизнес-логику
	result := s.service.GetOriginalURL(id)
	if result.Error != nil {
		logger.Logger.Error("gRPC: ошибка получения оригинального URL", zap.Error(result.Error))
		return nil, status.Error(codes.Internal, "ошибка получения URL")
	}

	if !result.Found {
		logger.Logger.Warn("gRPC: URL не найден", zap.String("id", id))
		return nil, status.Error(codes.NotFound, "URL не найден")
	}

	logger.Logger.Info("gRPC: оригинальный URL получен",
		zap.String("original_url", result.OriginalURL),
		zap.Bool("deleted", result.Deleted))

	return &pb.GetOriginalURLResponse{
		OriginalUrl: result.OriginalURL,
		Deleted:     result.Deleted,
	}, nil
}

// GetUserURLs получает все URL пользователя.
func (s *Server) GetUserURLs(ctx context.Context, req *pb.GetUserURLsRequest) (*pb.GetUserURLsResponse, error) {
	logger.Logger.Info("gRPC: GetUserURLs вызван")

	// Получаем user ID из контекста
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		logger.Logger.Error("gRPC: ошибка получения user ID", zap.Error(err))
		return nil, err
	}

	// Вызываем бизнес-логику
	result := s.service.GetUserURLs(userID)
	if result.Error != nil {
		logger.Logger.Error("gRPC: ошибка получения URL пользователя", zap.Error(result.Error))
		return nil, status.Error(codes.Internal, "ошибка получения URL пользователя")
	}

	// Преобразуем результат в proto ответ
	pbURLs := make([]*pb.UserURLItem, 0, len(result.URLs))
	for _, url := range result.URLs {
		pbURLs = append(pbURLs, &pb.UserURLItem{
			ShortUrl:    url.ShortURL,
			OriginalUrl: url.OriginalURL,
		})
	}

	logger.Logger.Info("gRPC: URL пользователя получены",
		zap.String("user_id", userID),
		zap.Int("count", len(pbURLs)))

	return &pb.GetUserURLsResponse{
		Urls: pbURLs,
	}, nil
}

// DeleteUserURLs удаляет URL пользователя.
func (s *Server) DeleteUserURLs(ctx context.Context, req *pb.DeleteUserURLsRequest) (*pb.DeleteUserURLsResponse, error) {
	logger.Logger.Info("gRPC: DeleteUserURLs вызван",
		zap.Int("urls_count", len(req.ShortUrls)))

	if len(req.ShortUrls) == 0 {
		return nil, status.Error(codes.InvalidArgument, "список URL не может быть пустым")
	}

	// Получаем user ID из контекста
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		logger.Logger.Error("gRPC: ошибка получения user ID", zap.Error(err))
		return nil, err
	}

	// Асинхронно удаляем URL
	go func() {
		if err := s.service.DeleteUserURLs(userID, req.ShortUrls); err != nil {
			logger.Logger.Error("gRPC: ошибка удаления URL",
				zap.String("user_id", userID),
				zap.Error(err))
		} else {
			logger.Logger.Info("gRPC: URL успешно удалены",
				zap.String("user_id", userID),
				zap.Int("count", len(req.ShortUrls)))
		}
	}()

	return &pb.DeleteUserURLsResponse{
		Accepted: true,
	}, nil
}

// Ping проверяет состояние базы данных.
func (s *Server) Ping(ctx context.Context, req *pb.PingRequest) (*pb.PingResponse, error) {
	logger.Logger.Info("gRPC: Ping вызван")

	err := s.service.PingDB()
	if err != nil {
		if err == service.ErrDBNotConfigured {
			return nil, status.Error(codes.Unimplemented, "база данных не настроена")
		}
		logger.Logger.Error("gRPC: ошибка проверки БД", zap.Error(err))
		return &pb.PingResponse{Ok: false}, nil
	}

	logger.Logger.Info("gRPC: БД доступна")
	return &pb.PingResponse{Ok: true}, nil
}

// GetStats получает статистику сервиса.
func (s *Server) GetStats(ctx context.Context, req *pb.GetStatsRequest) (*pb.GetStatsResponse, error) {
	logger.Logger.Info("gRPC: GetStats вызван")

	// Проверяем IP адрес клиента (если настроена доверенная подсеть)
	// Это будет реализовано в middleware

	// Вызываем бизнес-логику
	result := s.service.GetStats()
	if result.Error != nil {
		logger.Logger.Error("gRPC: ошибка получения статистики", zap.Error(result.Error))
		return nil, status.Error(codes.Internal, "ошибка получения статистики")
	}

	logger.Logger.Info("gRPC: статистика получена",
		zap.Int("urls", result.URLs),
		zap.Int("users", result.Users))

	return &pb.GetStatsResponse{
		Urls:  int32(result.URLs),
		Users: int32(result.Users),
	}, nil
}
