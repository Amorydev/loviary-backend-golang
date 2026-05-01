package persistence

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"

	domainChapters "loviary.app/backend/internal/domain/chapters"
	apperrors "loviary.app/backend/pkg/errors"
)

// ChapterRepository handles database operations for chapters.
type ChapterRepository struct {
	db *sql.DB
}

// NewChapterRepository creates a new chapter repository.
func NewChapterRepository(db *sql.DB) *ChapterRepository {
	return &ChapterRepository{db: db}
}

// GetDefinitionByDay returns the chapter definition that covers the given number of days together.
func (r *ChapterRepository) GetDefinitionByDay(ctx context.Context, daysTogether int) (*domainChapters.ChapterDefinition, error) {
	query := `
		SELECT chapter_number, title, COALESCE(subtitle, ''), day_start, day_end,
		       milestone_target, COALESCE(cover_art_url, ''), badge_icon_url, theme_color
		FROM chapter_definitions
		WHERE day_start <= $1
		ORDER BY day_start DESC
		LIMIT 1
	`
	var def domainChapters.ChapterDefinition
	err := r.db.QueryRowContext(ctx, query, daysTogether).Scan(
		&def.ChapterNumber,
		&def.Title,
		&def.Subtitle,
		&def.DayStart,
		&def.DayEnd,
		&def.MilestoneTarget,
		&def.CoverArtURL,
		&def.BadgeIconURL,
		&def.ThemeColor,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, apperrors.New("INTERNAL_ERROR", "Failed to query chapter definition")
	}
	return &def, nil
}

// GetAllDefinitions returns all chapter definitions ordered by chapter_number.
func (r *ChapterRepository) GetAllDefinitions(ctx context.Context) ([]domainChapters.ChapterDefinition, error) {
	query := `
		SELECT chapter_number, title, COALESCE(subtitle, ''), day_start, day_end,
		       milestone_target, COALESCE(cover_art_url, ''), badge_icon_url, theme_color
		FROM chapter_definitions
		ORDER BY chapter_number
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, apperrors.New("INTERNAL_ERROR", "Failed to query chapter definitions")
	}
	defer rows.Close()

	var defs []domainChapters.ChapterDefinition
	for rows.Next() {
		var def domainChapters.ChapterDefinition
		if err := rows.Scan(
			&def.ChapterNumber,
			&def.Title,
			&def.Subtitle,
			&def.DayStart,
			&def.DayEnd,
			&def.MilestoneTarget,
			&def.CoverArtURL,
			&def.BadgeIconURL,
			&def.ThemeColor,
		); err != nil {
			return nil, apperrors.New("INTERNAL_ERROR", "Failed to scan chapter definition")
		}
		defs = append(defs, def)
	}
	if err := rows.Err(); err != nil {
		return nil, apperrors.New("INTERNAL_ERROR", "Failed to iterate chapter definitions")
	}
	return defs, nil
}

// GetCoupleChapter returns the unlock state of a specific chapter for a couple.
func (r *ChapterRepository) GetCoupleChapter(ctx context.Context, coupleID uuid.UUID, chapterNumber int) (*domainChapters.CoupleChapter, error) {
	query := `
		SELECT chapter_id, couple_id, chapter_number, is_unlocked, unlocked_at, unlocked_by
		FROM love_chapters
		WHERE couple_id = $1 AND chapter_number = $2
	`
	var cc domainChapters.CoupleChapter
	err := r.db.QueryRowContext(ctx, query, coupleID, chapterNumber).Scan(
		&cc.ChapterID,
		&cc.CoupleID,
		&cc.ChapterNumber,
		&cc.IsUnlocked,
		&cc.UnlockedAt,
		&cc.UnlockedBy,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, apperrors.New("INTERNAL_ERROR", "Failed to get couple chapter")
	}
	return &cc, nil
}

// UnlockChapter marks a chapter as unlocked for a couple, upserting if necessary.
func (r *ChapterRepository) UnlockChapter(ctx context.Context, coupleID uuid.UUID, chapterNumber int, unlockedBy uuid.UUID) error {
	query := `
		INSERT INTO love_chapters (chapter_id, couple_id, chapter_number, is_unlocked, unlocked_at, unlocked_by)
		VALUES ($1, $2, $3, TRUE, $4, $5)
		ON CONFLICT (couple_id, chapter_number)
		DO UPDATE SET is_unlocked = TRUE, unlocked_at = $4, unlocked_by = $5
	`
	_, err := r.db.ExecContext(ctx, query,
		uuid.New(),
		coupleID,
		chapterNumber,
		time.Now(),
		unlockedBy,
	)
	if err != nil {
		return apperrors.New("INTERNAL_ERROR", "Failed to unlock chapter")
	}
	return nil
}
