package chapters

import (
	"time"

	"github.com/google/uuid"
)

// CoverPlaceholderURL is the fallback cover image when cover_art_url is NULL.
const CoverPlaceholderURL = "https://thumbs.dreamstime.com/b/anime-landscape-lush-greenery-rocky-cliffs-distant-water-under-bright-sky-stunning-style-showcases-green-fields-vibrant-354691808.jpg"

// ChapterDefinition represents a global chapter template (from chapter_definitions table).
// Shared across all couples — not per-couple.
type ChapterDefinition struct {
	ChapterNumber   int     `db:"chapter_number" json:"chapter_number"`
	Title           string  `db:"title"           json:"title"`
	Subtitle        string  `db:"subtitle"        json:"subtitle"`
	DayStart        int     `db:"day_start"       json:"day_start"`
	DayEnd          *int    `db:"day_end"         json:"day_end,omitempty"`
	MilestoneTarget *int    `db:"milestone_target" json:"milestone_target,omitempty"`
	CoverArtURL     string  `db:"cover_art_url"   json:"cover_art_url"`
	BadgeIconURL    *string `db:"badge_icon_url"  json:"badge_icon_url,omitempty"`
	ThemeColor      *string `db:"theme_color"     json:"theme_color,omitempty"`
}

// GetCoverURL returns cover_art_url, falling back to the placeholder if empty.
func (d *ChapterDefinition) GetCoverURL() string {
	if d.CoverArtURL != "" {
		return d.CoverArtURL
	}
	return CoverPlaceholderURL
}

// CalculateMilestonePercent returns the progress percentage (0–100) within this chapter.
func (d *ChapterDefinition) CalculateMilestonePercent(daysTogether int) int {
	if d.DayEnd == nil {
		return 100 // last chapter — unlimited
	}
	span := *d.DayEnd - d.DayStart
	if span <= 0 {
		return 100
	}
	pct := (daysTogether - d.DayStart) * 100 / span
	if pct > 100 {
		return 100
	}
	if pct < 0 {
		return 0
	}
	return pct
}

// GetMilestoneTarget returns the milestone target day, defaulting to DayEnd.
func (d *ChapterDefinition) GetMilestoneTarget() int {
	if d.MilestoneTarget != nil {
		return *d.MilestoneTarget
	}
	if d.DayEnd != nil {
		return *d.DayEnd
	}
	return 0
}

// CoupleChapter represents the per-couple unlock state (from love_chapters table).
type CoupleChapter struct {
	ChapterID     uuid.UUID  `db:"chapter_id"     json:"chapter_id"`
	CoupleID      uuid.UUID  `db:"couple_id"      json:"couple_id"`
	ChapterNumber int        `db:"chapter_number" json:"chapter_number"`
	IsUnlocked    bool       `db:"is_unlocked"    json:"is_unlocked"`
	UnlockedAt    *time.Time `db:"unlocked_at"    json:"unlocked_at,omitempty"`
	UnlockedBy    *uuid.UUID `db:"unlocked_by"    json:"unlocked_by,omitempty"`
}

// CurrentChapterView is the aggregate view used for dashboard responses.
type CurrentChapterView struct {
	Definition       ChapterDefinition
	DaysTogether     int
	MilestonePercent int
	IsUnlocked       bool
}
