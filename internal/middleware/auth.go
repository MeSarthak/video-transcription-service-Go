package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"video-transcription-service/pkg/jwt"
)

const (
	ContextUserIDKey    = "user_id"
	ContextUserEmailKey = "user_email"
)

// Auth returns a Gin middleware that validates the Bearer JWT access token.
func Auth(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "authorization header is required",
			})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "authorization header must be in format 'Bearer <token>'",
			})
			return
		}

		tokenStr := parts[1]
		claims, err := jwt.ValidateToken(tokenStr, jwtSecret)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid or expired token",
			})
			return
		}

		if claims.TokenType != jwt.TokenTypeAccess {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid token type: access token required",
			})
			return
		}

		c.Set(ContextUserIDKey, claims.UserID)
		c.Set(ContextUserEmailKey, claims.Email)
		c.Next()
	}
}

// GetUserID retrieves the authenticated user's UUID from the Gin context.
func GetUserID(c *gin.Context) (uuid.UUID, error) {
	val, exists := c.Get(ContextUserIDKey)
	if !exists {
		return uuid.Nil, errors.New("user ID not found in context")
	}

	userID, ok := val.(uuid.UUID)
	if !ok {
		return uuid.Nil, errors.New("invalid user ID type in context")
	}

	return userID, nil
}

// GetUserEmail retrieves the authenticated user's email from the Gin context.
func GetUserEmail(c *gin.Context) (string, error) {
	val, exists := c.Get(ContextUserEmailKey)
	if !exists {
		return "", errors.New("user email not found in context")
	}

	email, ok := val.(string)
	if !ok {
		return "", errors.New("invalid user email type in context")
	}

	return email, nil
}
