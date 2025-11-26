package models

import (
	"time"

	"gorm.io/gorm"
)

type Balance struct {
	UserID    uint           `json:"-" gorm:"primaryKey"`
	User      User           `json:"-" gorm:"foreignKey:UserID"`
	Current   float64        `json:"current" gorm:"default:0"`
	Withdrawn float64        `json:"withdrawn" gorm:"default:0"`
	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}
