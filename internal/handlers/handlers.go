package handlers

import (
	"errors"
	"net/http"

	"github.com/SZabrodskii/gophermart-stas/internal/database"
	"github.com/SZabrodskii/gophermart-stas/internal/models"
	"github.com/SZabrodskii/gophermart-stas/internal/services"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Handler struct {
	db          *database.DB
	logger      *zap.Logger
	userService *services.UserService
}

func New(db *database.DB, logger *zap.Logger) *Handler {
	return &Handler{
		db:          db,
		logger:      logger,
		userService: services.NewUserService(db, logger),
	}
}

func (h *Handler) Register(c *gin.Context) {
	var req models.UserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid request format", zap.Error(err))
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
		h.logger.Error("Registration failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.Header("Authorization", "Bearer "+token)
	c.Status(http.StatusOK)
}

func (h *Handler) Login(c *gin.Context) {
	var req models.UserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid request format", zap.Error(err))
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
		h.logger.Error("Login failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.Header("Authorization", "Bearer "+token)
	c.Status(http.StatusOK)
}

func (h *Handler) UploadOrder(c *gin.Context) {
	h.logger.Info("Upload order endpoint called")
	c.JSON(http.StatusNotImplemented, gin.H{"message": "Upload order endpoint - not implemented yet"})
}

func (h *Handler) GetOrders(c *gin.Context) {
	h.logger.Info("Get orders endpoint called")
	c.JSON(http.StatusNotImplemented, gin.H{"message": "Get orders endpoint - not implemented yet"})
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
