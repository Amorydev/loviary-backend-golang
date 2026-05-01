package streaks

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

// DayLog represents a single day in the weekly_log.
type DayLog struct {
	Day       string `json:"day"`       // "Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"
	Completed bool   `json:"completed"` // true = both users logged that day
}

// Streak represents the couple-level streak for a given activity type.
// Both users must log on the same day for current_streak to increment.
type Streak struct {
	StreakID           uuid.UUID  `json:"streak_id"           db:"streak_id"`
	CoupleID           uuid.UUID  `json:"couple_id"           db:"couple_id"`
	ActivityType       string     `json:"activity_type"       db:"activity_type"`
	CurrentStreak      int        `json:"current_streak"      db:"current_streak"`
	LongestStreak      int        `json:"longest_streak"      db:"longest_streak"`
	LastCompletedDate  *time.Time `json:"last_completed_date" db:"last_completed_date"`
	CreatedAt          time.Time  `json:"created_at"          db:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"          db:"updated_at"`

	// WeeklyLog is computed from streak_daily_logs, not stored in streaks table.
	WeeklyLog []DayLog `json:"weekly_log" db:"-"`
}

// GetStreakStatus returns the current status of the streak.
func (s *Streak) GetStreakStatus() StreakStatus {
	if s.LastCompletedDate == nil {
		return StreakStatusNone
	}
	now := time.Now()
	last := *s.LastCompletedDate
	// Compare calendar dates only
	nowDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	lastDate := time.Date(last.Year(), last.Month(), last.Day(), 0, 0, 0, 0, time.UTC)
	daysSince := int(nowDate.Sub(lastDate).Hours() / 24)

	switch {
	case daysSince == 0:
		return StreakStatusActive
	case daysSince == 1:
		return StreakStatusAtRisk
	case daysSince <= 3:
		return StreakStatusBroken
	default:
		return StreakStatusReset
	}
}

// StreakStatus represents the status of a streak.
type StreakStatus string

const (
	StreakStatusNone   StreakStatus = "none"
	StreakStatusActive StreakStatus = "active"
	StreakStatusAtRisk StreakStatus = "at_risk"
	StreakStatusBroken StreakStatus = "broken"
	StreakStatusReset  StreakStatus = "reset"
)

// Scan implements sql.Scanner for StreakStatus.
func (ss *StreakStatus) Scan(value interface{}) error {
	if value == nil {
		*ss = StreakStatusNone
		return nil
	}
	str, ok := value.(string)
	if !ok {
		return errors.New("invalid type for streak_status")
	}
	*ss = StreakStatus(str)
	return nil
}

// Value implements driver.Valuer for StreakStatus.
func (ss StreakStatus) Value() (driver.Value, error) {
	return string(ss), nil
}

// MarshalJSON implements json.Marshaler.
func (ss StreakStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(ss))
}

// UnmarshalJSON implements json.Unmarshaler.
func (ss *StreakStatus) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*ss = StreakStatus(s)
	return nil
}
