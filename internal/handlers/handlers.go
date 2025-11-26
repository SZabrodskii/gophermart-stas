package handlers

import (
	"net/http"

	"github.com/SZabrodskii/gophermart-stas/internal/database"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Handler struct {
	db     *database.DB
	logger *zap.Logger
}

func New(db *database.DB, logger *zap.Logger) *Handler {
	return &Handler{
		db:     db,
		logger: logger,
	}
}

func (h *Handler) Register(c *gin.Context) {
	h.logger.Info("Register endpoint called")
	c.JSON(http.StatusNotImplemented, gin.H{"message": "Register endpoint - not implemented yet"})
}

func (h *Handler) Login(c *gin.Context) {
	h.logger.Info("Login endpoint called")
	c.JSON(http.StatusNotImplemented, gin.H{"message": "Login endpoint - not implemented yet"})
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
