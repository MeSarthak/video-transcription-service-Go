package jwt

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestJWTGenerationAndValidation(t *testing.T) {
	secret := "test-secret-key-for-jwt-unit-tests-12345"
	userID := uuid.New()
	email := "test@example.com"
	accessDur := 15 * time.Minute
	refreshDur := 24 * time.Hour

	tokenPair, err := GenerateTokenPair(userID, email, secret, accessDur, refreshDur)
	if err != nil {
		t.Fatalf("expected no error generating token pair, got %v", err)
	}

	if tokenPair.AccessToken == "" || tokenPair.RefreshToken == "" {
		t.Fatalf("tokens should not be empty")
	}

	// Validate Access Token
	accessClaims, err := ValidateToken(tokenPair.AccessToken, secret)
	if err != nil {
		t.Fatalf("expected access token to be valid, got error: %v", err)
	}
	if accessClaims.UserID != userID {
		t.Errorf("expected user ID %v, got %v", userID, accessClaims.UserID)
	}
	if accessClaims.Email != email {
		t.Errorf("expected email %s, got %s", email, accessClaims.Email)
	}
	if accessClaims.TokenType != TokenTypeAccess {
		t.Errorf("expected token type %s, got %s", TokenTypeAccess, accessClaims.TokenType)
	}

	// Validate Refresh Token
	refreshClaims, err := ValidateToken(tokenPair.RefreshToken, secret)
	if err != nil {
		t.Fatalf("expected refresh token to be valid, got error: %v", err)
	}
	if refreshClaims.TokenType != TokenTypeRefresh {
		t.Errorf("expected token type %s, got %s", TokenTypeRefresh, refreshClaims.TokenType)
	}

	// Validate with wrong secret
	_, err = ValidateToken(tokenPair.AccessToken, "wrong-secret-key")
	if err == nil {
		t.Errorf("expected error validating token with wrong secret, got nil")
	}
}
