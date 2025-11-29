package accrual

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"github.com/SZabrodskii/gophermart-stas/internal/models"
	"github.com/SZabrodskii/gophermart-stas/pkg/logger"
)

func TestNewClient(t *testing.T) {
	log := zaptest.NewLogger(t)
	httpbaraLogger := logger.NewZapLogger(log)

	config := DefaultClientConfig("http://localhost:8081")
	client := NewClient(config, httpbaraLogger)

	assert.NotNil(t, client)
	assert.NoError(t, client.Close())
}

func TestHttpClient_GetOrderAccrual_Success(t *testing.T) {
	testResponse := models.AccrualResponse{
		Order:   "4561261212345467",
		Status:  StatusProcessed,
		Accrual: floatPtr(500.5),
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/orders/4561261212345467", r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(testResponse)
	}))
	defer server.Close()

	log := zaptest.NewLogger(t)
	httpbaraLogger := logger.NewZapLogger(log)
	config := DefaultClientConfig(server.URL)
	client := NewClient(config, httpbaraLogger)
	defer client.Close()

	ctx := context.Background()
	response, err := client.GetOrderAccrual(ctx, "4561261212345467")

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, testResponse.Order, response.Order)
	assert.Equal(t, testResponse.Status, response.Status)
	assert.Equal(t, *testResponse.Accrual, *response.Accrual)
}

func TestHttpClient_GetOrderAccrual_NoContent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	log := zaptest.NewLogger(t)
	httpbaraLogger := logger.NewZapLogger(log)
	config := DefaultClientConfig(server.URL)
	config.Retry.MaxRetries = 0
	client := NewClient(config, httpbaraLogger)
	defer client.Close()

	ctx := context.Background()
	response, err := client.GetOrderAccrual(ctx, "4561261212345467")

	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "order not registered")
}

func TestHttpClient_GetOrderAccrual_TooManyRequests(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "60")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	log := zaptest.NewLogger(t)
	httpbaraLogger := logger.NewZapLogger(log)
	config := DefaultClientConfig(server.URL)
	config.Retry.MaxRetries = 0
	client := NewClient(config, httpbaraLogger)
	defer client.Close()

	ctx := context.Background()
	response, err := client.GetOrderAccrual(ctx, "4561261212345467")

	assert.Error(t, err)
	assert.Nil(t, response)

	rateLimitErr, ok := IsRateLimitError(err)
	if assert.True(t, ok) {
		assert.Equal(t, 60*time.Second, rateLimitErr.RetryAfter)
	}
}

func floatPtr(f float64) *float64 {
	return &f
}
