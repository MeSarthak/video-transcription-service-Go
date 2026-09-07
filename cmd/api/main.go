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

	"video-transcription-service/internal/auth"
	"video-transcription-service/internal/config"
	"video-transcription-service/internal/database"
	"video-transcription-service/internal/health"
	"video-transcription-service/internal/jobs"
	"video-transcription-service/internal/middleware"
	"video-transcription-service/internal/queue"
	"video-transcription-service/internal/storage"
	"video-transcription-service/internal/transcription"
	"video-transcription-service/internal/users"
	"video-transcription-service/internal/videos"
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

	// 3. Initialize Database & run migrations
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db, err := database.New(ctx, cfg)
	if err != nil {
		if cfg.IsProduction() {
			slog.Error("Database connection failed in production", slog.Any("error", err))
			os.Exit(1)
		} else {
			slog.Warn("Database connection failed (development mode) - continuing with degraded ready probe", slog.Any("error", err))
		}
	} else {
		defer db.Close()

		// Run automated migrations
		if err := database.RunMigrations(cfg.DatabaseURL); err != nil {
			slog.Error("Failed to apply database migrations", slog.Any("error", err))
			if cfg.IsProduction() {
				os.Exit(1)
			}
		}
	}

	// 4. Initialize S3 Storage Service
	var storageService storage.Service
	s3Store, err := storage.NewS3Storage(ctx, cfg.AWSRegion, cfg.S3BucketName)
	if err != nil {
		if cfg.IsProduction() {
			slog.Error("AWS S3 initialization failed in production", slog.Any("error", err))
			os.Exit(1)
		} else {
			slog.Warn("AWS S3 initialization warning (using mock storage for local dev)", slog.Any("error", err))
			storageService = storage.NewMockStorage()
		}
	} else {
		storageService = s3Store
	}

	// 5. Initialize SQS Queue Service
	var jobQueue queue.Queue
	sqsQueue, err := queue.NewSQSQueue(ctx, cfg.AWSRegion, cfg.SQSQueueURL)
	if err != nil {
		if cfg.IsProduction() {
			slog.Error("AWS SQS initialization failed in production", slog.Any("error", err))
			os.Exit(1)
		} else {
			slog.Warn("AWS SQS initialization warning (using mock queue for local dev)", slog.Any("error", err))
			jobQueue = queue.NewMockQueue()
		}
	} else {
		jobQueue = sqsQueue
	}

	// 6. Set Gin mode
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	// 7. Setup Router & Global Middleware
	router := gin.New()
	router.Use(middleware.Recovery())
	router.Use(middleware.RequestLogger())
	router.Use(middleware.CORS(cfg.AllowedOrigins))

	// 8. Register Health & Readiness Checkers
	checkers := make(map[string]health.Checker)
	if db != nil {
		checkers["database"] = db.Ping
	} else {
		checkers["database"] = func(ctx context.Context) error {
			return errors.New("database not connected")
		}
	}

	healthHandler := health.NewHandler(checkers)
	healthHandler.RegisterRoutes(router)

	// 9. Initialize Repositories, Services, and Handlers
	authMiddleware := middleware.Auth(cfg.JWTSecret)

	v1 := router.Group("/api/v1")
	{
		v1.GET("/ping", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "pong"})
		})

		// Local / Mock Storage Endpoints (Upload & Playback)
		storageHandler := storage.NewHandler(storageService)
		storageHandler.RegisterRoutes(v1)

		if db != nil {
			// Auth Module
			userRepo := users.NewRepository(db.Pool)
			authService := auth.NewService(userRepo, cfg)
			authHandler := auth.NewHandler(authService)
			authHandler.RegisterRoutes(v1, authMiddleware)

			// Video Module
			videoRepo := videos.NewRepository(db.Pool)
			videoService := videos.NewService(videoRepo, storageService)
			videoHandler := videos.NewHandler(videoService)
			videoHandler.RegisterRoutes(v1, authMiddleware)

			// Jobs / Transcription Queue Trigger Module
			jobRepo := jobs.NewRepository(db.Pool)
			jobService := jobs.NewService(jobRepo, videoRepo, jobQueue)
			jobHandler := jobs.NewHandler(jobService)
			jobHandler.RegisterRoutes(v1, authMiddleware)

			// Transcript Retrieval & Export Module
			transcriptRepo := transcription.NewRepository(db.Pool)
			transcriptService := transcription.NewService(transcriptRepo, videoRepo)
			transcriptHandler := transcription.NewHandler(transcriptService)
			transcriptHandler.RegisterRoutes(v1, authMiddleware)
		}
	}

	// 10. Setup HTTP Server
	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  60 * time.Second,
	}

	// 11. Start server in a background goroutine
	go func() {
		slog.Info("HTTP server listening", slog.String("addr", server.Addr))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("HTTP server failed unexpectedly", slog.Any("error", err))
			os.Exit(1)
		}
	}()

	// 12. Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	sig := <-quit

	slog.Info("Shutdown signal received, shutting down gracefully...", slog.String("signal", sig.String()))

	// Context with 10 second timeout for active requests to complete
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("Server forced to shutdown", slog.Any("error", err))
		os.Exit(1)
	}

	slog.Info("Server stopped cleanly")
}
