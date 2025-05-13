package main

import (
	"github.com/Adigezalov/shortener/internal/app"
	"github.com/Adigezalov/shortener/internal/config"
	"go.uber.org/zap"
)

func main() {
	cfg := config.ParseFlags()

	server := app.NewServer(*cfg)

	defer func() {
		if err := server.Logger.Sync(); err != nil {
		}
	}()

	if err := server.ListenAndServe(); err != nil {
		server.Logger.Fatal("Server failed", zap.Error(err))
	}
}
