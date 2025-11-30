package accrual

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"time"

	"go.uber.org/zap"
	"golang.org/x/time/rate"

	"github.com/SZabrodskii/gophermart-stas/internal/models"
)

type httpClient struct {
	client  *http.Client
	config  ClientConfig
	limiter *rate.Limiter
	logger  *zap.Logger
}

func NewClient(config ClientConfig, logger *zap.Logger) Client {
	limiter := rate.NewLimiter(
		rate.Every(time.Minute/time.Duration(config.RateLimit.RequestsPerMinute)),
		config.RateLimit.BurstSize,
	)

	return &httpClient{
		client: &http.Client{
			Timeout: config.Timeout,
		},
		config:  config,
		limiter: limiter,
		logger:  logger,
	}
}

func (c *httpClient) GetOrderAccrual(ctx context.Context, orderNumber string) (*models.AccrualResponse, error) {
	url := fmt.Sprintf("%s/api/orders/%s", c.config.BaseURL, orderNumber)

	var response *models.AccrualResponse
	var lastErr error

	for attempt := 0; attempt <= c.config.Retry.MaxRetries; attempt++ {
		if err := c.limiter.Wait(ctx); err != nil {
			return nil, fmt.Errorf("rate limiter wait failed: %w", err)
		}

		resp, err := c.makeRequest(ctx, url)
		if err != nil {
			lastErr = err

			if attempt < c.config.Retry.MaxRetries {
				delay := c.calculateBackoffDelay(attempt)
				c.logger.Info("Request failed, retrying",
					zap.Int("attempt", attempt+1),
					zap.Int("max_attempts", c.config.Retry.MaxRetries+1),
					zap.String("delay", delay.String()),
					zap.String("error", err.Error()))

				select {
				case <-time.After(delay):
					continue
				case <-ctx.Done():
					return nil, ctx.Err()
				}
			}
			continue
		}

		response = resp
		break
	}

	if response == nil {
		return nil, fmt.Errorf("all retry attempts failed, last error: %w", lastErr)
	}

	return response, nil
}

func (c *httpClient) GetOrderInfo(ctx context.Context, orderNumber string) (*models.AccrualResponse, error) {
	return c.GetOrderAccrual(ctx, orderNumber)
}

func (c *httpClient) makeRequest(ctx context.Context, url string) (*models.AccrualResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request execution failed: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var accrualResp models.AccrualResponse
		if err := json.NewDecoder(resp.Body).Decode(&accrualResp); err != nil {
			return nil, fmt.Errorf("decode response failed: %w", err)
		}
		return &accrualResp, nil

	case http.StatusNoContent:
		return nil, fmt.Errorf("order not registered in accrual system")

	case http.StatusTooManyRequests:
		retryAfter := c.parseRetryAfter(resp)
		return nil, &RateLimitError{
			RetryAfter: retryAfter,
			Message:    "rate limit exceeded",
		}

	case http.StatusInternalServerError:
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("internal server error (500): %s", string(body))

	default:
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}
}

func (c *httpClient) calculateBackoffDelay(attempt int) time.Duration {
	delay := float64(c.config.Retry.InitialDelay) * math.Pow(c.config.Retry.BackoffFactor, float64(attempt))
	maxDelay := float64(c.config.Retry.MaxDelay)

	if delay > maxDelay {
		delay = maxDelay
	}

	return time.Duration(delay)
}

func (c *httpClient) parseRetryAfter(resp *http.Response) time.Duration {
	retryAfterHeader := resp.Header.Get("Retry-After")
	if retryAfterHeader == "" {
		return c.config.RateLimit.RetryAfter
	}

	if seconds, err := strconv.Atoi(retryAfterHeader); err == nil {
		return time.Duration(seconds) * time.Second
	}

	return c.config.RateLimit.RetryAfter
}

func (c *httpClient) Close() error {
	c.logger.Info("Accrual client closed")
	return nil
}

type RateLimitError struct {
	RetryAfter time.Duration
	Message    string
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("%s (retry after %v)", e.Message, e.RetryAfter)
}

func IsRateLimitError(err error) (*RateLimitError, bool) {
	var rateLimitErr *RateLimitError
	if errors.As(err, &rateLimitErr) {
		return rateLimitErr, true
	}
	return nil, false
}
