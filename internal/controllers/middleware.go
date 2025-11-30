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

type middlewareControllerDescription struct {
	JWTMiddleware httpbara.Middleware `middleware:"jwt"`
}

type newMiddlewareControllerIn struct {
	fx.In

	Logger httpbara.Logger
}

type middlewareController struct {
	middlewareControllerDescription

	logger httpbara.Logger
}

type UserIDKey string

const UserIDContextKey UserIDKey = "userID"

func NewMiddlewareController(in newMiddlewareControllerIn) (server.AsHandlerOut, error) {
	return server.AsHandler(&middlewareController{
		logger: in.Logger,
	})
}

func (mc *middlewareController) JWTMiddleware(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		mc.logger.Warn("Missing Authorization header")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
		c.Abort()
		return
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		mc.logger.Warn("Invalid Authorization header format")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format"})
		c.Abort()
		return
	}

	token := parts[1]
	claims, err := auth.ParseJWT(token)
	if err != nil {
		mc.logger.Warn("Invalid or expired token", "error", err)
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
