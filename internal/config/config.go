package config

import (
	"flag"
	"log"
	"os"
)

type Config struct {
	RunAddress     string
	DatabaseURI    string
	AccrualAddress string
}

func New() *Config {
	var (
		runAddress     = flag.String("a", getEnv("RUN_ADDRESS", ":8080"), "server address")
		databaseURI    = flag.String("d", getEnv("DATABASE_URI", ""), "database connection string")
		accrualAddress = flag.String("r", getEnv("ACCRUAL_SYSTEM_ADDRESS", ""), "accrual system address")
	)
	flag.Parse()

	if *databaseURI == "" {
		log.Fatal("DATABASE_URI is required")
	}

	if *accrualAddress == "" {
		log.Fatal("ACCRUAL_SYSTEM_ADDRESS is required")
	}

	return &Config{
		RunAddress:     *runAddress,
		DatabaseURI:    *databaseURI,
		AccrualAddress: *accrualAddress,
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
