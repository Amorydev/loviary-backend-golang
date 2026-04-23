package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Claims represents JWT claims
type Claims struct {
	UserID   uuid.UUID `json:"user_id"`
	Username string    `json:"username"`
	Email    string    `json:"email"`
	CoupleID *uuid.UUID `json:"couple_id,omitempty"`
	jwt.RegisteredClaims
}

// RefreshToken represents a refresh token stored in DB
type RefreshToken struct {
	ID          uuid.UUID  `db:"id" json:"-"`
	UserID      uuid.UUID  `db:"user_id" json:"-"`
	TokenHash   string     `db:"token_hash" json:"-"`
	ExpiresAt   time.Time  `db:"expires_at" json:"-"`
	DeviceInfo  *string    `db:"device_info" json:"-"`
	IsRevoked   bool       `db:"is_revoked" json:"-"`
	CreatedAt   time.Time  `db:"created_at" json:"-"`
}

// IsExpired checks if the token is expired
func (rt *RefreshToken) IsExpired() bool {
	return time.Now().After(rt.ExpiresAt)
}

// IsValid checks if the token is valid (not expired and not revoked)
func (rt *RefreshToken) IsValid() bool {
	return !rt.IsExpired() && !rt.IsRevoked
}
