package persistence

import (
    "context"
    "database/sql"
    "time"

    "github.com/google/uuid"
    "github.com/jmoiron/sqlx"
    "github.com/lib/pq"

    "loviary.app/backend/internal/domain/verification"
    apperrors "loviary.app/backend/pkg/errors"
)

// EmailVerificationRepository implements verification.Repository using PostgreSQL
type EmailVerificationRepository struct {
    db *sqlx.DB
}

// NewEmailVerificationRepository creates a new email verification repository
func NewEmailVerificationRepository(db *sqlx.DB) *EmailVerificationRepository {
    return &EmailVerificationRepository{db: db}
}

// Create inserts a new email verification
func (r *EmailVerificationRepository) Create(ctx context.Context, v *verification.EmailVerification) error {
    const query = `
        INSERT INTO email_verifications (id, user_id, code, expires_at, verified_at, created_at)
        VALUES (:id, :user_id, :code, :expires_at, :verified_at, :created_at)
    `

    _, err := r.db.NamedExecContext(ctx, query, v)
    if err != nil {
        if pqErr, ok := err.(*pq.Error); ok {
            if pqErr.Code == "23505" && pqErr.Constraint == "email_verifications_user_id_key" {
                return apperrors.New("VERIFICATION_EXISTS", "Verification already exists for user")
            }
        }
        return apperrors.NewWith("INTERNAL_ERROR", "Failed to create email verification", err)
    }
    return nil
}

// GetByUserID retrieves an email verification by user ID
func (r *EmailVerificationRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*verification.EmailVerification, error) {
    const query = `
        SELECT id, user_id, code, expires_at, verified_at, created_at
        FROM email_verifications
        WHERE user_id = $1
    `

    var v verification.EmailVerification
    err := r.db.GetContext(ctx, &v, query, userID)
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, nil
        }
        return nil, err
    }
    return &v, nil
}

// GetByCode retrieves an email verification by verification code
func (r *EmailVerificationRepository) GetByCode(ctx context.Context, code string) (*verification.EmailVerification, error) {
    const query = `
        SELECT id, user_id, code, expires_at, verified_at, created_at
        FROM email_verifications
        WHERE code = $1
    `

    var v verification.EmailVerification
    err := r.db.GetContext(ctx, &v, query, code)
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, apperrors.New("VERIFICATION_NOT_FOUND", "Verification code not found")
        }
        return nil, err
    }
    return &v, nil
}

// Verify marks the verification as verified by setting verified_at
func (r *EmailVerificationRepository) Verify(ctx context.Context, id uuid.UUID) error {
    const query = `
        UPDATE email_verifications
        SET verified_at = $1
        WHERE id = $2
    `
    now := time.Now()
    _, err := r.db.ExecContext(ctx, query, now, id)
    if err != nil {
        return apperrors.NewWith("INTERNAL_ERROR", "Failed to verify email", err)
    }
    return nil
}

// Delete removes an email verification record by user ID
func (r *EmailVerificationRepository) Delete(ctx context.Context, userID uuid.UUID) error {
    const query = `DELETE FROM email_verifications WHERE user_id = $1`
    _, err := r.db.ExecContext(ctx, query, userID)
    if err != nil {
        return apperrors.NewWith("INTERNAL_ERROR", "Failed to delete email verification", err)
    }
    return nil
}
