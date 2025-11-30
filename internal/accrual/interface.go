package accrual

import (
	"context"
	"time"

	"github.com/SZabrodskii/gophermart-stas/internal/models"
)

type Client interface {
	GetOrderAccrual(ctx context.Context, orderNumber string) (*models.AccrualResponse, error)
	GetOrderInfo(ctx context.Context, orderNumber string) (*models.AccrualResponse, error)
	Close() error
}

// Добавляем алиас для совместимости
type ClientI = Client

type ClientConfig struct {
	BaseURL   string
	Timeout   time.Duration
	Retry     RetryConfig
	RateLimit RateLimitConfig
}

func DefaultClientConfig(baseURL string) ClientConfig {
	return ClientConfig{
		BaseURL:   baseURL,
		Timeout:   30 * time.Second,
		Retry:     DefaultRetryConfig(),
		RateLimit: DefaultRateLimitConfig(),
	}
}
