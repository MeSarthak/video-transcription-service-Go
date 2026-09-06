package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"video-transcription-service/internal/config"
	"video-transcription-service/internal/health"
	"video-transcription-service/internal/middleware"
	"video-transcription-service/pkg/logger"
)

func main() {
	// 1. Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// 2. Initialize structured logger
	logger.Init(cfg.LogLevel, cfg.IsProduction())
	slog.Info("Starting Video Transcription API",
		slog.String("env", cfg.Env),
		slog.String("port", cfg.Port),
		slog.String("log_level", cfg.LogLevel),
	)

	// 3. Set Gin mode
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	// 4. Setup Router & Middleware
	router := gin.New()
	router.Use(middleware.Recovery())
	router.Use(middleware.RequestLogger())
	router.Use(middleware.CORS(cfg.AllowedOrigins))

	// 5. Register Routes
	healthHandler := health.NewHandler()
	healthHandler.RegisterRoutes(router)

	// Base API route placeholder for future modules
	v1 := router.Group("/api/v1")
	{
		v1.GET("/ping", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "pong"})
		})
	}

	// 6. Setup HTTP Server
	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  60 * time.Second,
	}

	// 7. Start server in a background goroutine
	go func() {
		slog.Info("HTTP server listening", slog.String("addr", server.Addr))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("HTTP server failed unexpectedly", slog.Any("error", err))
			os.Exit(1)
		}
	}()

	// 8. Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	sig := <-quit

	slog.Info("Shutdown signal received, shutting down gracefully...", slog.String("signal", sig.String()))

	// Context with 10 second timeout for active requests to complete
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown", slog.Any("error", err))
		os.Exit(1)
	}

	slog.Info("Server stopped cleanly")
}
