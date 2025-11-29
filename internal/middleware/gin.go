package middleware

import (
	"net/http"
	"strings"

	"github.com/SZabrodskii/gophermart-stas/internal/auth"

	"github.com/gin-gonic/gin"
	"github.com/gopybara/httpbara"
)

func GinZapLogger(logger httpbara.Logger) gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(params gin.LogFormatterParams) string {
		logger.Info("http request",
			"method", params.Method,
			"path", params.Path,
			"status", params.StatusCode,
			"latency", params.Latency,
			"client_ip", params.ClientIP,
			"user_agent", params.Request.UserAgent(),
		)
		return ""
	})
}

func GinGzipMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}

func GinJWTAuth(logger httpbara.Logger) gin.HandlerFunc {
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
			logger.Warn("Invalid JWT token", "error", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		c.Set("userID", claims.UserID)
		c.Set("login", claims.Login)
		c.Next()
	}
}
