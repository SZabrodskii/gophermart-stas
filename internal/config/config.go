package config

import (
	"flag"
	"os"
)

type Config struct {
	RunAddress     string
	DatabaseURI    string
	AccrualAddress string
}

func New() (*Config, error) {
	cfg := &Config{}

	flag.StringVar(&cfg.RunAddress, "a", ":8080", "server address")
	flag.StringVar(&cfg.DatabaseURI, "d", "postgres:///postgres?host=/var/run/postgresql&sslmode=disable", "database connection string")
	flag.StringVar(&cfg.AccrualAddress, "r", "http://localhost:8081", "accrual system address")

	flag.Parse()

	if v, ok := os.LookupEnv("RUN_ADDRESS"); ok {
		cfg.RunAddress = v
	}
	if v, ok := os.LookupEnv("DATABASE_URI"); ok {
		cfg.DatabaseURI = v
	}
	if v, ok := os.LookupEnv("ACCRUAL_SYSTEM_ADDRESS"); ok {
		cfg.AccrualAddress = v
	}

	return cfg, nil
}
