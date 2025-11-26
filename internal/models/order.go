package models

import (
	"time"

	"gorm.io/gorm"
)

const (
	OrderStatusNew        = "NEW"
	OrderStatusProcessing = "PROCESSING"
	OrderStatusInvalid    = "INVALID"
	OrderStatusProcessed  = "PROCESSED"
)

type Order struct {
	Number     string         `json:"number" gorm:"primaryKey"`
	UserID     uint           `json:"-" gorm:"not null;index"`
	User       User           `json:"-" gorm:"foreignKey:UserID"`
	Status     string         `json:"status" gorm:"default:NEW"`
	Accrual    *float64       `json:"accrual,omitempty"`
	UploadedAt time.Time      `json:"uploaded_at"`
	CreatedAt  time.Time      `json:"-"`
	UpdatedAt  time.Time      `json:"-"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"`
}

type AccrualResponse struct {
	Order   string   `json:"order"`
	Status  string   `json:"status"`
	Accrual *float64 `json:"accrual,omitempty"`
}
