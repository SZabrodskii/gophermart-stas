package accrual

import (
	"context"
	"os"
	"strconv"
	"time"

	"github.com/SZabrodskii/gophermart-stas/internal/config"
	"github.com/SZabrodskii/gophermart-stas/internal/models"
	"go.uber.org/zap"
)

func NewNoOpClient(logger *zap.Logger) Client {
	return &noOpClient{
		logger: logger,
	}
}

type noOpClient struct {
	logger *zap.Logger
}

func (c *noOpClient) GetOrderAccrual(ctx context.Context, orderNumber string) (*models.AccrualResponse, error) {
	return nil, nil
}

func (c *noOpClient) GetOrderInfo(ctx context.Context, orderNumber string) (*models.AccrualResponse, error) {
	return nil, nil
}

func (c *noOpClient) Close() error {
	return nil
}

func ProvideClient(cfg *config.Config, logger *zap.Logger) (Client, error) {
	baseURL := cfg.AccrualAddress
	if baseURL == "" {
		logger.Warn("Accrual system address not configured, creating no-op client")
		return NewNoOpClient(logger), nil
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
		zap.String("base_url", config.BaseURL),
		zap.String("timeout", config.Timeout.String()),
		zap.Int("max_retries", config.Retry.MaxRetries),
		zap.Int("requests_per_minute", config.RateLimit.RequestsPerMinute))

	return client, nil
}
