package main

import (
	"context"
	"log"

	"github.com/SZabrodskii/gophermart-stas/internal/config"
	"github.com/SZabrodskii/gophermart-stas/internal/database"
	"github.com/SZabrodskii/gophermart-stas/pkg/logger"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

func main() {
	app := fx.New(
		logger.Module,
		fx.Provide(
			config.New,
			database.New,
		),
		fx.Invoke(StartServer),
	)

	ctx := context.Background()
	if err := app.Start(ctx); err != nil {
		log.Fatalf("Failed to start application: %v", err)
	}

	log.Println("Gophermart loyalty system started successfully!")

	<-app.Done()
}

func StartServer(cfg *config.Config, db *database.DB, logger *zap.Logger) {
	logger.Info("Server starting",
		zap.String("address", cfg.RunAddress),
		zap.String("accrual_address", cfg.AccrualAddress))
	logger.Info("Database connection established")
}
