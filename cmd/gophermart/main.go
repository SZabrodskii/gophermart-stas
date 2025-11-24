package main

import (
	"context"
	"log"

	"github.com/SZabrodskii/gophermart-stas/internal/config"
	"github.com/SZabrodskii/gophermart-stas/internal/database"

	"go.uber.org/fx"
)

func main() {
	app := fx.New(
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

func StartServer(cfg *config.Config, db *database.DB) {
	log.Printf("Server starting on %s", cfg.RunAddress)
	log.Printf("Database connection established")
	log.Printf("Accrual system address: %s", cfg.AccrualAddress)
}
