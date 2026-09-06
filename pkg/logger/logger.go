package logger

import (
	"log/slog"
	"os"
	"strings"
)

// Init initializes and sets the default global slog logger.
func Init(levelStr string, isProduction bool) *slog.Logger {
	var level slog.Level
	switch strings.ToLower(levelStr) {
	case "debug":
		level = slog.LevelDebug
	case "warn", "warning":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: !isProduction, // include source file in non-prod for easier debugging
	}

	var handler slog.Handler
	if isProduction {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		// In development, text handler or json handler with colored/clean formatting
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)

	return logger
}
