package main

import (
	"fmt"
	"os"
)

func main() {
	// Очищаем переменные окружения для чистого теста
	os.Unsetenv("SERVER_ADDRESS")
	os.Unsetenv("BASE_URL")
	os.Unsetenv("DATABASE_DSN")

	fmt.Printf("Server Address: %s\n", ":8080")
	fmt.Printf("Base URL: %s\n", "http://localhost:8080")
	fmt.Printf("File Storage: %s\n", "storage.json")
	fmt.Printf("Database DSN: %s\n", "")
	fmt.Printf("Profiling Enabled: %t\n", false)
}
