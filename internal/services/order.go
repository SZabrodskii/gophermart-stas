package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/SZabrodskii/gophermart-stas/internal/database"
	"github.com/SZabrodskii/gophermart-stas/internal/models"
	"github.com/SZabrodskii/gophermart-stas/internal/utils"

	"github.com/gopybara/httpbara"
	"gorm.io/gorm"
)

var (
	ErrInvalidOrderNumber   = errors.New("invalid order number")
	ErrOrderExists          = errors.New("order already exists")
	ErrOrderExistsOtherUser = errors.New("order exists for another user")
)

type OrderService struct {
	db     *database.DB
	logger httpbara.Logger
}

func NewOrderService(db *database.DB, logger httpbara.Logger) *OrderService {
	return &OrderService{
		db:     db,
		logger: logger,
	}
}

func (s *OrderService) UploadOrder(userID uint, orderNumber string) error {
	if !utils.ValidateOrderNumber(orderNumber) {
		s.logger.Warn("Invalid order number format", "number", orderNumber)
		return ErrInvalidOrderNumber
	}

	var existingOrder models.Order
	err := s.db.GetDB().Where("number = ?", orderNumber).First(&existingOrder).Error
	if err == nil {
		if existingOrder.UserID == userID {
			s.logger.Info("Order already exists for user", "number", orderNumber, "userID", userID)
			return ErrOrderExists
		} else {
			s.logger.Warn("Order exists for another user", "number", orderNumber, "existingUserID", existingOrder.UserID, "requestUserID", userID)
			return ErrOrderExistsOtherUser
		}
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		s.logger.Error("Database error during order check", "error", err)
		return fmt.Errorf("database error: %w", err)
	}

	order := models.Order{
		Number:     orderNumber,
		UserID:     userID,
		Status:     models.OrderStatusNew,
		UploadedAt: time.Now(),
	}

	if err := s.db.GetDB().Create(&order).Error; err != nil {
		s.logger.Error("Failed to create order", "error", err)
		return fmt.Errorf("failed to create order: %w", err)
	}

	s.logger.Info("Order created successfully", "number", orderNumber, "userID", userID)
	return nil
}

func (s *OrderService) GetUserOrders(userID uint) ([]models.Order, error) {
	var orders []models.Order
	err := s.db.GetDB().Where("user_id = ?", userID).Order("uploaded_at DESC").Find(&orders).Error
	if err != nil {
		s.logger.Error("Failed to get user orders", "error", err, "userID", userID)
		return nil, fmt.Errorf("failed to get orders: %w", err)
	}

	s.logger.Debug("Retrieved user orders", "userID", userID, "count", len(orders))
	return orders, nil
}
