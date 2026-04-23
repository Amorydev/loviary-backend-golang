package persistence

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"loviary.app/backend/internal/domain/auth"
	"loviary.app/backend/pkg/errors"
)

// RefreshTokenRepository implements auth.Repository using PostgreSQL
type RefreshTokenRepository struct {
	db *sqlx.DB
}

// NewRefreshTokenRepository creates a new refresh token repository
func NewRefreshTokenRepository(db *sqlx.DB) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

// Create inserts a new refresh token
func (r *RefreshTokenRepository) Create(ctx context.Context, token *auth.RefreshToken) error {
	const query = `
		INSERT INTO refresh_tokens (
			id, user_id, token_hash, expires_at, device_info, is_revoked, created_at
		) VALUES (
			:id, :user_id, :token_hash, :expires_at, :device_info, :is_revoked, :created_at
		)
	`

	_, err := r.db.NamedExecContext(ctx, query, token)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code == "23505" && pqErr.Constraint == "refresh_tokens_token_hash_key" {
				return errors.New("TOKEN_ALREADY_EXISTS", "Refresh token already exists")
			}
		}
		return errors.NewWith("INTERNAL_ERROR", "Failed to create refresh token", err)
	}
	return nil
}

// GetByTokenHash retrieves a refresh token by its hash
func (r *RefreshTokenRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*auth.RefreshToken, error) {
	const query = `
		SELECT id, user_id, token_hash, expires_at, device_info, is_revoked, created_at
		FROM refresh_tokens
		WHERE token_hash = $1 AND is_revoked = false
	`

	var token auth.RefreshToken
	err := r.db.GetContext(ctx, &token, query, tokenHash)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}
	return &token, nil
}

// GetByUserID retrieves active refresh tokens for a user
func (r *RefreshTokenRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]auth.RefreshToken, error) {
	const query = `
		SELECT id, user_id, token_hash, expires_at, device_info, is_revoked, created_at
		FROM refresh_tokens
		WHERE user_id = $1 AND is_revoked = false
		ORDER BY created_at DESC
	`

	var tokens []auth.RefreshToken
	err := r.db.SelectContext(ctx, &tokens, query, userID)
	if err != nil {
		return nil, err
	}
	return tokens, nil
}

// Revoke revokes a refresh token
func (r *RefreshTokenRepository) Revoke(ctx context.Context, id uuid.UUID) error {
	const query = `
		UPDATE refresh_tokens
		SET is_revoked = true,
		    updated_at = NOW()
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return errors.NewWith("INTERNAL_ERROR", "Failed to revoke token", err)
	}
	return nil
}

// RevokeAllForUser revokes all refresh tokens for a user
func (r *RefreshTokenRepository) RevokeAllForUser(ctx context.Context, userID uuid.UUID) error {
	const query = `
		UPDATE refresh_tokens
		SET is_revoked = true,
		    updated_at = NOW()
		WHERE user_id = $1 AND is_revoked = false
	`

	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return errors.NewWith("INTERNAL_ERROR", "Failed to revoke all tokens", err)
	}
	return nil
}

// DeleteExpired deletes expired tokens
func (r *RefreshTokenRepository) DeleteExpired(ctx context.Context) error {
	const query = `
		DELETE FROM refresh_tokens
		WHERE expires_at < NOW() OR is_revoked = true
	`

	_, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return errors.NewWith("INTERNAL_ERROR", "Failed to delete expired tokens", err)
	}
	return nil
}
