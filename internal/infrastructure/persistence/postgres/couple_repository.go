package persistence

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"loviary.app/backend/internal/domain/couples"
	"loviary.app/backend/pkg/errors"
)

// CoupleRepository implements couples.Repository using PostgreSQL
type CoupleRepository struct {
	db *sqlx.DB
}

// NewCoupleRepository creates a new couple repository
func NewCoupleRepository(db *sqlx.DB) *CoupleRepository {
	return &CoupleRepository{db: db}
}

// Create inserts a new couple
func (r *CoupleRepository) Create(ctx context.Context, couple *couples.Couple) error {
	const query = `
		INSERT INTO couples (
			couple_id, user1_id, user2_id, couple_name,
			relationship_start_date, status, relationship_type,
			invitation_expires_at, breakup_initiated_by, breakup_grace_until,
			created_at, updated_at
		) VALUES (
			:couple_id, :user1_id, :user2_id, :couple_name,
			:relationship_start_date, :status, :relationship_type,
			:invitation_expires_at, :breakup_initiated_by, :breakup_grace_until,
			:created_at, :updated_at
		)
	`

	_, err := r.db.NamedExecContext(ctx, query, couple)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code == "23505" && pqErr.Constraint == "couples_user1_id_key" {
				return errors.New("USER1_ALREADY_HAS_ACTIVE_COUPLE", "User 1 already has an active couple")
			}
			if pqErr.Code == "23505" && pqErr.Constraint == "couples_user2_id_key" {
				return errors.New("USER2_ALREADY_HAS_ACTIVE_COUPLE", "User 2 already has an active couple")
			}
		}
		return errors.NewWith("INTERNAL_ERROR", "Failed to create couple", err)
	}

	return nil
}

// GetByID retrieves a couple by ID
func (r *CoupleRepository) GetByID(ctx context.Context, id uuid.UUID) (*couples.Couple, error) {
	const query = `
		SELECT couple_id, user1_id, user2_id, couple_name,
		       relationship_start_date, status, relationship_type,
		       invitation_expires_at, breakup_initiated_by, breakup_grace_until,
		       created_at, updated_at
		FROM couples
		WHERE couple_id = $1
	`

	var couple couples.Couple
	err := r.db.GetContext(ctx, &couple, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}
	return &couple, nil
}

// GetByUserID retrieves any couple for a user (all statuses)
func (r *CoupleRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*couples.Couple, error) {
	const query = `
		SELECT couple_id, user1_id, user2_id, couple_name,
		       relationship_start_date, status, relationship_type,
		       invitation_expires_at, breakup_initiated_by, breakup_grace_until,
		       created_at, updated_at
		FROM couples
		WHERE user1_id = $1 OR user2_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`

	var couple couples.Couple
	err := r.db.GetContext(ctx, &couple, query, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}
	return &couple, nil
}

// GetActiveByUserID retrieves the active couple for a user
func (r *CoupleRepository) GetActiveByUserID(ctx context.Context, userID uuid.UUID) (*couples.Couple, error) {
	const query = `
		SELECT couple_id, user1_id, user2_id, couple_name,
		       relationship_start_date, status, relationship_type,
		       invitation_expires_at, breakup_initiated_by, breakup_grace_until,
		       created_at, updated_at
		FROM couples
		WHERE (user1_id = $1 OR user2_id = $1) AND status = 'active'
		LIMIT 1
	`

	var couple couples.Couple
	err := r.db.GetContext(ctx, &couple, query, userID)
	if err != nil {
		return nil, err
	}
	return &couple, nil
}

// Update updates a couple
func (r *CoupleRepository) Update(ctx context.Context, couple *couples.Couple) error {
	const query = `
		UPDATE couples
		SET user1_id = :user1_id,
		    user2_id = :user2_id,
		    couple_name = :couple_name,
		    relationship_start_date = :relationship_start_date,
		    status = :status,
		    relationship_type = :relationship_type,
		    invitation_expires_at = :invitation_expires_at,
		    breakup_initiated_by = :breakup_initiated_by,
		    breakup_grace_until = :breakup_grace_until,
		    updated_at = :updated_at
		WHERE couple_id = :couple_id
	`

	couple.UpdatedAt = time.Now()
	_, err := r.db.NamedExecContext(ctx, query, couple)
	if err != nil {
		return errors.NewWith("INTERNAL_ERROR", "Failed to update couple", err)
	}
	return nil
}

