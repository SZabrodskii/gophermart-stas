package controllers

import (
	"errors"
	"net/http"

	"github.com/SZabrodskii/gophermart-stas/internal/domain"
	"github.com/SZabrodskii/gophermart-stas/internal/models"
	"github.com/SZabrodskii/gophermart-stas/internal/server"
	"github.com/SZabrodskii/gophermart-stas/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/gopybara/httpbara"
	"github.com/gopybara/httpbara/casual"
	"go.uber.org/fx"
)

type authControllerDescription struct {
	API      httpbara.Group `group:"/api/user"`
	Register httpbara.Route `route:"POST /register" group:"api" middlewares:""`
	Login    httpbara.Route `route:"POST /login" group:"api" middlewares:""`
}

type newAuthControllerIn struct {
	fx.In

	Logger      httpbara.Logger
	UserService domain.UserServiceI
}

type authController struct {
	authControllerDescription

	logger      httpbara.Logger
	userService domain.UserServiceI
}

type AuthResponse struct{}

func (ar *AuthResponse) StatusCode() int {
	return http.StatusOK
}

func NewAuthController(in newAuthControllerIn) (server.AsHandlerOut, error) {
	return server.AsHandler(&authController{
		logger:      in.Logger,
		userService: in.UserService,
	})
}

func (ac *authController) Register(ctx *gin.Context, req *models.UserRequest) (*AuthResponse, error) {
	if req.Login == "" || req.Password == "" {
		ac.logger.Warn("Missing login or password")
		return nil, casual.NewHTTPErrorFromMessage(400, "Login and password are required")
	}

	token, err := ac.userService.Register(req.Login, req.Password)
	if err != nil {
		if errors.Is(err, services.ErrUserExists) {
			ac.logger.Warn("User already exists", "login", req.Login)
			return nil, casual.NewHTTPErrorFromMessage(409, "User already exists")
		}
		ac.logger.Error("Failed to register user", "error", err, "login", req.Login)
		return nil, casual.NewHTTPErrorFromMessage(500, "Registration failed")
	}

	ctx.Header("Authorization", "Bearer "+token)
	ac.logger.Info("User registered successfully", "login", req.Login)
	return &AuthResponse{}, nil
}

func (ac *authController) Login(ctx *gin.Context, req *models.UserRequest) (*AuthResponse, error) {
	if req.Login == "" || req.Password == "" {
		ac.logger.Warn("Missing login or password")
		return nil, casual.NewHTTPErrorFromMessage(400, "Login and password are required")
	}

	token, err := ac.userService.Login(req.Login, req.Password)
	if err != nil {
		if errors.Is(err, services.ErrInvalidCredentials) {
			ac.logger.Warn("Invalid credentials", "login", req.Login)
			return nil, casual.NewHTTPErrorFromMessage(401, "Invalid credentials")
		}
		ac.logger.Error("Failed to login user", "error", err, "login", req.Login)
		return nil, casual.NewHTTPErrorFromMessage(500, "Login failed")
	}

	ctx.Header("Authorization", "Bearer "+token)
	ac.logger.Info("User logged in successfully", "login", req.Login)
	return &AuthResponse{}, nil
}
