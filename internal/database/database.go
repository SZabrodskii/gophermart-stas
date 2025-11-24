package database

import (
	"database/sql"
	"fmt"

	"github.com/SZabrodskii/gophermart-stas/internal/config"

	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

type DB struct {
	conn   *sql.DB
	logger *zap.Logger
}

func New(cfg *config.Config, logger *zap.Logger) (*DB, error) {
	logger.Info("Connecting to database", zap.String("uri", maskPassword(cfg.DatabaseURI)))

	conn, err := sql.Open("postgres", cfg.DatabaseURI)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := conn.Ping(); err != nil {
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

	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			login VARCHAR(255) UNIQUE NOT NULL,
			password VARCHAR(255) NOT NULL,
			created_at TIMESTAMP DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS orders (
			number VARCHAR(255) PRIMARY KEY,
			user_id INTEGER REFERENCES users(id),
			status VARCHAR(20) NOT NULL DEFAULT 'NEW',
			accrual DECIMAL(10,2),
			uploaded_at TIMESTAMP DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS balances (
			user_id INTEGER PRIMARY KEY REFERENCES users(id),
			current DECIMAL(10,2) DEFAULT 0,
			withdrawn DECIMAL(10,2) DEFAULT 0
		)`,
		`CREATE TABLE IF NOT EXISTS withdrawals (
			id SERIAL PRIMARY KEY,
			user_id INTEGER REFERENCES users(id),
			order_number VARCHAR(255) NOT NULL,
			sum DECIMAL(10,2) NOT NULL,
			processed_at TIMESTAMP DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_withdrawals_user_id ON withdrawals(user_id)`,
	}

	for i, query := range queries {
		db.logger.Debug("Executing migration query", zap.Int("step", i+1))
		if _, err := db.conn.Exec(query); err != nil {
			db.logger.Error("Migration failed",
				zap.Int("step", i+1),
				zap.Error(err))
			return fmt.Errorf("failed to execute query %d: %w", i+1, err)
		}
	}

	db.logger.Info("Database migration completed", zap.Int("queries", len(queries)))
	return nil
}

func (db *DB) Close() error {
	db.logger.Info("Closing database connection")
	return db.conn.Close()
}

func (db *DB) GetDB() *sql.DB {
	return db.conn
}
