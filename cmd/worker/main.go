package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"video-transcription-service/internal/config"
	"video-transcription-service/internal/database"
	"video-transcription-service/internal/jobs"
	"video-transcription-service/internal/queue"
	"video-transcription-service/internal/storage"
	"video-transcription-service/internal/videos"
	"video-transcription-service/internal/worker"
	"video-transcription-service/pkg/ffmpeg"
	"video-transcription-service/pkg/logger"
	"video-transcription-service/pkg/transcribe"
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
	slog.Info("Starting Transcription Worker Service",
		slog.String("env", cfg.Env),
		slog.String("log_level", cfg.LogLevel),
	)

	// 3. Initialize Database connection
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db, err := database.New(ctx, cfg)
	if err != nil {
		slog.Error("Worker failed to connect to database", slog.Any("error", err))
		os.Exit(1)
	}
	defer db.Close()

	// 4. Initialize S3 Storage
	var storageService storage.Service
	if !cfg.HasAWSCredentials() && !cfg.IsProduction() {
		slog.Info("Using Mock Storage for local development (no AWS credentials configured)")
		storageService = storage.NewMockStorage()
	} else {
		s3Store, err := storage.NewS3Storage(ctx, cfg.AWSRegion, cfg.S3BucketName)
		if err != nil {
			if cfg.IsProduction() {
				slog.Error("Worker failed to initialize AWS S3", slog.Any("error", err))
				os.Exit(1)
			} else {
				slog.Warn("AWS S3 initialization warning (using mock storage for dev)", slog.Any("error", err))
				storageService = storage.NewMockStorage()
			}
		} else {
			storageService = s3Store
		}
	}

	// 5. Initialize SQS Queue
	var jobQueue queue.Queue
	if !cfg.HasValidSQS() && !cfg.IsProduction() {
		slog.Info("Using Mock Queue for local development (no AWS credentials configured or placeholder SQS URL)")
		jobQueue = queue.NewMockQueue()
	} else {
		sqsQueue, err := queue.NewSQSQueue(ctx, cfg.AWSRegion, cfg.SQSQueueURL)
		if err != nil {
			if cfg.IsProduction() {
				slog.Error("Worker failed to initialize AWS SQS", slog.Any("error", err))
				os.Exit(1)
			} else {
				slog.Warn("AWS SQS initialization warning (using mock queue for dev)", slog.Any("error", err))
				jobQueue = queue.NewMockQueue()
			}
		} else {
			jobQueue = sqsQueue
		}
	}

	// 6. Repositories
	jobRepo := jobs.NewRepository(db.Pool)
	videoRepo := videos.NewRepository(db.Pool)

	// 7. Initialize FFmpeg Extractor
	var extractor ffmpeg.Extractor = ffmpeg.NewCLIFFmpeg(5 * time.Minute)

	// 8. Initialize Transcribe Provider
	var transcribeProvider transcribe.Provider
	if cfg.HasWhisperKey() {
		apiKey := cfg.GroqAPIKey
		apiURL := "https://api.groq.com/openai/v1/audio/transcriptions"
		model := "whisper-large-v3"
		providerName := "Groq Whisper (whisper-large-v3)"

		if cfg.OpenAIAPIKey != "" && cfg.GroqAPIKey == "" {
			apiKey = cfg.OpenAIAPIKey
			apiURL = "https://api.openai.com/v1/audio/transcriptions"
			model = "whisper-1"
			providerName = "OpenAI Whisper (whisper-1)"
		}

		if cfg.WhisperAPIURL != "" {
			apiURL = cfg.WhisperAPIURL
		}
		if cfg.WhisperModel != "" {
			model = cfg.WhisperModel
		}

		slog.Info("Initializing Whisper Speech-to-Text Provider",
			slog.String("provider", providerName),
			slog.String("model", model),
			slog.String("api_url", apiURL),
		)
		transcribeProvider = transcribe.NewWhisperProvider(apiKey, apiURL, model)
	} else if cfg.HasAWSCredentials() {
		awsProvider, err := transcribe.NewAWSTranscribeProvider(ctx, cfg.AWSRegion)
		if err != nil {
			if cfg.IsProduction() {
				slog.Error("Worker failed to initialize AWS Transcribe", slog.Any("error", err))
				os.Exit(1)
			} else {
				slog.Warn("AWS Transcribe initialization warning (using mock transcribe for dev)", slog.Any("error", err))
				transcribeProvider = transcribe.NewMockTranscribeProvider()
			}
		} else {
			transcribeProvider = awsProvider
		}
	} else {
		slog.Warn("================================================================================")
		slog.Warn("NO SPEECH-TO-TEXT API KEY CONFIGURED!")
		slog.Warn("Using Mock Transcribe Provider for local development (returns placeholder text).")
		slog.Warn("To enable REAL AI transcription, set GROQ_API_KEY (free & fast) or OPENAI_API_KEY in .env!")
		slog.Warn("Get a free Groq key in 30 seconds at: https://console.groq.com/keys")
		slog.Warn("================================================================================")
		transcribeProvider = transcribe.NewMockTranscribeProvider()
	}

	// 9. Pipeline Processor
	pipelineProcessor := worker.NewProcessor(
		db,
		storageService,
		extractor,
		transcribeProvider,
		jobRepo,
		videoRepo,
		cfg.S3BucketName,
	)

	// 10. Start Worker Engine
	workerEngine := worker.NewWorker(jobQueue, jobRepo, videoRepo, pipelineProcessor.ProcessJob)

	// Listen for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	workerCtx, workerCancel := context.WithCancel(ctx)

	go func() {
		if err := workerEngine.Start(workerCtx); err != nil && !errorsIs(err, context.Canceled) {
			slog.Error("Worker loop exited with error", slog.Any("error", err))
		}
	}()

	sig := <-quit
	slog.Info("Shutdown signal received, gracefully shutting down worker...", slog.String("signal", sig.String()))

	workerCancel()

	// Allow pending in-flight tasks to finish
	time.Sleep(1 * time.Second)
	slog.Info("Worker process stopped cleanly")
}

func errorsIs(err, target error) bool {
	return err == target
}
