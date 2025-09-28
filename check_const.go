package main

import (
	"fmt"
	"github.com/Adigezalov/shortener/internal/config"
)

func main() {
	fmt.Printf("DefaultDatabaseDSN: %q\n", config.DefaultDatabaseDSN)
}
