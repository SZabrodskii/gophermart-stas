package accrual

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/SZabrodskii/gophermart-stas/internal/config"
	"github.com/gopybara/httpbara"
)

func ProvideClient(cfg *config.Config, logger httpbara.Logger) (Client, error) {
	baseURL := cfg.AccrualAddress
	if baseURL == "" {
		return nil, fmt.Errorf("accrual system address is required (use -r flag or ACCRUAL_SYSTEM_ADDRESS environment variable)")
	}

	config := DefaultClientConfig(baseURL)

	if timeoutStr := os.Getenv("ACCRUAL_TIMEOUT"); timeoutStr != "" {
		if timeout, err := time.ParseDuration(timeoutStr); err == nil {
			config.Timeout = timeout
		}
	}

	if retriesStr := os.Getenv("ACCRUAL_MAX_RETRIES"); retriesStr != "" {
		if retries, err := strconv.Atoi(retriesStr); err == nil && retries >= 0 {
			config.Retry.MaxRetries = retries
		}
	}

	if rpmStr := os.Getenv("ACCRUAL_REQUESTS_PER_MINUTE"); rpmStr != "" {
		if rpm, err := strconv.Atoi(rpmStr); err == nil && rpm > 0 {
			config.RateLimit.RequestsPerMinute = rpm
		}
	}

	client := NewClient(config, logger)

	logger.Info("Accrual client created",
		"base_url", config.BaseURL,
		"timeout", config.Timeout.String(),
		"max_retries", config.Retry.MaxRetries,
		"requests_per_minute", config.RateLimit.RequestsPerMinute)

	return client, nil
}
