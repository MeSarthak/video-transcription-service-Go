package auth

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"video-transcription-service/internal/config"
	"video-transcription-service/internal/users"
)

type mockUserRepo struct {
	usersByEmail map[string]*users.User
	usersByID    map[uuid.UUID]*users.User
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{
		usersByEmail: make(map[string]*users.User),
		usersByID:    make(map[uuid.UUID]*users.User),
	}
}

func (m *mockUserRepo) Create(ctx context.Context, user *users.User) error {
	if _, exists := m.usersByEmail[user.Email]; exists {
		return users.ErrDuplicateEmail
	}
	user.ID = uuid.New()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	m.usersByEmail[user.Email] = user
	m.usersByID[user.ID] = user
	return nil
}

func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*users.User, error) {
	user, exists := m.usersByEmail[email]
	if !exists {
		return nil, users.ErrUserNotFound
	}
	return user, nil
}

func (m *mockUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*users.User, error) {
	user, exists := m.usersByID[id]
	if !exists {
		return nil, users.ErrUserNotFound
	}
	return user, nil
}

func TestAuthService(t *testing.T) {
	cfg := &config.Config{
		JWTSecret:        "unit-test-secret-key-12345",
		JWTAccessExpiry:  15 * time.Minute,
		JWTRefreshExpiry: 24 * time.Hour,
	}

	repo := newMockUserRepo()
	svc := NewService(repo, cfg)
	ctx := context.Background()

	t.Run("Register success", func(t *testing.T) {
		resp, err := svc.Register(ctx, "user1@example.com", "Password123!")
		if err != nil {
			t.Fatalf("expected register success, got error: %v", err)
		}

		if resp.User.Email != "user1@example.com" {
			t.Errorf("expected email user1@example.com, got %s", resp.User.Email)
		}
		if resp.Tokens.AccessToken == "" || resp.Tokens.RefreshToken == "" {
			t.Errorf("expected tokens to be generated")
		}
	})

	t.Run("Register duplicate email fails", func(t *testing.T) {
		_, err := svc.Register(ctx, "user1@example.com", "Password123!")
		if err != users.ErrDuplicateEmail {
			t.Errorf("expected ErrDuplicateEmail, got %v", err)
		}
	})

	t.Run("Register short password fails", func(t *testing.T) {
		_, err := svc.Register(ctx, "short@example.com", "1234")
		if err == nil {
			t.Errorf("expected error for short password, got nil")
		}
	})

	t.Run("Login success", func(t *testing.T) {
		resp, err := svc.Login(ctx, "user1@example.com", "Password123!")
		if err != nil {
			t.Fatalf("expected login success, got error: %v", err)
		}

		if resp.User.Email != "user1@example.com" {
			t.Errorf("expected user1@example.com, got %s", resp.User.Email)
		}
	})

	t.Run("Login with invalid password fails", func(t *testing.T) {
		_, err := svc.Login(ctx, "user1@example.com", "WrongPassword!")
		if err != ErrInvalidCredentials {
			t.Errorf("expected ErrInvalidCredentials, got %v", err)
		}
	})

	t.Run("Login with non-existent email fails", func(t *testing.T) {
		_, err := svc.Login(ctx, "nonexistent@example.com", "Password123!")
		if err != ErrInvalidCredentials {
			t.Errorf("expected ErrInvalidCredentials, got %v", err)
		}
	})

	t.Run("Refresh token rotation success", func(t *testing.T) {
		loginResp, _ := svc.Login(ctx, "user1@example.com", "Password123!")
		refreshResp, err := svc.RefreshToken(ctx, loginResp.Tokens.RefreshToken)
		if err != nil {
			t.Fatalf("expected refresh token to succeed, got %v", err)
		}
		if refreshResp.Tokens.AccessToken == "" {
			t.Errorf("expected new access token")
		}
	})

	t.Run("Refresh token with invalid token fails", func(t *testing.T) {
		_, err := svc.RefreshToken(ctx, "invalid.token.string")
		if err != ErrInvalidRefreshToken {
			t.Errorf("expected ErrInvalidRefreshToken, got %v", err)
		}
	})

	t.Run("GetMe returns user details", func(t *testing.T) {
		loginResp, _ := svc.Login(ctx, "user1@example.com", "Password123!")
		me, err := svc.GetMe(ctx, loginResp.User.ID)
		if err != nil {
			t.Fatalf("expected GetMe success, got error: %v", err)
		}
		if me.Email != "user1@example.com" {
			t.Errorf("expected email user1@example.com, got %s", me.Email)
		}
	})
}
