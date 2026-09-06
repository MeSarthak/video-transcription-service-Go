package database

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"video-transcription-service/internal/config"
)

type Database struct {
	Pool *pgxpool.Pool
}

// New initializes and returns a new PostgreSQL database connection pool.
func New(ctx context.Context, cfg *config.Config) (*Database, error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("unable to parse database config: %w", err)
	}

	poolConfig.MaxConns = int32(cfg.DBMaxOpenConns)
	poolConfig.MinConns = int32(cfg.DBMaxIdleConns)
	poolConfig.MaxConnIdleTime = cfg.DBMaxIdleTime
	poolConfig.MaxConnLifetime = 1 * time.Hour

	// Connect with timeout
	connectCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(connectCtx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create pgx pool: %w", err)
	}

	// Verify connection
	if err := pool.Ping(connectCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	slog.Info("PostgreSQL database pool initialized successfully",
		slog.Int("max_conns", cfg.DBMaxOpenConns),
		slog.Int("min_conns", cfg.DBMaxIdleConns),
	)

	return &Database{Pool: pool}, nil
}

// Ping verifies the database connection is alive.
func (db *Database) Ping(ctx context.Context) error {
	pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	return db.Pool.Ping(pingCtx)
}

// Close closes all connections in the database pool.
func (db *Database) Close() {
	if db.Pool != nil {
		slog.Info("Closing PostgreSQL connection pool")
		db.Pool.Close()
	}
}

// WithTx executes the given function within a database transaction.
func (db *Database) WithTx(ctx context.Context, fn func(tx pgx.Tx) error) error {
	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("error in tx: %v, rollback error: %w", err, rbErr)
		}
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
