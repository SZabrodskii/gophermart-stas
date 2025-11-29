package server

import (
	"net/http"
	"time"

	"github.com/SZabrodskii/gophermart-stas/internal/config"
	"github.com/SZabrodskii/gophermart-stas/internal/database"
	"github.com/SZabrodskii/gophermart-stas/internal/handlers"
	"github.com/SZabrodskii/gophermart-stas/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/gopybara/httpbara"
	"go.uber.org/fx"
)

type Server struct {
	config  *config.Config
	db      *database.DB
	logger  httpbara.Logger
	router  *gin.Engine
	handler *handlers.Handler
}

func New(cfg *config.Config, db *database.DB, logger httpbara.Logger, handler *handlers.Handler) *Server {
	gin.SetMode(gin.ReleaseMode)

	return &Server{
		config:  cfg,
		db:      db,
		logger:  logger,
		handler: handler,
	}
}

func (s *Server) setupRoutes() {
	s.router = gin.New()

	s.router.Use(middleware.GinZapLogger(s.logger))
	s.router.Use(middleware.GinGzipMiddleware())
	s.router.Use(gin.Recovery())

	api := s.router.Group("/api")

	userGroup := api.Group("/user")
	{
		userGroup.POST("/register", s.handler.Register)
		userGroup.POST("/login", s.handler.Login)

		authGroup := userGroup.Use(middleware.GinJWTAuth(s.logger))
		{
			authGroup.POST("/orders", s.handler.UploadOrder)
			authGroup.GET("/orders", s.handler.GetOrders)
			authGroup.GET("/balance", s.handler.GetBalance)
			authGroup.POST("/balance/withdraw", s.handler.WithdrawBalance)
			authGroup.GET("/withdrawals", s.handler.GetWithdrawals)
		}
	}
}

func (s *Server) Start() error {
	s.setupRoutes()

	srv := &http.Server{
		Addr:         s.config.RunAddress,
		Handler:      s.router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	s.logger.Info("Starting HTTP server", "address", s.config.RunAddress)
	return srv.ListenAndServe()
}

var Module = fx.Module("server",
	fx.Provide(New),
)
