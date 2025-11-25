package server

import (
	"net/http"
	"time"

	"github.com/SZabrodskii/gophermart-stas/internal/config"
	"github.com/SZabrodskii/gophermart-stas/internal/database"
	"github.com/SZabrodskii/gophermart-stas/internal/handlers"
	"github.com/SZabrodskii/gophermart-stas/internal/middleware"

	"github.com/gorilla/mux"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Server struct {
	config  *config.Config
	db      *database.DB
	logger  *zap.Logger
	router  *mux.Router
	handler *handlers.Handler
}

func New(cfg *config.Config, db *database.DB, logger *zap.Logger, handler *handlers.Handler) *Server {
	return &Server{
		config:  cfg,
		db:      db,
		logger:  logger,
		handler: handler,
	}
}

func (s *Server) setupRoutes() {
	s.router = mux.NewRouter()

	s.router.Use(middleware.ZapRequestLogger(s.logger))
	s.router.Use(middleware.CompressAccepted)
	s.router.Use(middleware.Decompress)

	api := s.router.PathPrefix("/api").Subrouter()

	userRoutes := api.PathPrefix("/user").Subrouter()
	userRoutes.HandleFunc("/register", s.handler.Register).Methods("POST")
	userRoutes.HandleFunc("/login", s.handler.Login).Methods("POST")

	authRoutes := userRoutes.NewRoute().Subrouter()
	authRoutes.Use(middleware.JWTAuth(s.logger))

	authRoutes.HandleFunc("/orders", s.handler.UploadOrder).Methods("POST")
	authRoutes.HandleFunc("/orders", s.handler.GetOrders).Methods("GET")
	authRoutes.HandleFunc("/balance", s.handler.GetBalance).Methods("GET")
	authRoutes.HandleFunc("/balance/withdraw", s.handler.WithdrawBalance).Methods("POST")
	authRoutes.HandleFunc("/withdrawals", s.handler.GetWithdrawals).Methods("GET")
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

	s.logger.Info("Starting HTTP server", zap.String("address", s.config.RunAddress))
	return srv.ListenAndServe()
}

var Module = fx.Module("server",
	fx.Provide(New),
)
