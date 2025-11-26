package middleware

import (
	"net/http"
	"strings"

	"github.com/SZabrodskii/gophermart-stas/internal/auth"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func GinZapLogger(logger *zap.Logger) gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(params gin.LogFormatterParams) string {
		logger.Info("http request",
			zap.String("method", params.Method),
			zap.String("path", params.Path),
			zap.Int("status", params.StatusCode),
			zap.Duration("latency", params.Latency),
			zap.String("client_ip", params.ClientIP),
			zap.String("user_agent", params.Request.UserAgent()),
		)
		return ""
	})
}

func GinGzipMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}

func GinJWTAuth(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			logger.Warn("Missing Authorization header")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization required"})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			logger.Warn("Invalid Authorization header format")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token format"})
			return
		}

		claims, err := auth.ParseJWT(tokenString)
		if err != nil {
			logger.Warn("Invalid JWT token", zap.Error(err))
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		c.Set("userID", claims.UserID)
		c.Set("login", claims.Login)
		c.Next()
	}
}