package main

import (
	"github.com/Adigezalov/shortener/internal/app"
	"github.com/Adigezalov/shortener/internal/config"
	"log"
)

func main() {
	cfg := config.ParseFlags()

	server := app.NewServer(*cfg)

	log.Printf("Server starting on %s\n", cfg.ServerAddress)
	log.Printf("Base URL for short links: %s\n", cfg.BaseURL)

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
