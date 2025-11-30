package controllers

import (
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/SZabrodskii/gophermart-stas/internal/domain"
	"github.com/SZabrodskii/gophermart-stas/internal/services"
	"github.com/SZabrodskii/gophermart-stas/internal/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type OrderControllerGin struct {
	logger       *zap.Logger
	orderService domain.OrderServiceI
}

func NewOrderControllerGin(logger *zap.Logger, orderService domain.OrderServiceI) *OrderControllerGin {
	return &OrderControllerGin{
		logger:       logger,
		orderService: orderService,
	}
}

func (oc *OrderControllerGin) UploadOrder(c *gin.Context) {
	userID, ok := GetUserIDFromGinContext(c)
	if !ok {
		oc.logger.Warn("User not authenticated")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		oc.logger.Error("Failed to read request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}

	orderNumber := strings.TrimSpace(string(body))
	if orderNumber == "" {
		oc.logger.Warn("Empty order number")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Empty order number"})
		return
	}

	if !utils.ValidateOrderNumber(orderNumber) {
		oc.logger.Warn("Invalid order number format (Luhn algorithm)", zap.String("order_number", orderNumber))
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Invalid order number format"})
		return
	}

	err = oc.orderService.UploadOrder(userID, orderNumber)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidOrderNumber):
			oc.logger.Warn("Invalid order number format", zap.String("order_number", orderNumber))
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Invalid order number format"})
		case errors.Is(err, services.ErrOrderExists):
			oc.logger.Info("Order already uploaded by user", zap.String("order_number", orderNumber), zap.Uint("user_id", userID))
			c.JSON(http.StatusOK, gin.H{"message": "Order already uploaded"})
		case errors.Is(err, services.ErrOrderExistsOtherUser):
			oc.logger.Warn("Order already uploaded by different user", zap.String("order_number", orderNumber))
			c.JSON(http.StatusConflict, gin.H{"error": "Order already uploaded by another user"})
		default:
			oc.logger.Error("Failed to upload order", zap.Error(err), zap.String("order_number", orderNumber), zap.Uint("user_id", userID))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	oc.logger.Info("Order uploaded successfully", zap.String("order_number", orderNumber), zap.Uint("user_id", userID))
	c.JSON(http.StatusAccepted, gin.H{"message": "Order accepted for processing"})
}

func (oc *OrderControllerGin) GetOrders(c *gin.Context) {
	userID, ok := GetUserIDFromGinContext(c)
	if !ok {
		oc.logger.Warn("User not authenticated")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	orders, err := oc.orderService.GetUserOrders(userID)
	if err != nil {
		oc.logger.Error("Failed to get orders", zap.Error(err), zap.Uint("user_id", userID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get orders"})
		return
	}

	if len(orders) == 0 {
		oc.logger.Info("No orders found", zap.Uint("user_id", userID))
		c.Status(http.StatusNoContent)
		return
	}

	oc.logger.Info("Orders retrieved successfully", zap.Int("count", len(orders)), zap.Uint("user_id", userID))
	c.JSON(http.StatusOK, orders)
}
