package controllers

import (
	"context"
	"net/http"
	"strings"

	"github.com/SZabrodskii/gophermart-stas/internal/auth"
	"github.com/SZabrodskii/gophermart-stas/internal/server"
	"github.com/gin-gonic/gin"
	"github.com/gopybara/httpbara"
	"go.uber.org/fx"
)

type jwtMiddlewareDescription struct {
	JWTMiddleware httpbara.Middleware `middleware:"jwt"`
}

type newJWTMiddlewareIn struct {
	fx.In

	Logger httpbara.Logger
}

type jwtMiddleware struct {
	jwtMiddlewareDescription

	logger httpbara.Logger
}

type UserIDKey string

const UserIDContextKey UserIDKey = "userID"

func NewJWTMiddleware(in newJWTMiddlewareIn) (server.AsHandlerOut, error) {
	return server.AsHandler(&jwtMiddleware{
		logger: in.Logger,
	})
}

func (jm *jwtMiddleware) JWTMiddleware(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		jm.logger.Warn("Missing Authorization header")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
		c.Abort()
		return
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		jm.logger.Warn("Invalid Authorization header format")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format"})
		c.Abort()
		return
	}

	token := parts[1]
	claims, err := auth.ParseJWT(token)
	if err != nil {
		jm.logger.Warn("Invalid or expired token", "error", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
		c.Abort()
		return
	}

	c.Set("userID", claims.UserID)

	ctx := context.WithValue(c.Request.Context(), UserIDContextKey, claims.UserID)
	c.Request = c.Request.WithContext(ctx)

	c.Next()
}

func GetUserIDFromContext(ctx context.Context) (uint, bool) {
	userID, ok := ctx.Value(UserIDContextKey).(uint)
	return userID, ok
}
