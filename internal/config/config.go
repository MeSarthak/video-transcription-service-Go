package config

import (
	"fmt"
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
	DatabaseURL    string
	DBMaxOpenConns int
	DBMaxIdleConns int
	DBMaxIdleTime  time.Duration

	// Authentication
	JWTSecret        string
	JWTAccessExpiry  time.Duration
	JWTRefreshExpiry time.Duration

	// AWS
	AWSRegion    string
	S3BucketName string
	SQSQueueURL  string

	// Whisper Speech-to-Text (Groq or OpenAI)
	OpenAIAPIKey  string
	GroqAPIKey    string
	WhisperAPIURL string
	WhisperModel  string
}

func Load() (*Config, error) {
	// Load .env file if present, ignore if not found (e.g. production/ECS/Docker)
	if err := godotenv.Load(); err != nil {
		if !os.IsNotExist(err) {
			slog.Debug("No .env file found, using system environment variables")
		}
	}

	// Resolve database URL:
	// 1. Explicit DATABASE_URL
	// 2. Constructed from DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME, DB_SSLMODE
	// 3. Default localhost connection string
	dbURL := getEnv("DATABASE_URL", "")
	if dbURL == "" {
		if dbHost := getEnv("DB_HOST", ""); dbHost != "" {
			dbUser := getEnv("DB_USER", "postgres")
			dbPass := getEnv("DB_PASSWORD", "postgres")
			dbPort := getEnv("DB_PORT", "5432")
			dbName := getEnv("DB_NAME", "transcription_db")
			dbSSL := getEnv("DB_SSLMODE", "disable")
			dbURL = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", dbUser, dbPass, dbHost, dbPort, dbName, dbSSL)
		} else {
			dbURL = "postgres://postgres:postgres@localhost:5432/transcription_db?sslmode=disable"
		}
	}

	cfg := &Config{
		Port:             getEnvWithFallback("PORT", "SERVER_PORT", "8080"),
		Env:              getEnvWithFallback("ENV", "SERVER_ENV", "development"),
		LogLevel:         getEnv("LOG_LEVEL", "info"),
		ReadTimeout:      getEnvDuration("READ_TIMEOUT", 15*time.Second),
		WriteTimeout:     getEnvDuration("WRITE_TIMEOUT", 15*time.Second),
		AllowedOrigins:   getEnvSlice("CORS_ALLOWED_ORIGINS", []string{"http://localhost:5173", "http://localhost:3000"}),
		DatabaseURL:      dbURL,
		DBMaxOpenConns:   getEnvInt("DB_MAX_OPEN_CONNS", 25),
		DBMaxIdleConns:   getEnvInt("DB_MAX_IDLE_CONNS", 10),
		DBMaxIdleTime:    getEnvDuration("DB_MAX_IDLE_TIME", 15*time.Minute),
		JWTSecret:        getEnv("JWT_SECRET", "default-dev-secret-change-in-production"),
		JWTAccessExpiry:  getEnvDuration("JWT_ACCESS_EXPIRY", 15*time.Minute),
		JWTRefreshExpiry: getEnvDuration("JWT_REFRESH_EXPIRY", 7*24*time.Hour),
		AWSRegion:        getEnv("AWS_REGION", "us-east-1"),
		S3BucketName:     getEnv("S3_BUCKET_NAME", "video-transcription-bucket"),
		SQSQueueURL:      getEnv("SQS_QUEUE_URL", ""),
		OpenAIAPIKey:     getEnv("OPENAI_API_KEY", ""),
		GroqAPIKey:       getEnv("GROQ_API_KEY", ""),
		WhisperAPIURL:    getEnv("WHISPER_API_URL", ""),
		WhisperModel:     getEnv("WHISPER_MODEL", ""),
	}

	return cfg, nil
}

func (c *Config) IsProduction() bool {
	return strings.ToLower(c.Env) == "production"
}

func (c *Config) HasAWSCredentials() bool {
	// 1. Explicit static AWS credentials in environment variables
	if os.Getenv("AWS_ACCESS_KEY_ID") != "" && os.Getenv("AWS_SECRET_ACCESS_KEY") != "" {
		return true
	}
	// 2. Production or explicit AWS cloud environment (uses EC2 IAM Instance Profile or ECS Task Role via AWS SDK default chain)
	if c.IsProduction() || strings.ToLower(os.Getenv("USE_AWS")) == "true" {
		return true
	}
	return false
}

func (c *Config) HasWhisperKey() bool {
	return c.GroqAPIKey != "" || c.OpenAIAPIKey != ""
}

func (c *Config) HasValidSQS() bool {
	return c.HasAWSCredentials() && c.SQSQueueURL != "" && !strings.Contains(c.SQSQueueURL, "123456789012")
}

func getEnv(key, defaultVal string) string {
	if val, exists := os.LookupEnv(key); exists && val != "" {
		return val
	}
	return defaultVal
}

func getEnvWithFallback(primaryKey, fallbackKey, defaultVal string) string {
	if val, exists := os.LookupEnv(primaryKey); exists && val != "" {
		return val
	}
	if val, exists := os.LookupEnv(fallbackKey); exists && val != "" {
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
