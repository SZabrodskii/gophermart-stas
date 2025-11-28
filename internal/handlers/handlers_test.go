package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/SZabrodskii/gophermart-stas/internal/models"
	"github.com/SZabrodskii/gophermart-stas/internal/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type mockOrderService struct {
	uploadOrderFunc   func(userID uint, orderNumber string) error
	getUserOrdersFunc func(userID uint) ([]models.Order, error)
}

func (m *mockOrderService) UploadOrder(userID uint, orderNumber string) error {
	if m.uploadOrderFunc != nil {
		return m.uploadOrderFunc(userID, orderNumber)
	}
	return nil
}

func (m *mockOrderService) GetUserOrders(userID uint) ([]models.Order, error) {
	if m.getUserOrdersFunc != nil {
		return m.getUserOrdersFunc(userID)
	}
	return []models.Order{}, nil
}

type mockUserService struct{}

func (m *mockUserService) Register(login, password string) (string, error) { return "", nil }
func (m *mockUserService) Login(login, password string) (string, error)    { return "", nil }

type mockBalanceService struct{}

func (m *mockBalanceService) GetBalance(userID uint) (*models.Balance, error) { return nil, nil }
func (m *mockBalanceService) WithdrawBalance(userID uint, orderNumber string, amount float64) error {
	return nil
}
func (m *mockBalanceService) GetWithdrawals(userID uint) ([]models.Withdrawal, error) {
	return nil, nil
}

func TestUploadOrder(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		body           string
		userID         uint
		mockError      error
		expectedStatus int
	}{
		{
			name:           "valid order",
			body:           "4561261212345467",
			userID:         1,
			mockError:      nil,
			expectedStatus: http.StatusAccepted,
		},
		{
			name:           "invalid order number",
			body:           "123",
			userID:         1,
			mockError:      services.ErrInvalidOrderNumber,
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name:           "order exists",
			body:           "4561261212345467",
			userID:         1,
			mockError:      services.ErrOrderExists,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "order exists for other user",
			body:           "4561261212345467",
			userID:         1,
			mockError:      services.ErrOrderExistsOtherUser,
			expectedStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockOrder := &mockOrderService{
				uploadOrderFunc: func(userID uint, orderNumber string) error {
					return tt.mockError
				},
			}

			handler := &Handler{
				logger:         zap.NewNop(),
				orderService:   mockOrder,
				userService:    &mockUserService{},
				balanceService: &mockBalanceService{},
			}

			router := gin.New()
			router.POST("/orders", func(c *gin.Context) {
				c.Set("userID", tt.userID)
				handler.UploadOrder(c)
			})

			req := httptest.NewRequest("POST", "/orders", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "text/plain")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestGetOrders(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		userID         uint
		orders         []models.Order
		expectedStatus int
	}{
		{
			name:           "no orders",
			userID:         1,
			orders:         []models.Order{},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:   "with orders",
			userID: 1,
			orders: []models.Order{
				{Number: "4561261212345467", UserID: 1, Status: "NEW"},
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockOrder := &mockOrderService{
				getUserOrdersFunc: func(userID uint) ([]models.Order, error) {
					return tt.orders, nil
				},
			}

			handler := &Handler{
				logger:         zap.NewNop(),
				orderService:   mockOrder,
				userService:    &mockUserService{},
				balanceService: &mockBalanceService{},
			}

			router := gin.New()
			router.GET("/orders", func(c *gin.Context) {
				c.Set("userID", tt.userID)
				handler.GetOrders(c)
			})

			req := httptest.NewRequest("GET", "/orders", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var orders []models.Order
				json.Unmarshal(w.Body.Bytes(), &orders)
				if len(orders) != len(tt.orders) {
					t.Errorf("Expected %d orders, got %d", len(tt.orders), len(orders))
				}
			}
		})
	}
}
