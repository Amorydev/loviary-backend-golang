package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"loviary.app/backend/pkg/errors"
)

// Manager handles JWT token operations
type Manager struct {
	signingKey string
	issuer     string
	audience   string
	accessTTL  time.Duration
	refreshTTL time.Duration
}

// NewManager creates a new JWT manager
func NewManager(signingKey string, accessTTL, refreshTTL time.Duration, issuer, audience string) *Manager {
	return &Manager{
		signingKey: signingKey,
		issuer:     issuer,
		audience:   audience,
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

// AccessClaims represents claims for access token
type AccessClaims struct {
	UserID   uuid.UUID `json:"user_id"`
	Username string    `json:"username"`
	Email    string    `json:"email"`
	CoupleID *uuid.UUID `json:"couple_id,omitempty"`
	jwt.RegisteredClaims
}

// GenerateAccessToken generates a new access token
func (m *Manager) GenerateAccessToken(userID uuid.UUID, username, email string, coupleID *uuid.UUID) (string, error) {
	now := time.Now()
	claims := &AccessClaims{
		UserID:   userID,
		Username: username,
		Email:    email,
		CoupleID: coupleID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(m.accessTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    m.issuer,
			Audience:  []string{m.audience},
			Subject:   userID.String(),
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(m.signingKey))
}

// GenerateRefreshToken generates a refresh token identifier (we store hashed version in DB)
func (m *Manager) GenerateRefreshTokenID() (string, error) {
	return uuid.New().String(), nil
}

// ValidateAccessToken validates the access token and returns claims
func (m *Manager) ValidateAccessToken(tokenString string) (*AccessClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &AccessClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("INVALID_SIGNING_METHOD", "unexpected signing method")
		}
		return []byte(m.signingKey), nil
	})

	if err != nil {
		if err == jwt.ErrTokenExpired {
			return nil, errors.New("TOKEN_EXPIRED", "Token has expired")
		}
		return nil, errors.New("INVALID_TOKEN", "Invalid token")
	}

	if claims, ok := token.Claims.(*AccessClaims); ok && token.Valid {
		// Additional validation - check issuer
		if claims.Issuer != m.issuer {
			return nil, errors.New("INVALID_TOKEN", "Invalid issuer")
		}
		// Check audience
		if len(claims.Audience) == 0 || claims.Audience[0] != m.audience {
			return nil, errors.New("INVALID_TOKEN", "Invalid audience")
		}
		// Check expiration (JWT library already does this but we double-check)
		if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
			return nil, errors.New("TOKEN_EXPIRED", "Token has expired")
		}
		return claims, nil
	}

	return nil, errors.New("INVALID_TOKEN", "Invalid token")
}
