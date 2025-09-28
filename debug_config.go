package main

import (
	"fmt"
	"github.com/Adigezalov/shortener/internal/config"
	"os"
)

func main() {
	os.Unsetenv("SERVER_ADDRESS")
	os.Unsetenv("BASE_URL")
	os.Unsetenv("DATABASE_DSN")
	cfg := config.NewConfigFromEnv()
	fmt.Printf("Database DSN: '%s' (len=%d)\n", cfg.DatabaseDSN, len(cfg.DatabaseDSN))
	for i, r := range cfg.DatabaseDSN {
		fmt.Printf("char[%d] = %q (code=%d)\n", i, r, int(r))
	}
}
