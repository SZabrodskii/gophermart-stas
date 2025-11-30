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
	ErrInsufficientFunds = errors.New("insufficient funds")
)

type BalanceService struct {
	db     *database.DB
	logger *zap.Logger
}

func NewBalanceService(db *database.DB, logger *zap.Logger) *BalanceService {
	return &BalanceService{
		db:     db,
		logger: logger,
	}
}

func (s *BalanceService) GetBalance(userID uint) (*models.Balance, error) {
	var balance models.Balance
	err := s.db.GetDB().Where("user_id = ?", userID).First(&balance).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			balance = models.Balance{
				UserID:    userID,
				Current:   0,
				Withdrawn: 0,
			}
			if createErr := s.db.GetDB().Create(&balance).Error; createErr != nil {
				s.logger.Error("Failed to create balance", zap.Error(createErr))
				return nil, fmt.Errorf("failed to create balance: %w", createErr)
			}
		} else {
			s.logger.Error("Failed to get balance", zap.Error(err), zap.Uint("userID", userID))
			return nil, fmt.Errorf("failed to get balance: %w", err)
		}
	}

	return &balance, nil
}

func (s *BalanceService) WithdrawBalance(userID uint, orderNumber string, amount float64) error {
	if !utils.ValidateOrderNumber(orderNumber) {
		s.logger.Warn("Invalid order number format for withdrawal", zap.String("number", orderNumber))
		return ErrInvalidOrderNumber
	}

	if amount <= 0 {
		return errors.New("withdrawal amount must be positive")
	}

	return s.db.GetDB().Transaction(func(tx *gorm.DB) error {
		var balance models.Balance
		if err := tx.Where("user_id = ?", userID).First(&balance).Error; err != nil {
			s.logger.Error("Failed to get balance for withdrawal", zap.Error(err))
			return fmt.Errorf("failed to get balance: %w", err)
		}

		if balance.Current < amount {
			s.logger.Warn("Insufficient funds for withdrawal",
				zap.Uint("userID", userID),
				zap.Float64("current", balance.Current),
				zap.Float64("requested", amount))
			return ErrInsufficientFunds
		}

		balance.Current -= amount
		balance.Withdrawn += amount

		if err := tx.Save(&balance).Error; err != nil {
			s.logger.Error("Failed to update balance", zap.Error(err))
			return fmt.Errorf("failed to update balance: %w", err)
		}

		withdrawal := models.Withdrawal{
			UserID:      userID,
			Order:       orderNumber,
			Sum:         amount,
			ProcessedAt: time.Now(),
		}

		if err := tx.Create(&withdrawal).Error; err != nil {
			s.logger.Error("Failed to create withdrawal record", zap.Error(err))
			return fmt.Errorf("failed to create withdrawal: %w", err)
		}

		s.logger.Info("Withdrawal processed successfully",
			zap.Uint("userID", userID),
			zap.String("orderNumber", orderNumber),
			zap.Float64("amount", amount))

		return nil
	})
}

func (s *BalanceService) GetWithdrawals(userID uint) ([]models.Withdrawal, error) {
	var withdrawals []models.Withdrawal
	err := s.db.GetDB().Where("user_id = ?", userID).Order("processed_at DESC").Find(&withdrawals).Error
	if err != nil {
		s.logger.Error("Failed to get withdrawals", zap.Error(err), zap.Uint("userID", userID))
		return nil, fmt.Errorf("failed to get withdrawals: %w", err)
	}

	s.logger.Debug("Retrieved user withdrawals", zap.Uint("userID", userID), zap.Int("count", len(withdrawals)))
	return withdrawals, nil
}

func (s *BalanceService) AddBalance(userID uint, amount float64) error {
	return s.db.GetDB().Transaction(func(tx *gorm.DB) error {
		var balance models.Balance
		err := tx.Where("user_id = ?", userID).First(&balance).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				balance = models.Balance{
					UserID:  userID,
					Current: amount,
				}
				err = tx.Create(&balance).Error
				if err != nil {
					s.logger.Error("Failed to create balance", zap.Error(err), zap.Uint("userID", userID))
					return fmt.Errorf("failed to create balance: %w", err)
				}
			} else {
				s.logger.Error("Failed to get balance", zap.Error(err), zap.Uint("userID", userID))
				return fmt.Errorf("failed to get balance: %w", err)
			}
		} else {
			balance.Current += amount
			err = tx.Save(&balance).Error
			if err != nil {
				s.logger.Error("Failed to update balance", zap.Error(err), zap.Uint("userID", userID))
				return fmt.Errorf("failed to update balance: %w", err)
			}
		}

		s.logger.Info("Balance added successfully", zap.Uint("userID", userID), zap.Float64("amount", amount), zap.Float64("newBalance", balance.Current))
		return nil
	})
}
