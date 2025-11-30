package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/SZabrodskii/gophermart-stas/internal/database"
	"github.com/SZabrodskii/gophermart-stas/internal/models"
	"github.com/SZabrodskii/gophermart-stas/internal/utils"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	ErrInvalidOrderNumber   = errors.New("invalid order number")
	ErrOrderExists          = errors.New("order already exists")
	ErrOrderExistsOtherUser = errors.New("order exists for another user")
)

type OrderService struct {
	db     *database.DB
	logger *zap.Logger
}

func NewOrderService(db *database.DB, logger *zap.Logger) *OrderService {
	return &OrderService{
		db:     db,
		logger: logger,
	}
}

func (s *OrderService) UploadOrder(userID uint, orderNumber string) error {
	if !utils.ValidateOrderNumber(orderNumber) {
		s.logger.Warn("Invalid order number format", zap.String("number", orderNumber))
		return ErrInvalidOrderNumber
	}

	var existingOrder models.Order
	err := s.db.GetDB().Where("number = ?", orderNumber).First(&existingOrder).Error
	if err == nil {
		if existingOrder.UserID == userID {
			s.logger.Info("Order already exists for user", zap.String("number", orderNumber), zap.Uint("userID", userID))
			return ErrOrderExists
		} else {
			s.logger.Warn("Order exists for another user", zap.String("number", orderNumber), zap.Uint("existingUserID", existingOrder.UserID), zap.Uint("requestUserID", userID))
			return ErrOrderExistsOtherUser
		}
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		s.logger.Error("Database error during order check", zap.Error(err))
		return fmt.Errorf("database error: %w", err)
	}

	order := models.Order{
		Number:     orderNumber,
		UserID:     userID,
		Status:     models.OrderStatusNew,
		UploadedAt: time.Now(),
	}

	if err := s.db.GetDB().Create(&order).Error; err != nil {
		s.logger.Error("Failed to create order", zap.Error(err))
		return fmt.Errorf("failed to create order: %w", err)
	}

	s.logger.Info("Order created successfully", zap.String("number", orderNumber), zap.Uint("userID", userID))
	return nil
}

func (s *OrderService) GetUserOrders(userID uint) ([]models.Order, error) {
	var orders []models.Order
	err := s.db.GetDB().Where("user_id = ?", userID).Order("uploaded_at DESC").Find(&orders).Error
	if err != nil {
		s.logger.Error("Failed to get user orders", zap.Error(err), zap.Uint("userID", userID))
		return nil, fmt.Errorf("failed to get orders: %w", err)
	}

	s.logger.Debug("Retrieved user orders", zap.Uint("userID", userID), zap.Int("count", len(orders)))
	return orders, nil
}

func (s *OrderService) GetOrdersForProcessing() ([]*models.Order, error) {
	var orders []*models.Order
	err := s.db.GetDB().Where("status IN (?)", []string{models.OrderStatusNew, models.OrderStatusProcessing}).Find(&orders).Error
	if err != nil {
		s.logger.Error("Failed to get orders for processing", zap.Error(err))
		return nil, fmt.Errorf("failed to get orders for processing: %w", err)
	}

	return orders, nil
}

func (s *OrderService) UpdateOrderStatus(orderNumber string, status string, accrual float64) error {
	updates := map[string]interface{}{
		"status": status,
	}

	if accrual > 0 {
		updates["accrual"] = &accrual
	}

	err := s.db.GetDB().Model(&models.Order{}).Where("number = ?", orderNumber).Updates(updates).Error
	if err != nil {
		s.logger.Error("Failed to update order status", zap.Error(err), zap.String("order", orderNumber), zap.String("status", status))
		return fmt.Errorf("failed to update order status: %w", err)
	}

	s.logger.Info("Order status updated", zap.String("order", orderNumber), zap.String("status", status), zap.Float64("accrual", accrual))
	return nil
}
