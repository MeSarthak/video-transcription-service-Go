package config

import (
	"os"
	"testing"
	"time"
)

func TestConfigLoadDefaults(t *testing.T) {
	// Clear any overrides for the test
	os.Unsetenv("PORT")
	os.Unsetenv("ENV")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error loading config, got %v", err)
	}

	if cfg.Port != "8080" {
		t.Errorf("expected default port 8080, got %s", cfg.Port)
	}

	if cfg.Env != "development" {
		t.Errorf("expected default env development, got %s", cfg.Env)
	}

	if cfg.ReadTimeout != 15*time.Second {
		t.Errorf("expected default read timeout 15s, got %v", cfg.ReadTimeout)
	}
}

func TestConfigCustomEnv(t *testing.T) {
	os.Setenv("PORT", "9090")
	os.Setenv("ENV", "production")
	defer func() {
		os.Unsetenv("PORT")
		os.Unsetenv("ENV")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error loading config, got %v", err)
	}

	if cfg.Port != "9090" {
		t.Errorf("expected port 9090, got %s", cfg.Port)
	}

	if !cfg.IsProduction() {
		t.Errorf("expected IsProduction() to be true")
	}
}

func TestConfigDBHostConstruction(t *testing.T) {
	os.Unsetenv("DATABASE_URL")
	os.Setenv("DB_HOST", "postgres-cluster")
	os.Setenv("DB_PORT", "5433")
	os.Setenv("DB_USER", "custom_user")
	os.Setenv("DB_PASSWORD", "custom_pass")
	os.Setenv("DB_NAME", "custom_db")
	os.Setenv("DB_SSLMODE", "require")
	defer func() {
		os.Unsetenv("DB_HOST")
		os.Unsetenv("DB_PORT")
		os.Unsetenv("DB_USER")
		os.Unsetenv("DB_PASSWORD")
		os.Unsetenv("DB_NAME")
		os.Unsetenv("DB_SSLMODE")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error loading config, got %v", err)
	}

	expectedURL := "postgres://custom_user:custom_pass@postgres-cluster:5433/custom_db?sslmode=require"
	if cfg.DatabaseURL != expectedURL {
		t.Errorf("expected DatabaseURL %s, got %s", expectedURL, cfg.DatabaseURL)
	}
}
