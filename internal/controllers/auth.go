package controllers

import (
	"net/http"
	"strings"

	"github.com/SZabrodskii/gophermart-stas/internal/auth"
	"github.com/SZabrodskii/gophermart-stas/internal/domain"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AuthControllerGin struct {
	logger      *zap.Logger
	userService domain.UserServiceI
}

func NewAuthControllerGin(logger *zap.Logger, userService domain.UserServiceI) *AuthControllerGin {
	return &AuthControllerGin{
		logger:      logger,
		userService: userService,
	}
}

type AuthRequest struct {
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (ac *AuthControllerGin) Register(c *gin.Context) {
	var req AuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ac.logger.Warn("Invalid request format", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	token, err := ac.userService.Register(req.Login, req.Password)
	if err != nil {
		switch err.Error() {
		case "user already exists":
			ac.logger.Warn("User already exists", zap.String("login", req.Login))
			c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
		default:
			ac.logger.Error("Failed to register user", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	ac.logger.Info("User registered successfully", zap.String("login", req.Login))
	c.Header("Authorization", "Bearer "+token)
	c.JSON(http.StatusOK, gin.H{"message": "User registered successfully"})
}

func (ac *AuthControllerGin) Login(c *gin.Context) {
	var req AuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ac.logger.Warn("Invalid request format", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	token, err := ac.userService.Login(req.Login, req.Password)
	if err != nil {
		ac.logger.Warn("Invalid credentials", zap.String("login", req.Login))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	ac.logger.Info("User logged in successfully", zap.String("login", req.Login))
	c.Header("Authorization", "Bearer "+token)
	c.JSON(http.StatusOK, gin.H{"message": "User logged in successfully"})
}

func (ac *AuthControllerGin) JWTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			ac.logger.Warn("Missing Authorization header")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization required"})
			c.Abort()
			return
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			ac.logger.Warn("Invalid Authorization header format")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := auth.ParseJWT(tokenString)
		if err != nil {
			ac.logger.Warn("Invalid JWT token", zap.Error(err))
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("login", claims.Login)
		c.Next()
	}
}

func GetUserIDFromGinContext(c *gin.Context) (uint, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}

	id, ok := userID.(uint)
	return id, ok
}
