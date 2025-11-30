package controllers

import (
	"errors"
	"net/http"

	"github.com/SZabrodskii/gophermart-stas/internal/domain"
	"github.com/SZabrodskii/gophermart-stas/internal/models"
	"github.com/SZabrodskii/gophermart-stas/internal/services"
	"github.com/SZabrodskii/gophermart-stas/internal/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type BalanceControllerGin struct {
	logger         *zap.Logger
	balanceService domain.BalanceServiceI
}

func NewBalanceControllerGin(logger *zap.Logger, balanceService domain.BalanceServiceI) *BalanceControllerGin {
	return &BalanceControllerGin{
		logger:         logger,
		balanceService: balanceService,
	}
}

func (bc *BalanceControllerGin) GetBalance(c *gin.Context) {
	userID, ok := GetUserIDFromGinContext(c)
	if !ok {
		bc.logger.Warn("User not authenticated")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	balance, err := bc.balanceService.GetBalance(userID)
	if err != nil {
		bc.logger.Error("Failed to get balance", zap.Error(err), zap.Uint("user_id", userID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get balance"})
		return
	}

	bc.logger.Info("Balance retrieved successfully", zap.Uint("user_id", userID), zap.Float64("balance", balance.Current))
	c.JSON(http.StatusOK, balance)
}

func (bc *BalanceControllerGin) Withdraw(c *gin.Context) {
	userID, ok := GetUserIDFromGinContext(c)
	if !ok {
		bc.logger.Warn("User not authenticated")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req models.WithdrawalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bc.logger.Warn("Invalid request format", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	if !utils.ValidateOrderNumber(req.Order) {
		bc.logger.Warn("Invalid order number format (Luhn algorithm)", zap.String("order", req.Order))
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Invalid order number format"})
		return
	}

	err := bc.balanceService.WithdrawBalance(userID, req.Order, req.Sum)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInsufficientFunds):
			bc.logger.Warn("Insufficient funds for withdrawal", zap.Uint("user_id", userID), zap.Float64("amount", req.Sum))
			c.JSON(http.StatusPaymentRequired, gin.H{"error": "Insufficient funds"})
		case errors.Is(err, services.ErrInvalidOrderNumber):
			bc.logger.Warn("Invalid order number format", zap.String("order", req.Order))
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Invalid order number format"})
		default:
			bc.logger.Error("Failed to withdraw", zap.Error(err), zap.Uint("user_id", userID), zap.String("order", req.Order), zap.Float64("amount", req.Sum))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	bc.logger.Info("Withdrawal successful", zap.Uint("user_id", userID), zap.String("order", req.Order), zap.Float64("amount", req.Sum))
	c.JSON(http.StatusOK, gin.H{"message": "Withdrawal successful"})
}

func (bc *BalanceControllerGin) GetWithdrawals(c *gin.Context) {
	userID, ok := GetUserIDFromGinContext(c)
	if !ok {
		bc.logger.Warn("User not authenticated")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	withdrawals, err := bc.balanceService.GetWithdrawals(userID)
	if err != nil {
		bc.logger.Error("Failed to get withdrawals", zap.Error(err), zap.Uint("user_id", userID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get withdrawals"})
		return
	}

	if len(withdrawals) == 0 {
		bc.logger.Info("No withdrawals found", zap.Uint("user_id", userID))
		c.Status(http.StatusNoContent)
		return
	}

	bc.logger.Info("Withdrawals retrieved successfully", zap.Int("count", len(withdrawals)), zap.Uint("user_id", userID))
	c.JSON(http.StatusOK, withdrawals)
}
