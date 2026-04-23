package persistence

import (
    "context"
    "database/sql"
    "errors"
    "time"

    "github.com/google/uuid"

    "loviary.app/backend/internal/domain/moods"
    apperrors "loviary.app/backend/pkg/errors"
)

// MoodRepository handles database operations for moods
type MoodRepository struct {
    db *sql.DB
}

// NewMoodRepository creates a new mood repository
func NewMoodRepository(db *sql.DB) *MoodRepository {
    return &MoodRepository{db: db}
}

// Create inserts a new mood
func (r *MoodRepository) Create(ctx context.Context, mood *moods.Mood) error {
    query := `
        INSERT INTO moods (id, user_id, date, mood_type, intensity, note, is_shared, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
    `
    _, err := r.db.ExecContext(ctx, query,
        mood.ID,
        mood.UserID,
        mood.Date,
        mood.MoodType,
        mood.Intensity,
        mood.Note,
        mood.IsShared,
        mood.CreatedAt,
        mood.UpdatedAt,
    )
    if err != nil {
        return apperrors.New("INTERNAL_ERROR", "Failed to create mood")
    }
    return nil
}

// GetByID retrieves a mood by ID
func (r *MoodRepository) GetByID(ctx context.Context, id uuid.UUID) (*moods.Mood, error) {
    query := `
        SELECT id, user_id, date, mood_type, intensity, note, is_shared, created_at, updated_at
        FROM moods
        WHERE id = $1
    `
    var mood moods.Mood
    err := r.db.QueryRowContext(ctx, query, id).Scan(
        &mood.ID,
        &mood.UserID,
        &mood.Date,
        &mood.MoodType,
        &mood.Intensity,
        &mood.Note,
        &mood.IsShared,
        &mood.CreatedAt,
        &mood.UpdatedAt,
    )
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, apperrors.MoodNotFound
        }
        return nil, apperrors.New("INTERNAL_ERROR", "Failed to get mood")
    }
    return &mood, nil
}

// GetByUserIDAndDate retrieves a mood by user ID and date
func (r *MoodRepository) GetByUserIDAndDate(ctx context.Context, userID uuid.UUID, date time.Time) (*moods.Mood, error) {
    query := `
        SELECT id, user_id, date, mood_type, intensity, note, is_shared, created_at, updated_at
        FROM moods
        WHERE user_id = $1 AND DATE(date) = DATE($2)
    `
    var mood moods.Mood
    err := r.db.QueryRowContext(ctx, query, userID, date).Scan(
        &mood.ID,
        &mood.UserID,
        &mood.Date,
        &mood.MoodType,
        &mood.Intensity,
        &mood.Note,
        &mood.IsShared,
        &mood.CreatedAt,
        &mood.UpdatedAt,
    )
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, apperrors.MoodNotFound
        }
        return nil, apperrors.New("INTERNAL_ERROR", "Failed to get mood by date")
    }
    return &mood, nil
}

// GetByUserID retrieves all moods for a user within a date range
func (r *MoodRepository) GetByUserID(ctx context.Context, userID uuid.UUID, startDate, endDate time.Time) ([]moods.Mood, error) {
    query := `
        SELECT id, user_id, date, mood_type, intensity, note, is_shared, created_at, updated_at
        FROM moods
        WHERE user_id = $1 AND date >= $2 AND date <= $3
        ORDER BY date DESC
    `
    rows, err := r.db.QueryContext(ctx, query, userID, startDate, endDate)
    if err != nil {
        return nil, apperrors.New("INTERNAL_ERROR", "Failed to query moods")
    }
    defer rows.Close()

    var moodList []moods.Mood
    for rows.Next() {
        var mood moods.Mood
        if err := rows.Scan(
            &mood.ID,
            &mood.UserID,
            &mood.Date,
            &mood.MoodType,
            &mood.Intensity,
            &mood.Note,
            &mood.IsShared,
            &mood.CreatedAt,
            &mood.UpdatedAt,
        ); err != nil {
            return nil, apperrors.New("INTERNAL_ERROR", "Failed to scan mood row")
        }
        moodList = append(moodList, mood)
    }
    if err := rows.Err(); err != nil {
        return nil, apperrors.New("INTERNAL_ERROR", "Failed to iterate mood rows")
    }
    return moodList, nil
}

// Update updates an existing mood
func (r *MoodRepository) Update(ctx context.Context, mood *moods.Mood) error {
    query := `
        UPDATE moods
        SET mood_type = $2, intensity = $3, note = $4, is_shared = $5, updated_at = $6
        WHERE id = $1
    `
    _, err := r.db.ExecContext(ctx, query,
        mood.ID,
        mood.MoodType,
        mood.Intensity,
        mood.Note,
        mood.IsShared,
        mood.UpdatedAt,
    )
    if err != nil {
        return apperrors.New("INTERNAL_ERROR", "Failed to update mood")
    }
    return nil
}

// Delete removes a mood
func (r *MoodRepository) Delete(ctx context.Context, id uuid.UUID) error {
    query := `DELETE FROM moods WHERE id = $1`
    result, err := r.db.ExecContext(ctx, query, id)
    if err != nil {
        return apperrors.New("INTERNAL_ERROR", "Failed to delete mood")
    }
    if rowsAffected, _ := result.RowsAffected(); rowsAffected == 0 {
        return apperrors.MoodNotFound
    }
    return nil
}

// GetSharedMoods retrieves shared moods for a couple
func (r *MoodRepository) GetSharedMoods(ctx context.Context, userID1, userID2 uuid.UUID, startDate, endDate time.Time) ([]moods.Mood, error) {
    query := `
        SELECT id, user_id, date, mood_type, intensity, note, is_shared, created_at, updated_at
        FROM moods
        WHERE user_id IN ($1, $2) AND is_shared = true AND date >= $3 AND date <= $4
        ORDER BY date DESC
    `
    rows, err := r.db.QueryContext(ctx, query, userID1, userID2, startDate, endDate)
    if err != nil {
        return nil, apperrors.New("INTERNAL_ERROR", "Failed to query shared moods")
    }
    defer rows.Close()

    var moodList []moods.Mood
    for rows.Next() {
        var mood moods.Mood
        if err := rows.Scan(
            &mood.ID,
            &mood.UserID,
            &mood.Date,
            &mood.MoodType,
            &mood.Intensity,
            &mood.Note,
            &mood.IsShared,
            &mood.CreatedAt,
            &mood.UpdatedAt,
        ); err != nil {
            return nil, apperrors.New("INTERNAL_ERROR", "Failed to scan mood row")
        }
        moodList = append(moodList, mood)
    }
    if err := rows.Err(); err != nil {
        return nil, apperrors.New("INTERNAL_ERROR", "Failed to iterate mood rows")
    }
    return moodList, nil
}
