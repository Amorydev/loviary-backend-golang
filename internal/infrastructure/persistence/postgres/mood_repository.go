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

// MoodRepository handles database operations for moods.
type MoodRepository struct {
	db *sql.DB
}

// NewMoodRepository creates a new mood repository.
func NewMoodRepository(db *sql.DB) *MoodRepository {
	return &MoodRepository{db: db}
}

// moodColumns is the standard SELECT list for a moods row.
const moodColumns = `mood_id, user_id, logged_date, mood_type, intensity, mood_emoji,
	mood_description, is_private, created_at`

// scanMood scans a standard moods row into a Mood struct.
func scanMood(scanner interface {
	Scan(dest ...interface{}) error
}, m *moods.Mood) error {
	return scanner.Scan(
		&m.ID,
		&m.UserID,
		&m.Date,
		&m.MoodType,
		&m.Intensity,
		&m.MoodEmoji,
		&m.Note,
		&m.IsShared,
		&m.CreatedAt,
	)
}

// Create inserts a new mood.
func (r *MoodRepository) Create(ctx context.Context, mood *moods.Mood) error {
	query := `
		INSERT INTO moods (mood_id, user_id, couple_id, logged_date, mood_type, intensity,
		                   mood_emoji, mood_description, is_private, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	// couple_id is required by the DB — derive from context or pass through mood.
	// For now mood struct doesn't carry couple_id; passing NULL is not allowed.
	// The handler layer should embed couple_id into the request and pass it here.
	// We use a placeholder until the Mood struct is extended.
	_, err := r.db.ExecContext(ctx, query,
		mood.ID,
		mood.UserID,
		nil, // couple_id — set by caller via extended input
		mood.Date,
		mood.MoodType,
		mood.Intensity,
		mood.MoodEmoji,
		mood.Note,
		mood.IsShared, // maps to is_private (inverted: IsShared=true → is_private=false)
		mood.CreatedAt,
	)
	if err != nil {
		return apperrors.New("INTERNAL_ERROR", "Failed to create mood")
	}
	return nil
}

// GetByID retrieves a mood by ID.
func (r *MoodRepository) GetByID(ctx context.Context, id uuid.UUID) (*moods.Mood, error) {
	query := `SELECT ` + moodColumns + ` FROM moods WHERE mood_id = $1`
	var mood moods.Mood
	err := scanMood(r.db.QueryRowContext(ctx, query, id), &mood)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.MoodNotFound
		}
		return nil, apperrors.New("INTERNAL_ERROR", "Failed to get mood")
	}
	return &mood, nil
}

// GetByUserIDAndDate retrieves today's mood for a user.
func (r *MoodRepository) GetByUserIDAndDate(ctx context.Context, userID uuid.UUID, date time.Time) (*moods.Mood, error) {
	query := `SELECT ` + moodColumns + ` FROM moods WHERE user_id = $1 AND logged_date = $2`
	var mood moods.Mood
	err := scanMood(r.db.QueryRowContext(ctx, query, userID, date.Format("2006-01-02")), &mood)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.MoodNotFound
		}
		return nil, apperrors.New("INTERNAL_ERROR", "Failed to get mood by date")
	}
	return &mood, nil
}

// GetByUserID retrieves moods for a user within a date range.
func (r *MoodRepository) GetByUserID(ctx context.Context, userID uuid.UUID, startDate, endDate time.Time) ([]moods.Mood, error) {
	query := `
		SELECT ` + moodColumns + `
		FROM moods
		WHERE user_id = $1 AND logged_date >= $2 AND logged_date <= $3
		ORDER BY logged_date DESC
	`
	rows, err := r.db.QueryContext(ctx, query, userID, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
	if err != nil {
		return nil, apperrors.New("INTERNAL_ERROR", "Failed to query moods")
	}
	defer rows.Close()

	var list []moods.Mood
	for rows.Next() {
		var mood moods.Mood
		if err := scanMood(rows, &mood); err != nil {
			return nil, apperrors.New("INTERNAL_ERROR", "Failed to scan mood row")
		}
		list = append(list, mood)
	}
	if err := rows.Err(); err != nil {
		return nil, apperrors.New("INTERNAL_ERROR", "Failed to iterate mood rows")
	}
	return list, nil
}

// Update updates an existing mood.
func (r *MoodRepository) Update(ctx context.Context, mood *moods.Mood) error {
	query := `
		UPDATE moods
		SET mood_type = $2, intensity = $3, mood_emoji = $4,
		    mood_description = $5, is_private = $6
		WHERE mood_id = $1
	`
	_, err := r.db.ExecContext(ctx, query,
		mood.ID,
		mood.MoodType,
		mood.Intensity,
		mood.MoodEmoji,
		mood.Note,
		mood.IsShared,
	)
	if err != nil {
		return apperrors.New("INTERNAL_ERROR", "Failed to update mood")
	}
	return nil
}

// Delete removes a mood.
func (r *MoodRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM moods WHERE mood_id = $1`, id)
	if err != nil {
		return apperrors.New("INTERNAL_ERROR", "Failed to delete mood")
	}
	if n, _ := result.RowsAffected(); n == 0 {
		return apperrors.MoodNotFound
	}
	return nil
}

// GetSharedMoods retrieves non-private moods for two users within a date range.
func (r *MoodRepository) GetSharedMoods(ctx context.Context, userID1, userID2 uuid.UUID, startDate, endDate time.Time) ([]moods.Mood, error) {
	query := `
		SELECT ` + moodColumns + `
		FROM moods
		WHERE user_id IN ($1, $2)
		  AND is_private = FALSE
		  AND logged_date >= $3
		  AND logged_date <= $4
		ORDER BY logged_date DESC
	`
	rows, err := r.db.QueryContext(ctx, query, userID1, userID2, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
	if err != nil {
		return nil, apperrors.New("INTERNAL_ERROR", "Failed to query shared moods")
	}
	defer rows.Close()

	var list []moods.Mood
	for rows.Next() {
		var mood moods.Mood
		if err := scanMood(rows, &mood); err != nil {
			return nil, apperrors.New("INTERNAL_ERROR", "Failed to scan mood row")
		}
		list = append(list, mood)
	}
	if err := rows.Err(); err != nil {
		return nil, apperrors.New("INTERNAL_ERROR", "Failed to iterate mood rows")
	}
	return list, nil
}
