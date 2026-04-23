package persistence

import (
    "context"
    "database/sql"
    "errors"
    "time"

    "github.com/google/uuid"

    "loviary.app/backend/internal/domain/streaks"
    apperrors "loviary.app/backend/pkg/errors"
)

// StreakRepository handles database operations for streaks
type StreakRepository struct {
    db *sql.DB
}

// NewStreakRepository creates a new streak repository
func NewStreakRepository(db *sql.DB) *StreakRepository {
    return &StreakRepository{db: db}
}

// Create inserts a new streak
func (r *StreakRepository) Create(ctx context.Context, streak *streaks.Streak) error {
    query := `
        INSERT INTO streaks (id, user_id, activity_type, current_streak, longest_streak, last_active_date, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
    `
    _, err := r.db.ExecContext(ctx, query,
        streak.ID,
        streak.UserID,
        streak.ActivityType,
        streak.CurrentStreak,
        streak.LongestStreak,
        streak.LastActiveDate,
        streak.CreatedAt,
        streak.UpdatedAt,
    )
    if err != nil {
        return apperrors.New("INTERNAL_ERROR", "Failed to create streak")
    }
    return nil
}

// GetByID retrieves a streak by ID
func (r *StreakRepository) GetByID(ctx context.Context, id uuid.UUID) (*streaks.Streak, error) {
    query := `
        SELECT id, user_id, activity_type, current_streak, longest_streak, last_active_date, created_at, updated_at
        FROM streaks
        WHERE id = $1
    `
    var streak streaks.Streak
    err := r.db.QueryRowContext(ctx, query, id).Scan(
        &streak.ID,
        &streak.UserID,
        &streak.ActivityType,
        &streak.CurrentStreak,
        &streak.LongestStreak,
        &streak.LastActiveDate,
        &streak.CreatedAt,
        &streak.UpdatedAt,
    )
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, apperrors.StreakNotFound
        }
        return nil, apperrors.New("INTERNAL_ERROR", "Failed to get streak")
    }
    return &streak, nil
}

// GetByUserIDAndActivity retrieves a streak by user ID and activity type
func (r *StreakRepository) GetByUserIDAndActivity(ctx context.Context, userID uuid.UUID, activityType string) (*streaks.Streak, error) {
    query := `
        SELECT id, user_id, activity_type, current_streak, longest_streak, last_active_date, created_at, updated_at
        FROM streaks
        WHERE user_id = $1 AND activity_type = $2
    `
    var streak streaks.Streak
    err := r.db.QueryRowContext(ctx, query, userID, activityType).Scan(
        &streak.ID,
        &streak.UserID,
        &streak.ActivityType,
        &streak.CurrentStreak,
        &streak.LongestStreak,
        &streak.LastActiveDate,
        &streak.CreatedAt,
        &streak.UpdatedAt,
    )
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, apperrors.StreakNotFound
        }
        return nil, apperrors.New("INTERNAL_ERROR", "Failed to get streak")
    }
    return &streak, nil
}

// Update updates an existing streak
func (r *StreakRepository) Update(ctx context.Context, streak *streaks.Streak) error {
    query := `
        UPDATE streaks
        SET current_streak = $2, longest_streak = $3, last_active_date = $4, updated_at = $5
        WHERE id = $1
    `
    _, err := r.db.ExecContext(ctx, query,
        streak.ID,
        streak.CurrentStreak,
        streak.LongestStreak,
        streak.LastActiveDate,
        streak.UpdatedAt,
    )
    if err != nil {
        return apperrors.New("INTERNAL_ERROR", "Failed to update streak")
    }
    return nil
}

// IncrementStreak increments the streak for a user
func (r *StreakRepository) IncrementStreak(ctx context.Context, userID uuid.UUID, activityType string) (*streaks.Streak, error) {
    // Get current streak
    streak, err := r.GetByUserIDAndActivity(ctx, userID, activityType)
    if err != nil {
        if errors.Is(err, apperrors.StreakNotFound) {
            // Create new streak
            now := time.Now()
            newStreak := &streaks.Streak{
                ID:             uuid.New(),
                UserID:         userID,
                ActivityType:   activityType,
                CurrentStreak:  1,
                LongestStreak:  1,
                LastActiveDate: &now,
                CreatedAt:      now,
                UpdatedAt:      now,
            }
            if err := r.Create(ctx, newStreak); err != nil {
                return nil, err
            }
            return newStreak, nil
        }
        return nil, err
    }

    // Check if we need to reset or increment
    now := time.Now()
    if streak.LastActiveDate == nil {
        streak.CurrentStreak = 1
    } else {
        lastActive := *streak.LastActiveDate
        daysDiff := int(now.Truncate(24 * time.Hour).Sub(lastActive.Truncate(24 * time.Hour)).Hours() / 24)

        if daysDiff == 0 {
            // Already logged today, no change
        } else if daysDiff == 1 {
            // Consecutive day, increment
            streak.CurrentStreak++
        } else {
            // Broken streak, reset
            streak.CurrentStreak = 1
        }
    }

    // Update longest streak if needed
    if streak.CurrentStreak > streak.LongestStreak {
        streak.LongestStreak = streak.CurrentStreak
    }

    streak.LastActiveDate = &now
    streak.UpdatedAt = now

    if err := r.Update(ctx, streak); err != nil {
        return nil, err
    }

    return streak, nil
}

// GetByUserID retrieves all streaks for a user
func (r *StreakRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]streaks.Streak, error) {
    query := `
        SELECT id, user_id, activity_type, current_streak, longest_streak, last_active_date, created_at, updated_at
        FROM streaks
        WHERE user_id = $1
        ORDER BY activity_type
    `
    rows, err := r.db.QueryContext(ctx, query, userID)
    if err != nil {
        return nil, apperrors.New("INTERNAL_ERROR", "Failed to query streaks")
    }
    defer rows.Close()

    var streakList []streaks.Streak
    for rows.Next() {
        var streak streaks.Streak
        if err := rows.Scan(
            &streak.ID,
            &streak.UserID,
            &streak.ActivityType,
            &streak.CurrentStreak,
            &streak.LongestStreak,
            &streak.LastActiveDate,
            &streak.CreatedAt,
            &streak.UpdatedAt,
        ); err != nil {
            return nil, apperrors.New("INTERNAL_ERROR", "Failed to scan streak row")
        }
        streakList = append(streakList, streak)
    }
    if err := rows.Err(); err != nil {
        return nil, apperrors.New("INTERNAL_ERROR", "Failed to iterate streak rows")
    }
    return streakList, nil
}
