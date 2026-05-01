package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	domainStreaks "loviary.app/backend/internal/domain/streaks"
	apperrors "loviary.app/backend/pkg/errors"
)

// StreakRepository handles database operations for streaks.
// Streaks are per-couple per-activity. Both users must log for the streak to increment.
type StreakRepository struct {
	db *sql.DB
}

// NewStreakRepository creates a new streak repository.
func NewStreakRepository(db *sql.DB) *StreakRepository {
	return &StreakRepository{db: db}
}

// GetByCoupleAndActivity retrieves the streak for a couple and activity type.
func (r *StreakRepository) GetByCoupleAndActivity(ctx context.Context, coupleID uuid.UUID, activityType string) (*domainStreaks.Streak, error) {
	query := `
		SELECT streak_id, couple_id, activity_type, current_streak, longest_streak,
		       last_completed_date, created_at, updated_at
		FROM streaks
		WHERE couple_id = $1 AND activity_type = $2
	`
	var s domainStreaks.Streak
	err := r.db.QueryRowContext(ctx, query, coupleID, activityType).Scan(
		&s.StreakID,
		&s.CoupleID,
		&s.ActivityType,
		&s.CurrentStreak,
		&s.LongestStreak,
		&s.LastCompletedDate,
		&s.CreatedAt,
		&s.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, apperrors.New("INTERNAL_ERROR", fmt.Sprintf("failed to get streak for couple %s", coupleID))
	}
	return &s, nil
}

// GetAllByCouple retrieves all streaks for a couple.
func (r *StreakRepository) GetAllByCouple(ctx context.Context, coupleID uuid.UUID) ([]domainStreaks.Streak, error) {
	query := `
		SELECT streak_id, couple_id, activity_type, current_streak, longest_streak,
		       last_completed_date, created_at, updated_at
		FROM streaks
		WHERE couple_id = $1
		ORDER BY activity_type
	`
	rows, err := r.db.QueryContext(ctx, query, coupleID)
	if err != nil {
		return nil, apperrors.New("INTERNAL_ERROR", "Failed to query streaks")
	}
	defer rows.Close()

	var list []domainStreaks.Streak
	for rows.Next() {
		var s domainStreaks.Streak
		if err := rows.Scan(
			&s.StreakID,
			&s.CoupleID,
			&s.ActivityType,
			&s.CurrentStreak,
			&s.LongestStreak,
			&s.LastCompletedDate,
			&s.CreatedAt,
			&s.UpdatedAt,
		); err != nil {
			return nil, apperrors.New("INTERNAL_ERROR", "Failed to scan streak row")
		}
		list = append(list, s)
	}
	if err := rows.Err(); err != nil {
		return nil, apperrors.New("INTERNAL_ERROR", "Failed to iterate streak rows")
	}
	return list, nil
}

// Upsert inserts or updates a streak record.
func (r *StreakRepository) Upsert(ctx context.Context, s *domainStreaks.Streak) error {
	query := `
		INSERT INTO streaks (streak_id, couple_id, activity_type, current_streak, longest_streak,
		                     last_completed_date, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (couple_id, activity_type)
		DO UPDATE SET
		  current_streak      = EXCLUDED.current_streak,
		  longest_streak      = EXCLUDED.longest_streak,
		  last_completed_date = EXCLUDED.last_completed_date,
		  updated_at          = EXCLUDED.updated_at
	`
	_, err := r.db.ExecContext(ctx, query,
		s.StreakID,
		s.CoupleID,
		s.ActivityType,
		s.CurrentStreak,
		s.LongestStreak,
		s.LastCompletedDate,
		s.CreatedAt,
		s.UpdatedAt,
	)
	if err != nil {
		return apperrors.New("INTERNAL_ERROR", "Failed to upsert streak")
	}
	return nil
}

// RecordLog records that a user has completed an activity for a specific date.
// Uses INSERT ... ON CONFLICT DO NOTHING for idempotency.
func (r *StreakRepository) RecordLog(ctx context.Context, coupleID uuid.UUID, userID uuid.UUID, activityType string, date time.Time) error {
	query := `
		INSERT INTO streak_daily_logs (log_id, couple_id, user_id, activity_type, log_date)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (couple_id, user_id, activity_type, log_date) DO NOTHING
	`
	_, err := r.db.ExecContext(ctx, query,
		uuid.New(),
		coupleID,
		userID,
		activityType,
		date.Format("2006-01-02"),
	)
	if err != nil {
		return apperrors.New("INTERNAL_ERROR", "Failed to record streak daily log")
	}
	return nil
}

// CountLogsForDay returns how many distinct users have logged a given activity for the couple on a date.
func (r *StreakRepository) CountLogsForDay(ctx context.Context, coupleID uuid.UUID, activityType string, date time.Time) (int, error) {
	query := `
		SELECT COUNT(DISTINCT user_id)
		FROM streak_daily_logs
		WHERE couple_id = $1 AND activity_type = $2 AND log_date = $3
	`
	var count int
	err := r.db.QueryRowContext(ctx, query,
		coupleID,
		activityType,
		date.Format("2006-01-02"),
	).Scan(&count)
	if err != nil {
		return 0, apperrors.New("INTERNAL_ERROR", "Failed to count daily logs")
	}
	return count, nil
}

// GetWeeklyLog returns the last 7 days of log status for a couple and activity.
// completed = true means both users have logged on that day.
func (r *StreakRepository) GetWeeklyLog(ctx context.Context, coupleID uuid.UUID, activityType string) ([]domainStreaks.DayLog, error) {
	query := `
		SELECT
		  gs.day::date                                             AS log_date,
		  COALESCE(COUNT(DISTINCT sdl.user_id) = 2, FALSE)        AS completed
		FROM generate_series(
		  CURRENT_DATE - INTERVAL '6 days',
		  CURRENT_DATE,
		  INTERVAL '1 day'
		) AS gs(day)
		LEFT JOIN streak_daily_logs sdl
		  ON sdl.log_date   = gs.day::date
		 AND sdl.couple_id   = $1
		 AND sdl.activity_type = $2
		GROUP BY gs.day
		ORDER BY gs.day
	`
	rows, err := r.db.QueryContext(ctx, query, coupleID, activityType)
	if err != nil {
		return nil, apperrors.New("INTERNAL_ERROR", "Failed to query weekly log")
	}
	defer rows.Close()

	var logs []domainStreaks.DayLog
	for rows.Next() {
		var date time.Time
		var completed bool
		if err := rows.Scan(&date, &completed); err != nil {
			return nil, apperrors.New("INTERNAL_ERROR", "Failed to scan weekly log row")
		}
		logs = append(logs, domainStreaks.DayLog{
			Day:       date.Format("Mon"), // "Mon", "Tue", ...
			Completed: completed,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, apperrors.New("INTERNAL_ERROR", "Failed to iterate weekly log rows")
	}
	return logs, nil
}
