package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"video-transcription-service/internal/config"
	"video-transcription-service/internal/users"
	"video-transcription-service/pkg/crypto"
	"video-transcription-service/pkg/jwt"
)

var (
	ErrInvalidCredentials  = errors.New("invalid email or password")
	ErrInvalidRefreshToken = errors.New("invalid or expired refresh token")
)

type AuthResponse struct {
	User   *users.UserResponse `json:"user"`
	Tokens *jwt.TokenPair      `json:"tokens"`
}

type Service interface {
	Register(ctx context.Context, email, password string) (*AuthResponse, error)
	Login(ctx context.Context, email, password string) (*AuthResponse, error)
	RefreshToken(ctx context.Context, refreshToken string) (*AuthResponse, error)
	GetMe(ctx context.Context, userID uuid.UUID) (*users.UserResponse, error)
}

type service struct {
	userRepo users.Repository
	cfg      *config.Config
}

func NewService(userRepo users.Repository, cfg *config.Config) Service {
	return &service{
		userRepo: userRepo,
		cfg:      cfg,
	}
}

func (s *service) Register(ctx context.Context, email, password string) (*AuthResponse, error) {
	email = strings.TrimSpace(email)
	if email == "" || len(password) < 8 {
		return nil, errors.New("password must be at least 8 characters")
	}

	hash, err := crypto.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("failed to process password: %w", err)
	}

	user := &users.User{
		Email:        email,
		PasswordHash: hash,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	tokenPair, err := jwt.GenerateTokenPair(
		user.ID,
		user.Email,
		s.cfg.JWTSecret,
		s.cfg.JWTAccessExpiry,
		s.cfg.JWTRefreshExpiry,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate auth tokens: %w", err)
	}

	return &AuthResponse{
		User:   user.ToResponse(),
		Tokens: tokenPair,
	}, nil
}

func (s *service) Login(ctx context.Context, email, password string) (*AuthResponse, error) {
	email = strings.TrimSpace(email)
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, users.ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if !crypto.CheckPassword(password, user.PasswordHash) {
		return nil, ErrInvalidCredentials
	}

	tokenPair, err := jwt.GenerateTokenPair(
		user.ID,
		user.Email,
		s.cfg.JWTSecret,
		s.cfg.JWTAccessExpiry,
		s.cfg.JWTRefreshExpiry,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate auth tokens: %w", err)
	}

	return &AuthResponse{
		User:   user.ToResponse(),
		Tokens: tokenPair,
	}, nil
}

func (s *service) RefreshToken(ctx context.Context, refreshToken string) (*AuthResponse, error) {
	claims, err := jwt.ValidateToken(refreshToken, s.cfg.JWTSecret)
	if err != nil || claims.TokenType != jwt.TokenTypeRefresh {
		return nil, ErrInvalidRefreshToken
	}

	user, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, ErrInvalidRefreshToken
	}

	tokenPair, err := jwt.GenerateTokenPair(
		user.ID,
		user.Email,
		s.cfg.JWTSecret,
		s.cfg.JWTAccessExpiry,
		s.cfg.JWTRefreshExpiry,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate auth tokens: %w", err)
	}

	return &AuthResponse{
		User:   user.ToResponse(),
		Tokens: tokenPair,
	}, nil
}

func (s *service) GetMe(ctx context.Context, userID uuid.UUID) (*users.UserResponse, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return user.ToResponse(), nil
}
