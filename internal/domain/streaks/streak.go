package streaks

import (
    "database/sql/driver"
    "encoding/json"
    "errors"
    "time"

    "github.com/google/uuid"
)

// Streak represents a consecutive activity streak for a user
type Streak struct {
    ID            uuid.UUID `json:"id" db:"id"`
    UserID        uuid.UUID `json:"user_id" db:"user_id"`
    ActivityType  string    `json:"activity_type" db:"activity_type"` // e.g., "mood_log", "date_night"
    CurrentStreak int       `json:"current_streak" db:"current_streak"`
    LongestStreak int       `json:"longest_streak" db:"longest_streak"`
    LastActiveDate *time.Time `json:"last_active_date,omitempty" db:"last_active_date"`
    CreatedAt     time.Time `json:"created_at" db:"created_at"`
    UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

// GetStreakStatus returns the current streak status
func (s *Streak) GetStreakStatus() StreakStatus {
    if s.LastActiveDate == nil {
        return StreakStatusNone
    }
    now := time.Now()
    lastActive := *s.LastActiveDate
    daysSinceLastActive := int(now.Sub(lastActive).Hours() / 24)

    if daysSinceLastActive == 0 {
        return StreakStatusActive
    } else if daysSinceLastActive == 1 {
        return StreakStatusAtRisk
    } else if daysSinceLastActive <= 3 {
        return StreakStatusBroken
    }
    return StreakStatusReset
}

// StreakStatus represents the status of a streak
type StreakStatus string

const (
    StreakStatusNone   StreakStatus = "none"
    StreakStatusActive StreakStatus = "active"
    StreakStatusAtRisk StreakStatus = "at_risk"
    StreakStatusBroken StreakStatus = "broken"
    StreakStatusReset  StreakStatus = "reset"
)

// Scan implements sql.Scanner for StreakStatus
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

// Value implements driver.Valuer for StreakStatus
func (ss StreakStatus) Value() (driver.Value, error) {
    return string(ss), nil
}

// JSONMarshalStreakStatus custom marshal for StreakStatus
type JSONMarshalStreakStatus StreakStatus

// MarshalJSON implements json.Marshaler
func (ss StreakStatus) MarshalJSON() ([]byte, error) {
    return json.Marshal(string(ss))
}

// UnmarshalJSON implements json.Unmarshaler
func (ss *StreakStatus) UnmarshalJSON(data []byte) error {
    var s string
    if err := json.Unmarshal(data, &s); err != nil {
        return err
    }
    *ss = StreakStatus(s)
    return nil
}
