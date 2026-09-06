package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"video-transcription-service/pkg/jwt"
)

func TestAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := "test-auth-middleware-secret-12345"
	userID := uuid.New()
	email := "auth_test@example.com"

	tokens, err := jwt.GenerateTokenPair(userID, email, secret, 15*time.Minute, 24*time.Hour)
	if err != nil {
		t.Fatalf("failed to generate tokens: %v", err)
	}

	setupRouter := func() *gin.Engine {
		r := gin.New()
		r.Use(Auth(secret))
		r.GET("/protected", func(c *gin.Context) {
			extractedID, _ := GetUserID(c)
			extractedEmail, _ := GetUserEmail(c)
			c.JSON(http.StatusOK, gin.H{
				"user_id": extractedID.String(),
				"email":   extractedEmail,
			})
		})
		return r
	}

	t.Run("Valid Access Token succeeds with 200 OK", func(t *testing.T) {
		router := setupRouter()
		req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d. Body: %s", w.Code, w.Body.String())
		}
	})

	t.Run("Missing Authorization Header fails with 401", func(t *testing.T) {
		router := setupRouter()
		req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected status 401, got %d", w.Code)
		}
	})

	t.Run("Invalid Token Signature fails with 401", func(t *testing.T) {
		router := setupRouter()
		req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer invalid.token.string")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected status 401, got %d", w.Code)
		}
	})

	t.Run("Refresh Token Passed as Access Token fails with 401", func(t *testing.T) {
		router := setupRouter()
		req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+tokens.RefreshToken)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected status 401 for refresh token, got %d", w.Code)
		}
	})
}
