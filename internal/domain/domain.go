package domain

import "github.com/SZabrodskii/gophermart-stas/internal/models"

type UserServiceI interface {
	Register(login, password string) (string, error)
	Login(login, password string) (string, error)
}

type OrderServiceI interface {
	UploadOrder(userID uint, orderNumber string) error
	GetUserOrders(userID uint) ([]models.Order, error)
}

type BalanceServiceI interface {
	GetBalance(userID uint) (*models.Balance, error)
	WithdrawBalance(userID uint, orderNumber string, amount float64) error
	GetWithdrawals(userID uint) ([]models.Withdrawal, error)
}
