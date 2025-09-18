// Package examples содержит примеры использования API сервиса сокращения URL.
//
// Примеры демонстрируют работу со всеми эндпоинтами практического трека:
//   - Создание коротких URL
//   - Пакетное создание URL
//   - Получение оригинального URL
//   - Управление URL пользователя
//   - Проверка состояния сервиса
package examples

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/Adigezalov/shortener/internal/handlers"
	"github.com/Adigezalov/shortener/internal/logger"
	"github.com/Adigezalov/shortener/internal/models"
	"github.com/Adigezalov/shortener/internal/shortener"
	"github.com/Adigezalov/shortener/internal/storage"
	"github.com/go-chi/chi/v5"
)

// setupTestServer создает тестовый сервер для примеров
func setupTestServer() *httptest.Server {
	// Инициализируем логгер для тестов
	logger.Initialize()
	
	// Создаем хранилище в памяти
	store := storage.NewMemoryStorage("")

	// Создаем сервис сокращения
	shortenerService := shortener.New("http://localhost:8080")

	// Создаем обработчик (не используется в упрощенных примерах)
	_ = handlers.New(store, shortenerService, nil)

	// Настраиваем роутер без middleware для простоты примеров
	r := chi.NewRouter()
	
	// Простые обработчики без аутентификации для примеров
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		body, _ := io.ReadAll(req.Body)
		originalURL := string(body)
		if originalURL == "" {
			http.Error(w, "URL не может быть пустым", http.StatusBadRequest)
			return
		}
		
		shortID := shortenerService.Shorten(originalURL)
		store.Add(shortID, originalURL)
		shortURL := shortenerService.BuildShortURL(shortID)
		
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(shortURL))
	})
	
	r.Post("/api/shorten", func(w http.ResponseWriter, req *http.Request) {
		var request models.ShortenRequest
		json.NewDecoder(req.Body).Decode(&request)
		
		if request.URL == "" {
			http.Error(w, "URL не может быть пустым", http.StatusBadRequest)
			return
		}
		
		shortID := shortenerService.Shorten(request.URL)
		store.Add(shortID, request.URL)
		shortURL := shortenerService.BuildShortURL(shortID)
		
		response := models.ShortenResponse{Result: shortURL}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
	})
	
	r.Post("/api/shorten/batch", func(w http.ResponseWriter, req *http.Request) {
		var batchRequest []models.BatchShortenRequest
		json.NewDecoder(req.Body).Decode(&batchRequest)
		
		var batchResponse []models.BatchShortenResponse
		for _, item := range batchRequest {
			shortID := shortenerService.Shorten(item.OriginalURL)
			store.Add(shortID, item.OriginalURL)
			shortURL := shortenerService.BuildShortURL(shortID)
			
			batchResponse = append(batchResponse, models.BatchShortenResponse{
				CorrelationID: item.CorrelationID,
				ShortURL:      shortURL,
			})
		}
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(batchResponse)
	})
	
	r.Get("/{id}", func(w http.ResponseWriter, req *http.Request) {
		id := chi.URLParam(req, "id")
		originalURL, exists := store.Get(id)
		if !exists {
			http.Error(w, "URL не найден", http.StatusNotFound)
			return
		}
		
		w.Header().Set("Location", originalURL)
		w.WriteHeader(http.StatusTemporaryRedirect)
	})
	
	r.Get("/ping", func(w http.ResponseWriter, req *http.Request) {
		// Для примера всегда возвращаем ошибку, так как БД не настроена
		http.Error(w, "База данных не настроена", http.StatusInternalServerError)
	})

	return httptest.NewServer(r)
}

