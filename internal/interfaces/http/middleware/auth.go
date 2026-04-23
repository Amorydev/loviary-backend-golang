package middleware

import (
	stdErr "errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"loviary.app/backend/internal/infrastructure/jwt"
	"loviary.app/backend/pkg/errors"
)

// AuthMiddleware validates JWT token
func AuthMiddleware(jwtManager *jwt.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, errors.New("UNAUTHORIZED", "Missing authorization header"))
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.JSON(http.StatusUnauthorized, errors.New("INVALID_TOKEN", "Invalid authorization format"))
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Validate token
		claims, err := jwtManager.ValidateAccessToken(tokenString)
		if err != nil {
			status := http.StatusUnauthorized
			if stdErr.Is(err, errors.ErrInvalidInput) {
				status = http.StatusBadRequest
			}
			c.JSON(status, err)
			c.Abort()
			return
		}

		// Store user info in context
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("email", claims.Email)
		c.Set("couple_id", claims.CoupleID)

		c.Next()
	}
}

// OptionalAuth allows optional authentication
func OptionalAuth(jwtManager *jwt.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.Next()
			return
		}

		tokenString := parts[1]
		claims, err := jwtManager.ValidateAccessToken(tokenString)
		if err != nil {
			c.Next()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("email", claims.Email)
		c.Set("couple_id", claims.CoupleID)
		c.Next()
	}
}

// GetUserID extracts user ID from context
func GetUserID(c *gin.Context) (uuid.UUID, bool) {
	val, exists := c.Get("user_id")
	if !exists {
		return uuid.Nil, false
	}
	id, ok := val.(uuid.UUID)
	return id, ok
}

// GetCoupleID extracts couple ID from context
func GetCoupleID(c *gin.Context) (*uuid.UUID, bool) {
	val, exists := c.Get("couple_id")
	if !exists {
		return nil, false
	}
	id, ok := val.(*uuid.UUID)
	if !ok || id == nil {
		return nil, false
	}
	return id, true
}