// Delete removes a couple
func (r *CoupleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	const query = `DELETE FROM couples WHERE couple_id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return errors.NewWith("INTERNAL_ERROR", "Failed to delete couple", err)
	}
	return nil
}

// AcceptInvitation marks the invitation as accepted (activates couple)
func (r *CoupleRepository) AcceptInvitation(ctx context.Context, coupleID, userID uuid.UUID) error {
	const query = `
		UPDATE couples
		SET status = 'active',
		    invitation_expires_at = NULL,
		    updated_at = NOW()
		WHERE couple_id = $1 AND user2_id = $2 AND status = 'pending_invitation'
	`

	result, err := r.db.ExecContext(ctx, query, coupleID, userID)
	if err != nil {
		return errors.NewWith("INTERNAL_ERROR", "Failed to accept invitation", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("INVITATION_NOT_FOUND", "Invitation not found or already accepted")
	}
	return nil
}

// InitiateBreakup sets couple to grace_period status
func (r *CoupleRepository) InitiateBreakup(ctx context.Context, coupleID, userID uuid.UUID) error {
	const query = `
		UPDATE couples
		SET status = 'grace_period',
		    breakup_initiated_by = $2,
		    breakup_grace_until = NOW() + INTERVAL '7 days',
		    updated_at = NOW()
		WHERE couple_id = $1 AND status = 'active'
	`

	result, err := r.db.ExecContext(ctx, query, coupleID, userID)
	if err != nil {
		return errors.NewWith("INTERNAL_ERROR", "Failed to initiate breakup", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("COUPLE_NOT_ACTIVE", "Couple is not active")
	}
	return nil
}

// ConfirmBreakup marks couple as ended
func (r *CoupleRepository) ConfirmBreakup(ctx context.Context, coupleID uuid.UUID) error {
	const query = `
		UPDATE couples
		SET status = 'ended',
		    updated_at = NOW()
		WHERE couple_id = $1 AND status = 'grace_period'
	`

	result, err := r.db.ExecContext(ctx, query, coupleID)
	if err != nil {
		return errors.NewWith("INTERNAL_ERROR", "Failed to confirm breakup", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("COUPLE_NOT_IN_GRACE_PERIOD", "Couple is not in grace period")
	}
	return nil
}

// CancelBreakup reverts couple back to active
func (r *CoupleRepository) CancelBreakup(ctx context.Context, coupleID uuid.UUID) error {
	const query = `
		UPDATE couples
		SET status = 'active',
		    breakup_initiated_by = NULL,
		    breakup_grace_until = NULL,
		    updated_at = NOW()
		WHERE couple_id = $1 AND status = 'grace_period'
	`

	result, err := r.db.ExecContext(ctx, query, coupleID)
	if err != nil {
		return errors.NewWith("INTERNAL_ERROR", "Failed to cancel breakup", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("COUPLE_NOT_IN_GRACE_PERIOD", "Couple is not in grace period")
	}
	return nil
}

// List retrieves a paginated list of couples
func (r *CoupleRepository) List(ctx context.Context, limit, offset int) ([]couples.Couple, error) {
	const query = `
		SELECT couple_id, user1_id, user2_id, couple_name,
		       relationship_start_date, status, relationship_type,
		       invitation_expires_at, breakup_initiated_by, breakup_grace_until,
		       created_at, updated_at
		FROM couples
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	var couplesList []couples.Couple
	err := r.db.SelectContext(ctx, &couplesList, query, limit, offset)
	if err != nil {
		return nil, err
	}
	return couplesList, nil
}

// Count returns the total number of couples
func (r *CoupleRepository) Count(ctx context.Context) (int, error) {
	const query = `SELECT COUNT(*) FROM couples`
	var count int
	err := r.db.GetContext(ctx, &count, query)
	if err != nil {
		return 0, err
	}
	return count, nil
}
