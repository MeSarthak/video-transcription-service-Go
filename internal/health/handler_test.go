package health

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestHealthEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("GET /health returns 200 OK", func(t *testing.T) {
		router := gin.New()
		handler := NewHandler(nil)
		handler.RegisterRoutes(router)

		req, _ := http.NewRequest(http.MethodGet, "/health", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}

		var resp map[string]any
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if resp["status"] != "ok" {
			t.Errorf("expected status 'ok', got %v", resp["status"])
		}
	})

	t.Run("GET /ready returns 200 OK when all checkers pass", func(t *testing.T) {
		router := gin.New()
		checkers := map[string]Checker{
			"database": func(ctx context.Context) error { return nil },
		}
		handler := NewHandler(checkers)
		handler.RegisterRoutes(router)

		req, _ := http.NewRequest(http.MethodGet, "/ready", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}

		var resp struct {
			Status string            `json:"status"`
			Checks map[string]string `json:"checks"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if resp.Status != "ready" {
			t.Errorf("expected status 'ready', got %s", resp.Status)
		}
		if resp.Checks["database"] != "ok" {
			t.Errorf("expected database check 'ok', got %s", resp.Checks["database"])
		}
	})

	t.Run("GET /ready returns 503 when a checker fails", func(t *testing.T) {
		router := gin.New()
		checkers := map[string]Checker{
			"database": func(ctx context.Context) error { return errors.New("connection refused") },
		}
		handler := NewHandler(checkers)
		handler.RegisterRoutes(router)

		req, _ := http.NewRequest(http.MethodGet, "/ready", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusServiceUnavailable {
			t.Fatalf("expected status 503, got %d", w.Code)
		}

		var resp struct {
			Status string            `json:"status"`
			Checks map[string]string `json:"checks"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if resp.Status != "unhealthy" {
			t.Errorf("expected status 'unhealthy', got %s", resp.Status)
		}
	})
}
