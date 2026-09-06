package health

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestHealthEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	handler := NewHandler()
	handler.RegisterRoutes(router)

	t.Run("GET /health returns 200 OK", func(t *testing.T) {
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

	t.Run("GET /ready returns 200 OK", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/ready", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}

		var resp map[string]any
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if resp["status"] != "ready" {
			t.Errorf("expected status 'ready', got %v", resp["status"])
		}
	})
}
