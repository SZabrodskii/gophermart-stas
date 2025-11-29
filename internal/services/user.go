package services

import (
	"errors"
	"fmt"

	"github.com/SZabrodskii/gophermart-stas/internal/auth"
	"github.com/SZabrodskii/gophermart-stas/internal/database"
	"github.com/SZabrodskii/gophermart-stas/internal/models"
	"github.com/SZabrodskii/gophermart-stas/internal/utils"

	"github.com/gopybara/httpbara"
	"gorm.io/gorm"
)

var (
	ErrUserExists         = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type UserService struct {
	db     *database.DB
	logger httpbara.Logger
}

func NewUserService(db *database.DB, logger httpbara.Logger) *UserService {
	return &UserService{
		db:     db,
		logger: logger,
	}
}

func (s *UserService) Register(login, password string) (string, error) {
	var existingUser models.User
	err := s.db.GetDB().Where("login = ?", login).First(&existingUser).Error
	if err == nil {
		return "", ErrUserExists
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		s.logger.Error("Database error during user check", "error", err)
		return "", fmt.Errorf("database error: %w", err)
	}

	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		s.logger.Error("Failed to hash password", "error", err)
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	user := models.User{
		Login:    login,
		Password: hashedPassword,
	}

	if err := s.db.GetDB().Create(&user).Error; err != nil {
		s.logger.Error("Failed to create user", "error", err)
		return "", fmt.Errorf("failed to create user: %w", err)
	}

	balance := models.Balance{
		UserID:    user.ID,
		Current:   0,
		Withdrawn: 0,
	}

	if err := s.db.GetDB().Create(&balance).Error; err != nil {
		s.logger.Error("Failed to create user balance", "error", err)
	}

	token, err := auth.GenerateJWT(user.ID, user.Login)
	if err != nil {
		s.logger.Error("Failed to generate JWT", "error", err)
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	s.logger.Info("User registered successfully", "login", login)
	return token, nil
}

func (s *UserService) Login(login, password string) (string, error) {
	var user models.User
	err := s.db.GetDB().Where("login = ?", login).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", ErrInvalidCredentials
		}
		s.logger.Error("Database error during login", "error", err)
		return "", fmt.Errorf("database error: %w", err)
	}

	if !utils.CheckPassword(password, user.Password) {
		return "", ErrInvalidCredentials
	}

	token, err := auth.GenerateJWT(user.ID, user.Login)
	if err != nil {
		s.logger.Error("Failed to generate JWT", "error", err)
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	s.logger.Info("User logged in successfully", "login", login)
	return token, nil
}
