package main

import (
	"context"
	"log"

	"github.com/SZabrodskii/gophermart-stas/internal/config"
	"github.com/SZabrodskii/gophermart-stas/internal/database"
	"github.com/SZabrodskii/gophermart-stas/internal/handlers"
	"github.com/SZabrodskii/gophermart-stas/internal/server"
	"github.com/SZabrodskii/gophermart-stas/pkg/logger"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

func main() {
	app := fx.New(
		logger.Module,
		server.Module,
		fx.Provide(
			config.New,
			database.New,
			handlers.New,
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

func StartServer(srv *server.Server, logger *zap.Logger) {
	go func() {
		if err := srv.Start(); err != nil {
			logger.Fatal("HTTP server failed", zap.Error(err))
		}
	}()
}
