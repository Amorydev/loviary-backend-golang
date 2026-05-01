package persistence

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"

	domainSparks "loviary.app/backend/internal/domain/sparks"
	apperrors "loviary.app/backend/pkg/errors"
)

// SparkRepository handles database operations for daily sparks.
type SparkRepository struct {
	db *sql.DB
}

// NewSparkRepository creates a new spark repository.
func NewSparkRepository(db *sql.DB) *SparkRepository {
	return &SparkRepository{db: db}
}

// GetTodayAssignment returns the spark assignment for a couple on a given date.
func (r *SparkRepository) GetTodayAssignment(ctx context.Context, coupleID uuid.UUID, date time.Time) (*domainSparks.SparkAssignment, error) {
	query := `
		SELECT assignment_id, spark_id, couple_id, assigned_date, expires_at
		FROM spark_assignments
		WHERE couple_id = $1 AND assigned_date = $2
	`
	var sa domainSparks.SparkAssignment
	err := r.db.QueryRowContext(ctx, query, coupleID, date.Format("2006-01-02")).Scan(
		&sa.AssignmentID,
		&sa.SparkID,
		&sa.CoupleID,
		&sa.AssignedDate,
		&sa.ExpiresAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, apperrors.New("INTERNAL_ERROR", "Failed to get spark assignment")
	}
	return &sa, nil
}

// GetSparkByID returns a daily spark by its ID.
func (r *SparkRepository) GetSparkByID(ctx context.Context, sparkID uuid.UUID) (*domainSparks.DailySpark, error) {
	query := `
		SELECT spark_id, question_text, category
		FROM daily_sparks
		WHERE spark_id = $1
	`
	var spark domainSparks.DailySpark
	err := r.db.QueryRowContext(ctx, query, sparkID).Scan(
		&spark.SparkID,
		&spark.Question,
		&spark.Category,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, apperrors.New("INTERNAL_ERROR", "Failed to get spark")
	}
	return &spark, nil
}

// GetUserResponse returns the spark response of a specific user.
func (r *SparkRepository) GetUserResponse(ctx context.Context, sparkID uuid.UUID, userID uuid.UUID) (*domainSparks.SparkResponse, error) {
	query := `
		SELECT response_id, spark_id, user_id, couple_id, answer, responded_at
		FROM spark_responses
		WHERE spark_id = $1 AND user_id = $2
	`
	var sr domainSparks.SparkResponse
	err := r.db.QueryRowContext(ctx, query, sparkID, userID).Scan(
		&sr.ResponseID,
		&sr.SparkID,
		&sr.UserID,
		&sr.CoupleID,
		&sr.Answer,
		&sr.RespondedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, apperrors.New("INTERNAL_ERROR", "Failed to get spark response")
	}
	return &sr, nil
}

// GetRandomSpark picks a random spark from the pool.
func (r *SparkRepository) GetRandomSpark(ctx context.Context) (*domainSparks.DailySpark, error) {
	query := `
		SELECT spark_id, question_text, category
		FROM daily_sparks
		ORDER BY RANDOM()
		LIMIT 1
	`
	var spark domainSparks.DailySpark
	err := r.db.QueryRowContext(ctx, query).Scan(
		&spark.SparkID,
		&spark.Question,
		&spark.Category,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, apperrors.New("INTERNAL_ERROR", "Failed to get random spark")
	}
	return &spark, nil
}

// AssignSpark inserts a new spark assignment for a couple.
// If one already exists for that date it is silently ignored (idempotent).
func (r *SparkRepository) AssignSpark(ctx context.Context, assignment *domainSparks.SparkAssignment) error {
	query := `
		INSERT INTO spark_assignments (assignment_id, spark_id, couple_id, assigned_date, expires_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (couple_id, assigned_date) DO NOTHING
	`
	_, err := r.db.ExecContext(ctx, query,
		assignment.AssignmentID,
		assignment.SparkID,
		assignment.CoupleID,
		assignment.AssignedDate.Format("2006-01-02"),
		assignment.ExpiresAt,
	)
	if err != nil {
		return apperrors.New("INTERNAL_ERROR", "Failed to assign spark")
	}
	return nil
}