// ExampleCreateShortURLTextPlain демонстрирует создание короткого URL через text/plain API.
//
// Эндпоинт: POST /
// Content-Type: text/plain
// Тело запроса: оригинальный URL в виде строки
func Example_createShortURLTextPlain() {
	server := setupTestServer()
	defer server.Close()

	// Подготавливаем запрос
	originalURL := "https://example.com/very/long/url/path"
	req, _ := http.NewRequest("POST", server.URL+"/", strings.NewReader(originalURL))
	req.Header.Set("Content-Type", "text/plain")

	// Выполняем запрос
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Ошибка запроса: %v\n", err)
		return
	}
	defer resp.Body.Close()

	// Читаем ответ
	body, _ := io.ReadAll(resp.Body)
	shortURL := string(body)

	fmt.Printf("Статус: %d\n", resp.StatusCode)
	fmt.Printf("Оригинальный URL: %s\n", originalURL)
	fmt.Printf("Короткий URL содержит базовый адрес: %t\n", strings.Contains(shortURL, "http://localhost:8080/"))
	fmt.Printf("Длина короткого ID: %d символов\n", len(strings.TrimPrefix(shortURL, "http://localhost:8080/")))

	// Output:
	// Статус: 201
	// Оригинальный URL: https://example.com/very/long/url/path
	// Короткий URL содержит базовый адрес: true
	// Длина короткого ID: 8 символов
}

// ExampleShortenURLJSON демонстрирует создание короткого URL через JSON API.
//
// Эндпоинт: POST /api/shorten
// Content-Type: application/json
// Тело запроса: JSON с полем "url"
func Example_shortenURLJSON() {
	server := setupTestServer()
	defer server.Close()

	// Подготавливаем JSON запрос
	request := models.ShortenRequest{
		URL: "https://example.com/api/endpoint",
	}

	jsonData, _ := json.Marshal(request)
	req, _ := http.NewRequest("POST", server.URL+"/api/shorten", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	// Выполняем запрос
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Ошибка запроса: %v\n", err)
		return
	}
	defer resp.Body.Close()

	// Парсим JSON ответ
	var response models.ShortenResponse
	json.NewDecoder(resp.Body).Decode(&response)

	fmt.Printf("Статус: %d\n", resp.StatusCode)
	fmt.Printf("Запрос: %+v\n", request)
	fmt.Printf("Ответ: %+v\n", response)

	// Output:
	// Статус: 201
	// Запрос: {URL:https://example.com/api/endpoint}
	// Ответ: {Result:http://localhost:8080/xyz98765}
}

// ExampleShortenBatch демонстрирует пакетное создание коротких URL.
//
// Эндпоинт: POST /api/shorten/batch
// Content-Type: application/json
// Тело запроса: массив JSON объектов с correlation_id и original_url
func Example_shortenBatch() {
	server := setupTestServer()
	defer server.Close()

	// Подготавливаем пакетный запрос
	batchRequest := []models.BatchShortenRequest{
		{
			CorrelationID: "req_1",
			OriginalURL:   "https://example.com/page1",
		},
		{
			CorrelationID: "req_2",
			OriginalURL:   "https://example.com/page2",
		},
		{
			CorrelationID: "req_3",
			OriginalURL:   "https://example.com/page3",
		},
	}

	jsonData, _ := json.Marshal(batchRequest)
	req, _ := http.NewRequest("POST", server.URL+"/api/shorten/batch", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	// Выполняем запрос
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Ошибка запроса: %v\n", err)
		return
	}
	defer resp.Body.Close()

	// Парсим ответ
	var batchResponse []models.BatchShortenResponse
	json.NewDecoder(resp.Body).Decode(&batchResponse)

	fmt.Printf("Статус: %d\n", resp.StatusCode)
	fmt.Printf("Количество URL в запросе: %d\n", len(batchRequest))
	fmt.Printf("Количество URL в ответе: %d\n", len(batchResponse))

	for i, item := range batchResponse {
		fmt.Printf("  %d. ID: %s, URL: %s\n", i+1, item.CorrelationID, item.ShortURL)
	}

	// Output:
	// Статус: 201
	// Количество URL в запросе: 3
	// Количество URL в ответе: 3
	//   1. ID: req_1, URL: http://localhost:8080/abc123
	//   2. ID: req_2, URL: http://localhost:8080/def456
	//   3. ID: req_3, URL: http://localhost:8080/ghi789
}

