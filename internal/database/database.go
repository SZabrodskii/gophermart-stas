package database

import (
	"fmt"

	"github.com/SZabrodskii/gophermart-stas/internal/config"
	"github.com/SZabrodskii/gophermart-stas/internal/models"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DB struct {
	conn   *gorm.DB
	logger *zap.Logger
}

func New(cfg *config.Config, logger *zap.Logger) (*DB, error) {
	logger.Info("Connecting to database", zap.String("uri", maskPassword(cfg.DatabaseURI)))

	conn, err := gorm.Open(postgres.Open(cfg.DatabaseURI), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	sqlDB, err := conn.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	db := &DB{
		conn:   conn,
		logger: logger,
	}

	if err := db.migrate(); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	logger.Info("Database migration completed successfully")
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
		db.logger.Error("Migration failed", zap.Error(err))
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	db.logger.Info("GORM AutoMigrate completed successfully")
	return nil
}

func (db *DB) Close() error {
	db.logger.Info("Closing database connection")
	sqlDB, err := db.conn.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func (db *DB) GetDB() *gorm.DB {
	return db.conn
}
