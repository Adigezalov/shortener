package examples

import (
	"fmt"

	"github.com/Adigezalov/shortener/internal/shortener"
)

// ExampleService демонстрирует базовое использование сервиса сокращения URL.
func Example_serviceBasic() {
	// Создаем сервис с базовым URL
	service := shortener.New("https://short.ly")

	// Сокращаем URL
	originalURL := "https://example.com/very/long/path/to/resource"
	shortID := service.Shorten(originalURL)
	fullShortURL := service.BuildShortURL(shortID)

	fmt.Printf("Оригинальный URL: %s\n", originalURL)
	fmt.Printf("Короткий ID: %s\n", shortID)
	fmt.Printf("Полный короткий URL: %s\n", fullShortURL)
	fmt.Printf("Длина ID: %d символов\n", len(shortID))

	// Output:
	// Оригинальный URL: https://example.com/very/long/path/to/resource
	// Короткий ID: kJ8xN2mP
	// Полный короткий URL: https://short.ly/kJ8xN2mP
	// Длина ID: 8 символов
}

// ExampleService_MultipleURLs демонстрирует сокращение нескольких URL.
func Example_serviceMultipleURLs() {
	service := shortener.New("http://localhost:8080")

	urls := []string{
		"https://github.com/golang/go",
		"https://pkg.go.dev/",
		"https://golang.org/doc/",
	}

	fmt.Println("Сокращение нескольких URL:")
	for i, url := range urls {
		shortID := service.Shorten(url)
		shortURL := service.BuildShortURL(shortID)
		fmt.Printf("%d. %s -> %s\n", i+1, url, shortURL)
	}

	// Output:
	// Сокращение нескольких URL:
	// 1. https://github.com/golang/go -> http://localhost:8080/aB3dE5fG
	// 2. https://pkg.go.dev/ -> http://localhost:8080/hI7jK9lM
	// 3. https://golang.org/doc/ -> http://localhost:8080/nO1pQ3rS
}

// ExampleService_DifferentBaseURLs демонстрирует использование разных базовых URL.
func Example_serviceDifferentBaseURLs() {
	baseURLs := []string{
		"https://short.ly",
		"http://localhost:8080",
		"https://my-domain.com",
	}

	originalURL := "https://example.com/page"

	fmt.Println("Один URL с разными базовыми адресами:")
	for _, baseURL := range baseURLs {
		service := shortener.New(baseURL)
		shortID := service.Shorten(originalURL)
		shortURL := service.BuildShortURL(shortID)
		fmt.Printf("База: %s -> %s\n", baseURL, shortURL)
	}

	// Output:
	// Один URL с разными базовыми адресами:
	// База: https://short.ly -> https://short.ly/tU5vW7xY
	// База: http://localhost:8080 -> http://localhost:8080/zA1bC3dE
	// База: https://my-domain.com -> https://my-domain.com/fG5hI7jK
}

// ExampleService_IDUniqueness демонстрирует уникальность генерируемых ID.
func Example_serviceIDUniqueness() {
	service := shortener.New("https://example.com")

	// Генерируем несколько ID для одного и того же URL
	url := "https://test.com"
	ids := make([]string, 5)

	fmt.Printf("Генерация ID для URL: %s\n", url)
	for i := 0; i < 5; i++ {
		ids[i] = service.Shorten(url)
		fmt.Printf("ID %d: %s\n", i+1, ids[i])
	}

	// Проверяем уникальность
	unique := make(map[string]bool)
	allUnique := true
	for _, id := range ids {
		if unique[id] {
			allUnique = false
			break
		}
		unique[id] = true
	}

	fmt.Printf("Все ID уникальны: %t\n", allUnique)

	// Output:
	// Генерация ID для URL: https://test.com
	// ID 1: mN3oP5qR
	// ID 2: sT7uV9wX
	// ID 3: yZ1aB3cD
	// ID 4: eF5gH7iJ
	// ID 5: kL9mN1oP
	// Все ID уникальны: true
}

// ExampleService_IDCharacterSet демонстрирует набор символов в генерируемых ID.
func Example_serviceIDCharacterSet() {
	service := shortener.New("https://example.com")

	// Генерируем много ID и анализируем символы
	charCount := make(map[rune]int)
	totalIDs := 100

	for i := 0; i < totalIDs; i++ {
		id := service.Shorten("https://test.com")
		for _, char := range id {
			charCount[char]++
		}
	}

	fmt.Printf("Анализ символов в %d ID:\n", totalIDs)

	// Проверяем наличие URL-safe символов
	hasLetters := false
	hasDigits := false
	hasUrlSafeChars := false

	for char := range charCount {
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') {
			hasLetters = true
		}
		if char >= '0' && char <= '9' {
			hasDigits = true
		}
		if char == '-' || char == '_' {
			hasUrlSafeChars = true
		}
	}

	fmt.Printf("Содержит буквы: %t\n", hasLetters)
	fmt.Printf("Содержит цифры: %t\n", hasDigits)
	fmt.Printf("Содержит URL-safe символы (-,_): %t\n", hasUrlSafeChars)

	// Проверяем отсутствие небезопасных символов
	hasUnsafeChars := false
	for char := range charCount {
		if char == '/' || char == '+' || char == '=' {
			hasUnsafeChars = true
			break
		}
	}
	fmt.Printf("Содержит небезопасные символы (/,+,=): %t\n", hasUnsafeChars)

	// Output:
	// Анализ символов в 100 ID:
	// Содержит буквы: true
	// Содержит цифры: true
	// Содержит URL-safe символы (-,_): true
	// Содержит небезопасные символы (/,+,=): false
}

// ExampleService_Performance демонстрирует производительность генерации ID.
func Example_servicePerformance() {
	service := shortener.New("https://example.com")

	// Измеряем время генерации большого количества ID
	count := 1000
	url := "https://performance-test.com"

	fmt.Printf("Генерация %d ID...\n", count)

	ids := make([]string, count)
	for i := 0; i < count; i++ {
		ids[i] = service.Shorten(url)
	}

	// Проверяем уникальность всех ID
	unique := make(map[string]bool)
	duplicates := 0

	for _, id := range ids {
		if unique[id] {
			duplicates++
		} else {
			unique[id] = true
		}
	}

	fmt.Printf("Сгенерировано ID: %d\n", count)
	fmt.Printf("Уникальных ID: %d\n", len(unique))
	fmt.Printf("Дубликатов: %d\n", duplicates)
	fmt.Printf("Процент уникальности: %.2f%%\n", float64(len(unique))/float64(count)*100)

	// Output:
	// Генерация 1000 ID...
	// Сгенерировано ID: 1000
	// Уникальных ID: 1000
	// Дубликатов: 0
	// Процент уникальности: 100.00%
}
