package config

import (
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	// Server
	Port         string
	Env          string
	LogLevel     string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration

	// CORS
	AllowedOrigins []string

	// Database
	DatabaseURL     string
	DBMaxOpenConns  int
	DBMaxIdleConns  int
	DBMaxIdleTime   time.Duration

	// Authentication
	JWTSecret        string
	JWTAccessExpiry  time.Duration
	JWTRefreshExpiry time.Duration

	// AWS
	AWSRegion    string
	S3BucketName string
	SQSQueueURL  string
}

func Load() (*Config, error) {
	// Load .env file if present, ignore if not found (e.g. production/ECS)
	if err := godotenv.Load(); err != nil {
		if !os.IsNotExist(err) {
			slog.Debug("No .env file found, using system environment variables")
		}
	}

	cfg := &Config{
		Port:             getEnv("PORT", "8080"),
		Env:              getEnv("ENV", "development"),
		LogLevel:         getEnv("LOG_LEVEL", "info"),
		ReadTimeout:      getEnvDuration("READ_TIMEOUT", 15*time.Second),
		WriteTimeout:     getEnvDuration("WRITE_TIMEOUT", 15*time.Second),
		AllowedOrigins:   getEnvSlice("CORS_ALLOWED_ORIGINS", []string{"http://localhost:5173", "http://localhost:3000"}),
		DatabaseURL:      getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/transcription_db?sslmode=disable"),
		DBMaxOpenConns:   getEnvInt("DB_MAX_OPEN_CONNS", 25),
		DBMaxIdleConns:   getEnvInt("DB_MAX_IDLE_CONNS", 10),
		DBMaxIdleTime:    getEnvDuration("DB_MAX_IDLE_TIME", 15*time.Minute),
		JWTSecret:        getEnv("JWT_SECRET", "default-dev-secret-change-in-production"),
		JWTAccessExpiry:  getEnvDuration("JWT_ACCESS_EXPIRY", 15*time.Minute),
		JWTRefreshExpiry: getEnvDuration("JWT_REFRESH_EXPIRY", 7*24*time.Hour),
		AWSRegion:        getEnv("AWS_REGION", "us-east-1"),
		S3BucketName:     getEnv("S3_BUCKET_NAME", "video-transcription-bucket"),
		SQSQueueURL:      getEnv("SQS_QUEUE_URL", ""),
	}

	return cfg, nil
}

func (c *Config) IsProduction() bool {
	return strings.ToLower(c.Env) == "production"
}

func getEnv(key, defaultVal string) string {
	if val, exists := os.LookupEnv(key); exists && val != "" {
		return val
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if val, exists := os.LookupEnv(key); exists && val != "" {
		if intVal, err := strconv.Atoi(val); err == nil {
			return intVal
		}
	}
	return defaultVal
}

func getEnvDuration(key string, defaultVal time.Duration) time.Duration {
	if val, exists := os.LookupEnv(key); exists && val != "" {
		if d, err := time.ParseDuration(val); err == nil {
			return d
		}
	}
	return defaultVal
}

func getEnvSlice(key string, defaultVal []string) []string {
	if val, exists := os.LookupEnv(key); exists && val != "" {
		items := strings.Split(val, ",")
		var cleaned []string
		for _, item := range items {
			trimmed := strings.TrimSpace(item)
			if trimmed != "" {
				cleaned = append(cleaned, trimmed)
			}
		}
		if len(cleaned) > 0 {
			return cleaned
		}
	}
	return defaultVal
}
