package main

import (
	"fmt"
	"github.com/Adigezalov/shortener/internal/app"
	"log"
)

func main() {
	server := app.NewServer()

	fmt.Printf("Server starting on %s", app.Address)

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
