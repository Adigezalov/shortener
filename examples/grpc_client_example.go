package examples

import (
	"context"
	"fmt"
	"log"
	"time"

	pb "github.com/Adigezalov/shortener/pkg/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

// ExampleGRPCClient демонстрирует использование gRPC клиента для сервиса сокращения URL.
func ExampleGRPCClient() {
	// Подключаемся к gRPC серверу
	conn, err := grpc.Dial("localhost:3200", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Не удалось подключиться: %v", err)
	}
	defer conn.Close()

	// Создаем клиента
	client := pb.NewShortenerServiceClient(conn)

	// Создаем контекст с таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Пример 1: Создание короткого URL
	fmt.Println("=== Пример 1: Создание короткого URL ===")
	createResp, err := client.CreateShortURL(ctx, &pb.CreateShortURLRequest{
		Url: "https://example.com/very/long/url",
	})
	if err != nil {
		log.Printf("Ошибка создания URL: %v", err)
	} else {
		fmt.Printf("Короткий URL: %s\n", createResp.ShortUrl)
		fmt.Printf("Конфликт: %v\n\n", createResp.Conflict)
	}

	// Пример 2: Сокращение URL с получением токена
	fmt.Println("=== Пример 2: Сокращение URL с получением токена ===")
	var header metadata.MD
	shortenResp, err := client.ShortenURL(ctx, &pb.ShortenURLRequest{
		Url: "https://example.com/another/url",
	}, grpc.Header(&header))
	if err != nil {
		log.Printf("Ошибка сокращения URL: %v", err)
	} else {
		fmt.Printf("Результат: %s\n", shortenResp.Result)
		fmt.Printf("Конфликт: %v\n", shortenResp.Conflict)

		// Получаем токен из заголовка
		if authTokens := header.Get("authorization"); len(authTokens) > 0 {
			fmt.Printf("Получен токен: %s\n\n", authTokens[0])
		}
	}

	// Пример 3: Пакетное сокращение URL
	fmt.Println("=== Пример 3: Пакетное сокращение URL ===")
	batchResp, err := client.ShortenBatch(ctx, &pb.ShortenBatchRequest{
		Items: []*pb.BatchShortenItem{
			{CorrelationId: "1", OriginalUrl: "https://example1.com"},
			{CorrelationId: "2", OriginalUrl: "https://example2.com"},
			{CorrelationId: "3", OriginalUrl: "https://example3.com"},
		},
	})
	if err != nil {
		log.Printf("Ошибка пакетного сокращения: %v", err)
	} else {
		fmt.Printf("Обработано URL: %d\n", len(batchResp.Items))
		for _, item := range batchResp.Items {
			fmt.Printf("  [%s] -> %s\n", item.CorrelationId, item.ShortUrl)
		}
		fmt.Println()
	}

	// Пример 4: Получение оригинального URL
	fmt.Println("=== Пример 4: Получение оригинального URL ===")
	if len(batchResp.Items) > 0 {
		// Извлекаем ID из короткого URL
		shortURL := batchResp.Items[0].ShortUrl
		id := shortURL[len(shortURL)-8:] // Предполагаем, что ID длиной 8 символов

		getResp, err := client.GetOriginalURL(ctx, &pb.GetOriginalURLRequest{
			Id: id,
		})
		if err != nil {
			log.Printf("Ошибка получения оригинального URL: %v", err)
		} else {
			fmt.Printf("Оригинальный URL: %s\n", getResp.OriginalUrl)
			fmt.Printf("Удален: %v\n\n", getResp.Deleted)
		}
	}

	// Пример 5: Получение URL пользователя
	fmt.Println("=== Пример 5: Получение URL пользователя ===")
	userURLsResp, err := client.GetUserURLs(ctx, &pb.GetUserURLsRequest{})
	if err != nil {
		log.Printf("Ошибка получения URL пользователя: %v", err)
	} else {
		fmt.Printf("Найдено URL пользователя: %d\n", len(userURLsResp.Urls))
		for i, url := range userURLsResp.Urls {
			if i < 5 { // Показываем только первые 5
				fmt.Printf("  %s -> %s\n", url.ShortUrl, url.OriginalUrl)
			}
		}
		if len(userURLsResp.Urls) > 5 {
			fmt.Printf("  ... и еще %d URL\n", len(userURLsResp.Urls)-5)
		}
		fmt.Println()
	}

	// Пример 6: Проверка БД
	fmt.Println("=== Пример 6: Проверка состояния БД ===")
	pingResp, err := client.Ping(ctx, &pb.PingRequest{})
	if err != nil {
		log.Printf("Ошибка проверки БД: %v", err)
	} else {
		fmt.Printf("База данных доступна: %v\n\n", pingResp.Ok)
	}

	// Пример 7: Удаление URL (демонстрация, но не выполняется)
	fmt.Println("=== Пример 7: Удаление URL (демонстрация) ===")
	fmt.Println("Для удаления URL используйте:")
	fmt.Println(`
	deleteResp, err := client.DeleteUserURLs(ctx, &pb.DeleteUserURLsRequest{
		ShortUrls: []string{"abc123", "def456"},
	})
	if err != nil {
		log.Printf("Ошибка удаления: %v", err)
	} else {
		fmt.Printf("Запрос принят: %v\n", deleteResp.Accepted)
	}
	`)
	fmt.Println()
}

// ExampleGRPCClientWithAuth демонстрирует использование gRPC клиента с авторизацией.
func ExampleGRPCClientWithAuth() {
	conn, err := grpc.Dial("localhost:3200", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Не удалось подключиться: %v", err)
	}
	defer conn.Close()

	client := pb.NewShortenerServiceClient(conn)

	// Первый запрос - получаем токен
	ctx1 := context.Background()
	var header metadata.MD

	fmt.Println("=== Первый запрос - получение токена ===")
	resp1, err := client.ShortenURL(ctx1, &pb.ShortenURLRequest{
		Url: "https://example.com",
	}, grpc.Header(&header))
	if err != nil {
		log.Fatalf("Ошибка: %v", err)
	}

	fmt.Printf("Короткий URL: %s\n", resp1.Result)

	// Извлекаем токен из заголовка
	var token string
	if authTokens := header.Get("authorization"); len(authTokens) > 0 {
		token = authTokens[0]
		fmt.Printf("Получен токен: %s\n\n", token)
	}

	// Второй запрос - используем полученный токен
	if token != "" {
		fmt.Println("=== Второй запрос - использование токена ===")
		md := metadata.New(map[string]string{
			"authorization": token,
		})
		ctx2 := metadata.NewOutgoingContext(context.Background(), md)

		resp2, err := client.GetUserURLs(ctx2, &pb.GetUserURLsRequest{})
		if err != nil {
			log.Printf("Ошибка: %v", err)
		} else {
			fmt.Printf("Найдено URL пользователя: %d\n", len(resp2.Urls))
			for _, url := range resp2.Urls {
				fmt.Printf("  %s -> %s\n", url.ShortUrl, url.OriginalUrl)
			}
		}
	}
}

// ExampleGRPCClientWithTLS демонстрирует подключение к gRPC серверу с TLS.
func ExampleGRPCClientWithTLS() {
	// Для реального использования замените на путь к вашему сертификату
	// creds, err := credentials.NewClientTLSFromFile("grpc_cert.pem", "")
	// if err != nil {
	// 	log.Fatalf("Ошибка загрузки сертификата: %v", err)
	// }
	//
	// conn, err := grpc.Dial("localhost:3200", grpc.WithTransportCredentials(creds))
	// if err != nil {
	// 	log.Fatalf("Не удалось подключиться: %v", err)
	// }
	// defer conn.Close()

	fmt.Println("=== Пример подключения с TLS ===")
	fmt.Println("Раскомментируйте код в функции ExampleGRPCClientWithTLS")
	fmt.Println("и замените путь к сертификату на ваш")
}
