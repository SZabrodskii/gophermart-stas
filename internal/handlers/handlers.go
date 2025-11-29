package handlers

import (
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/SZabrodskii/gophermart-stas/internal/domain"
	"github.com/SZabrodskii/gophermart-stas/internal/models"
	"github.com/SZabrodskii/gophermart-stas/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/gopybara/httpbara"
)

type Handler struct {
	logger         httpbara.Logger
	userService    domain.UserServiceI
	orderService   domain.OrderServiceI
	balanceService domain.BalanceServiceI
}

func New(
	logger httpbara.Logger,
	userService domain.UserServiceI,
	orderService domain.OrderServiceI,
	balanceService domain.BalanceServiceI,
) *Handler {
	return &Handler{
		logger:         logger,
		userService:    userService,
		orderService:   orderService,
		balanceService: balanceService,
	}
}

func (h *Handler) Register(c *gin.Context) {
	var req models.UserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid request format", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	if req.Login == "" || req.Password == "" {
		h.logger.Warn("Missing login or password")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Login and password are required"})
		return
	}

	token, err := h.userService.Register(req.Login, req.Password)
	if err != nil {
		if errors.Is(err, services.ErrUserExists) {
			c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
			return
		}
		h.logger.Error("Registration failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.Header("Authorization", "Bearer "+token)
	c.Status(http.StatusOK)
}

func (h *Handler) Login(c *gin.Context) {
	var req models.UserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid request format", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	if req.Login == "" || req.Password == "" {
		h.logger.Warn("Missing login or password")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Login and password are required"})
		return
	}

	token, err := h.userService.Login(req.Login, req.Password)
	if err != nil {
		if errors.Is(err, services.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}
		h.logger.Error("Login failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.Header("Authorization", "Bearer "+token)
	c.Status(http.StatusOK)
}

func (h *Handler) getUserIDFromContext(c *gin.Context) (uint, error) {
	userID, exists := c.Get("userID")
	if !exists {
		return 0, errors.New("user not authenticated")
	}

	id, ok := userID.(uint)
	if !ok {
		return 0, errors.New("invalid user ID format")
	}

	return id, nil
}

func (h *Handler) UploadOrder(c *gin.Context) {
	userID, err := h.getUserIDFromContext(c)
	if err != nil {
		h.logger.Warn("Failed to get user ID", "error", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		h.logger.Warn("Failed to read request body", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	orderNumber := strings.TrimSpace(string(body))
	if orderNumber == "" {
		h.logger.Warn("Empty order number")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Order number is required"})
		return
	}

	err = h.orderService.UploadOrder(userID, orderNumber)
	if err != nil {
		if errors.Is(err, services.ErrInvalidOrderNumber) {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Invalid order number"})
			return
		}
		if errors.Is(err, services.ErrOrderExists) {
			c.Status(http.StatusOK)
			return
		}
		if errors.Is(err, services.ErrOrderExistsOtherUser) {
			c.JSON(http.StatusConflict, gin.H{"error": "Order already exists for another user"})
			return
		}

		h.logger.Error("Failed to upload order", "error", err, "userID", userID, "orderNumber", orderNumber)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	h.logger.Info("Order uploaded successfully", "userID", userID, "orderNumber", orderNumber)
	c.Status(http.StatusAccepted)
}

func (h *Handler) GetOrders(c *gin.Context) {
	userID, err := h.getUserIDFromContext(c)
	if err != nil {
		h.logger.Warn("Failed to get user ID", "error", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	orders, err := h.orderService.GetUserOrders(userID)
	if err != nil {
		h.logger.Error("Failed to get user orders", "error", err, "userID", userID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	if len(orders) == 0 {
		c.Status(http.StatusNoContent)
		return
	}

	h.logger.Info("Orders retrieved successfully", "userID", userID, "count", len(orders))
	c.JSON(http.StatusOK, orders)
}

func (h *Handler) GetBalance(c *gin.Context) {
	h.logger.Info("Get balance endpoint called")
	c.JSON(http.StatusNotImplemented, gin.H{"message": "Get balance endpoint - not implemented yet"})
}

func (h *Handler) WithdrawBalance(c *gin.Context) {
	h.logger.Info("Withdraw balance endpoint called")
	c.JSON(http.StatusNotImplemented, gin.H{"message": "Withdraw balance endpoint - not implemented yet"})
}

func (h *Handler) GetWithdrawals(c *gin.Context) {
	h.logger.Info("Get withdrawals endpoint called")
	c.JSON(http.StatusNotImplemented, gin.H{"message": "Get withdrawals endpoint - not implemented yet"})
}
