package server

import (
	"context"
	"net/http"
	"time"

	"github.com/SZabrodskii/gophermart-stas/internal/config"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type GinEngine struct {
	*gin.Engine
	server *http.Server
	logger *zap.Logger
}

type AuthControllerI interface {
	Register(c *gin.Context)
	Login(c *gin.Context)
	JWTMiddleware() gin.HandlerFunc
}

type OrderControllerI interface {
	UploadOrder(c *gin.Context)
	GetOrders(c *gin.Context)
}

type BalanceControllerI interface {
	GetBalance(c *gin.Context)
	Withdraw(c *gin.Context)
	GetWithdrawals(c *gin.Context)
}

type NewGinEngineParams struct {
	fx.In

	Config *config.Config
	Logger *zap.Logger

	AuthController    AuthControllerI
	OrderController   OrderControllerI
	BalanceController BalanceControllerI
}

func NewGinEngine(params NewGinEngineParams) *GinEngine {
	gin.SetMode(gin.ReleaseMode)

	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.Use(gzip.Gzip(gzip.DefaultCompression))

	ginEngine := &GinEngine{
		Engine: engine,
		logger: params.Logger,
	}

	ginEngine.setupRoutes(params)

	ginEngine.server = &http.Server{
		Addr:         params.Config.RunAddress,
		Handler:      engine,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	return ginEngine
}

func (ge *GinEngine) setupRoutes(params NewGinEngineParams) {
	api := ge.Group("/api")
	user := api.Group("/user")

	user.POST("/register", params.AuthController.Register)
	user.POST("/login", params.AuthController.Login)

	protected := user.Group("")
	protected.Use(params.AuthController.JWTMiddleware())

	protected.POST("/orders", params.OrderController.UploadOrder)
	protected.GET("/orders", params.OrderController.GetOrders)

	protected.GET("/balance", params.BalanceController.GetBalance)
	protected.POST("/balance/withdraw", params.BalanceController.Withdraw)
	protected.GET("/withdrawals", params.BalanceController.GetWithdrawals)
}

func (ge *GinEngine) Start(ctx context.Context) error {
	ge.logger.Info("Starting HTTP server", zap.String("address", ge.server.Addr))

	go func() {
		<-ctx.Done()
		ge.logger.Info("Shutting down HTTP server")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := ge.server.Shutdown(shutdownCtx); err != nil {
			ge.logger.Error("Server forced to shutdown", zap.Error(err))
		}
	}()

	if err := ge.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}