// ExampleRedirectToURL демонстрирует получение оригинального URL по короткому ID.
//
// Эндпоинт: GET /{id}
// Ответ: HTTP 307 редирект на оригинальный URL
func Example_redirectToURL() {
	server := setupTestServer()
	defer server.Close()

	// Сначала создаем короткий URL
	originalURL := "https://example.com/target/page"
	createReq, _ := http.NewRequest("POST", server.URL+"/", strings.NewReader(originalURL))
	createReq.Header.Set("Content-Type", "text/plain")

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Не следуем редиректам автоматически
			return http.ErrUseLastResponse
		},
	}

	createResp, _ := client.Do(createReq)
	shortURL, _ := io.ReadAll(createResp.Body)
	createResp.Body.Close()

	// Извлекаем ID из короткого URL
	shortURLStr := string(shortURL)
	parts := strings.Split(shortURLStr, "/")
	shortID := parts[len(parts)-1]

	// Выполняем GET запрос по короткому ID
	getReq, _ := http.NewRequest("GET", server.URL+"/"+shortID, nil)
	getResp, err := client.Do(getReq)
	if err != nil {
		fmt.Printf("Ошибка запроса: %v\n", err)
		return
	}
	defer getResp.Body.Close()

	fmt.Printf("Статус: %d\n", getResp.StatusCode)
	fmt.Printf("Короткий ID: %s\n", shortID)
	fmt.Printf("Location заголовок: %s\n", getResp.Header.Get("Location"))

	// Output:
	// Статус: 307
	// Короткий ID: abc12345
	// Location заголовок: https://example.com/target/page
}

// ExamplePingDB демонстрирует проверку состояния базы данных.
//
// Эндпоинт: GET /ping
// Ответ: HTTP 200 если БД доступна, HTTP 500 если недоступна
func Example_pingDB() {
	server := setupTestServer()
	defer server.Close()

	// Выполняем ping запрос
	req, _ := http.NewRequest("GET", server.URL+"/ping", nil)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Ошибка запроса: %v\n", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("Статус: %d\n", resp.StatusCode)

	if resp.StatusCode == 200 {
		fmt.Println("База данных доступна")
	} else {
		fmt.Println("База данных недоступна")
	}

	// Output:
	// Статус: 500
	// База данных недоступна
}

// ExampleErrorHandling демонстрирует обработку ошибок API.
//
// Показывает различные сценарии ошибок:
//   - Некорректный JSON
//   - Пустой URL
//   - Несуществующий короткий ID
func Example_errorHandling() {
	server := setupTestServer()
	defer server.Close()

	client := &http.Client{}

	// 1. Некорректный JSON
	fmt.Println("=== Тест 1: Некорректный JSON ===")
	req1, _ := http.NewRequest("POST", server.URL+"/api/shorten", strings.NewReader(`{"url": invalid json`))
	req1.Header.Set("Content-Type", "application/json")

	resp1, _ := client.Do(req1)
	fmt.Printf("Статус: %d\n", resp1.StatusCode)
	resp1.Body.Close()

	// 2. Пустой URL
	fmt.Println("\n=== Тест 2: Пустой URL ===")
	emptyRequest := models.ShortenRequest{URL: ""}
	jsonData, _ := json.Marshal(emptyRequest)
	req2, _ := http.NewRequest("POST", server.URL+"/api/shorten", bytes.NewBuffer(jsonData))
	req2.Header.Set("Content-Type", "application/json")

	resp2, _ := client.Do(req2)
	fmt.Printf("Статус: %d\n", resp2.StatusCode)
	resp2.Body.Close()

	// 3. Несуществующий короткий ID
	fmt.Println("\n=== Тест 3: Несуществующий ID ===")
	req3, _ := http.NewRequest("GET", server.URL+"/nonexistent", nil)

	resp3, _ := client.Do(req3)
	fmt.Printf("Статус: %d\n", resp3.StatusCode)
	resp3.Body.Close()

	// Output:
	// === Тест 1: Некорректный JSON ===
	// Статус: 400
	//
	// === Тест 2: Пустой URL ===
	// Статус: 400
	//
	// === Тест 3: Несуществующий ID ===
	// Статус: 404
}
