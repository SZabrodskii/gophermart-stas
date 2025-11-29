package database

import (
	"context"
	"fmt"
	"time"

	"github.com/SZabrodskii/gophermart-stas/internal/config"
	"github.com/SZabrodskii/gophermart-stas/internal/models"

	"github.com/gopybara/httpbara"
	"go.uber.org/fx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DB struct {
	conn   *gorm.DB
	logger httpbara.Logger
}

func New(lc fx.Lifecycle, cfg *config.Config, logger httpbara.Logger) (*DB, error) {
	logger.Info("Connecting to database", "uri", maskPassword(cfg.DatabaseURI))

	conn, err := gorm.Open(postgres.Open(cfg.DatabaseURI), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	sqlDB, err := conn.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(time.Minute * 30)

	db := &DB{
		conn:   conn,
		logger: logger,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
			defer cancel()
			if err := sqlDB.PingContext(ctx); err != nil {
				return fmt.Errorf("failed to ping database: %w", err)
			}

			if err := db.migrate(); err != nil {
				return fmt.Errorf("failed to migrate database: %w", err)
			}

			logger.Info("Database connected and migrated successfully")
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("Closing database connection")
			return sqlDB.Close()
		},
	})

	return db, nil
}

func maskPassword(uri string) string {
	return "***masked***"
}

func (db *DB) migrate() error {
	db.logger.Info("Starting database migration")

	err := db.conn.AutoMigrate(
		&models.User{},
		&models.Order{},
		&models.Balance{},
		&models.Withdrawal{},
	)
	if err != nil {
		db.logger.Error("Migration failed", "error", err)
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	db.logger.Info("GORM AutoMigrate completed successfully")
	return nil
}

func (db *DB) GetDB() *gorm.DB {
	return db.conn
}
