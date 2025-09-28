package main

import (
	"fmt"
	"os"
)

func main() {
	os.Unsetenv("DATABASE_DSN")

	databaseDSN := ""
	if envDatabaseDSN := os.Getenv("DATABASE_DSN"); envDatabaseDSN != "" {
		databaseDSN = envDatabaseDSN
	}

	fmt.Printf("Database DSN: '%s'\n", databaseDSN)
	fmt.Printf("Length: %d\n", len(databaseDSN))
}
