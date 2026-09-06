package database

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"

	"video-transcription-service/migrations"
)

// RunMigrations executes all pending database migrations against the target database URL.
func RunMigrations(databaseURL string) error {
	d, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return fmt.Errorf("failed to create iofs driver for embedded migrations: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", d, databaseURL)
	if err != nil {
		return fmt.Errorf("failed to initialize migrator: %w", err)
	}
	defer func() {
		srcErr, dbErr := m.Close()
		if srcErr != nil {
			slog.Debug("Error closing migration source", slog.Any("error", srcErr))
		}
		if dbErr != nil {
			slog.Debug("Error closing migration db connection", slog.Any("error", dbErr))
		}
	}()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	version, dirty, err := m.Version()
	if err != nil && !errors.Is(err, migrate.ErrNilVersion) {
		slog.Warn("Could not fetch migration version", slog.Any("error", err))
	} else {
		slog.Info("Database migrations applied successfully",
			slog.Uint64("version", uint64(version)),
			slog.Bool("dirty", dirty),
		)
	}

	return nil
}
