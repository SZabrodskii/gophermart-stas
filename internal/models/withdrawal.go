package models

import (
	"time"

	"gorm.io/gorm"
)

type Withdrawal struct {
	ID          uint           `json:"-" gorm:"primaryKey"`
	UserID      uint           `json:"-" gorm:"not null;index"`
	User        User           `json:"-" gorm:"foreignKey:UserID"`
	Order       string         `json:"order" gorm:"column:order_number;not null"`
	Sum         float64        `json:"sum" gorm:"not null"`
	ProcessedAt time.Time      `json:"processed_at"`
	CreatedAt   time.Time      `json:"-"`
	UpdatedAt   time.Time      `json:"-"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

type WithdrawalRequest struct {
	Order string  `json:"order" binding:"required"`
	Sum   float64 `json:"sum" binding:"required"`
}
